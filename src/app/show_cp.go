package app

import (
	"autossh/src/utils"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type ResType int

const (
	ResTypeSrc ResType = iota
	ResTypeDst
)

type TransferObject struct {
	raw     string  // 原始数据，如 vagrant:/root/example.txt
	resType ResType // 类型，ResTypeSrc-源，ResTypeDst-目的
	server  *Server // 服务器，当raw为本地址地时，该字段为空
	path    string  // 从raw解析得到的文件路径，如 /root/example.txt
}

type Cp struct {
	isDir bool
	cfg   *Config

	sources []*TransferObject
	target  *TransferObject
}

// 复制
func showCp() {
	cp := Cp{cfg: cfg}
	if err := cp.parse(); err != nil {
		utils.Errorln(err)
		return
	}

	var dstIoClient IOClient
	if cp.target.server == nil {
		dstIoClient = new(LocalIOClient)
	} else {
		sftpClient, err := cp.target.server.GetSftpClient()
		if err != nil {
			utils.Errorln(err)
			return
		}

		defer func() {
			_ = sftpClient.Close()
		}()

		c := SftpIOClient{SftpClient: sftpClient}
		dstIoClient = &c
	}

	for _, source := range cp.sources {
		var srcIoClient IOClient
		var sftpClient *sftp.Client

		if source.server == nil {
			srcIoClient = new(LocalIOClient)
		} else {
			sftpClient, err := source.server.GetSftpClient()
			if err != nil {
				cp.printFileError(source.path, err)
				continue
			}

			srcIoClient = &SftpIOClient{SftpClient: sftpClient}
		}

		func() {
			defer func() {
				if sftpClient != nil {
					_ = sftpClient.Close()
				}
			}()

			if file, err := cp.transferNew(srcIoClient, dstIoClient, source.path, cp.target.path, ""); err != nil {
				cp.printFileError(file, err)
			}
		}()
	}
}

// 解析参数
func (cp *Cp) parse() error {
	os.Args = flag.Args()
	flag.BoolVar(&cp.isDir, "r", false, "文件夹")
	flag.Parse()

	var args = flag.Args()
	var length = len(args)
	var err error

	if len(args) < 1 {
		return errors.New("请输入完整参数")
	}

	cp.target, err = newTransferObject(*cp.cfg, args[length-1])
	if err != nil {
		return err
	}

	cp.sources = make([]*TransferObject, 0)
	for _, arg := range args[:length-1] {
		s, err := newTransferObject(*cp.cfg, arg)
		if err != nil {
			return err
		}

		if s.resType == ResTypeSrc && s.resType == cp.target.resType {
			return errors.New("源和目标不能同时为本地地址")
		}

		cp.sources = append(cp.sources, s)
	}

	return nil
}

// IO复制 src -> dst
func (cp *Cp) ioCopy(srcIO IOClient, dstIO IOClient, srcFile FileLike, dst string) (string, error) {
	var err error

	dst, err = cp.parseDstFilename(dstIO, srcFile.Name(), dst)
	if err != nil {
		return dst, err
	}

	dstFile, err := dstIO.Create(dst)
	if err != nil {
		return dst, err
	}

	defer func() {
		_ = dstFile.Close()
	}()

	bytesCount := 0
	filename := path.Base(srcFile.Name())
	startTime := time.Now()
	speed := 0.0
	var process = 0.0

	go func() {
		for {
			cp.printProcess(filename, process, startTime, speed)
			time.Sleep(time.Second)
			if process >= 100 {
				break
			}
		}
	}()

	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return srcFile.Name(), err
	}

	bytes := make([]byte, 64*1024)
	for {
		n, err := srcFile.Read(bytes[:])
		eof := err == io.EOF
		if err != nil && err != io.EOF {
			return srcFile.Name(), err
		}

		wn, err := dstFile.Write(bytes[:n])
		if err != nil {
			return cp.target.path, err
		}
		bytesCount += wn
		process = float64(bytesCount) / float64(srcFileInfo.Size()) * 100
		speed = float64(bytesCount) / time.Now().Sub(startTime).Seconds()

		if eof {
			cp.printProcess(filename, 100.0, startTime, speed)
			break
		}
	}

	fmt.Println("")
	return "", nil
}

// 传输
// 上传时，src = 本地，dst = 远程
// 下载时，src = 远程，dst = 本地
func (cp *Cp) transferNew(srcIO IOClient, dstIO IOClient, src string, dst string, vPath string) (string, error) {
	srcFile, err := srcIO.Open(src)
	if err != nil {
		return src, err
	}

	defer func() {
		_ = srcFile.Close()
	}()

	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return srcFile.Name(), err
	}

	if srcFileInfo.IsDir() {
		if !cp.isDir {
			return src, errors.New("是一个目录")
		}

		childFiles, err := srcIO.ReadDir(srcFile.Name())
		if err != nil {
			return srcFile.Name(), err
		}

		if vPath == "" {
			vPath = string(os.PathSeparator)
		} else {
			vPath = path.Join(vPath, srcFileInfo.Name())
		}

		for _, childFile := range childFiles {
			childFilename := path.Join(src, childFile.Name())
			if str, err := cp.transferNew(srcIO, dstIO, childFilename, dst, vPath); err != nil {
				cp.printFileError(str, err)
			}
		}
	} else {
		newDst := path.Join(dst, vPath)
		if file, err := cp.ioCopy(srcIO, dstIO, srcFile, newDst); err != nil {
			return file, err
		}
	}

	return "", nil
}

// 解析dst文件名
// src = /root/example.txt dst = /root/ => /root/example.txt
// src = /root/example.txt dst = /root => /root/example.txt
// src = /root/example.txt dst = /root/new-name.txt => /root/new-name.txt
func (cp *Cp) parseDstFilename(client IOClient, src string, dst string) (string, error) {
	dstFileInfo, err := client.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return dst, err
		}

		if cp.isDir {
			if err := client.Mkdir(dst); err != nil {
				return dst, err
			}

			dst = path.Join(dst, path.Base(src))
		} else {
			var p = path.Dir(dst)
			if _, err = client.Stat(p); err != nil {
				return dst, err
			}

			dst = path.Join(path.Dir(dst), path.Base(dst))
		}

	} else {
		if dstFileInfo.IsDir() {
			dst = path.Join(dst, path.Base(src))
		}
	}

	return dst, nil
}

func (cp *Cp) printProcess(name string, process float64, startTime time.Time, speed float64) {
	// TODO 文件大小
	execTime := time.Now().Sub(startTime)

	type winSize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	ws := &winSize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	padding := 0
	if int(retCode) != -1 {
		padding = int(ws.Col) - utils.ZhLen(name) - 40
	}

	extInfo := fmt.Sprintf("%.2f%%  %10s/s  %02.0f:%02.0f:%02.0f",
		process,
		utils.SizeFormat(speed),
		execTime.Hours(),
		execTime.Minutes(),
		execTime.Seconds())

	format := "\r%s%-" + strconv.Itoa(padding) + "s%40s"
	fmt.Printf(format, name, "", extInfo)
}

func (cp *Cp) printFileError(name string, err error) {
	fmt.Println(name, ": ", err)
}

// 创建传输对象
func newTransferObject(cfg Config, raw string) (*TransferObject, error) {
	obj := TransferObject{
		raw: raw,
	}

	args := strings.Split(raw, ":")
	switch len(args) {
	case 1:
		obj.resType = ResTypeSrc
		obj.path = args[0]
	case 2:
		obj.path = strings.TrimSpace(args[1])
		serverIndex, exists := cfg.serverIndex[args[0]]
		if !exists {
			return nil, errors.New("服务器" + args[0] + "不存在")
		}
		obj.resType = ResTypeDst
		obj.server = serverIndex.server

	default:
		return nil, errors.New(raw + " 格式错误")
	}

	return &obj, nil
}
