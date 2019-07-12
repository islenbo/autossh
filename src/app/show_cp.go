package app

import (
	"autossh/src/utils"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type TransferObjectType int

const (
	TransferObjectTypeLocal TransferObjectType = iota
	TransferObjectTypeRemote
)

type TransferObject struct {
	raw    string             // 原始数据，如 vagrant:/root/example.txt
	cpType TransferObjectType // 类型，TransferObjectTypeLocal-本地，TransferObjectTypeRemote-远程
	server Server             // 服务器，cpType = TransferObjectTypeRemote 时为空
	path   string             // 从raw解析得到的文件路径，如 /root/example.txt
}

type CpType int

const (
	CpTypeUpload CpType = iota
	CpTypeDownload
)

type Cp struct {
	isDir bool
	cfg   *Config

	cpType  CpType
	sources []*TransferObject
	target  *TransferObject
}

// 复制
func showCp(configFile string) {
	var err error
	cfg, err := loadConfig(configFile)
	if err != nil {
		utils.Errorln(err)
		return
	}

	cp := Cp{cfg: cfg}
	if err := cp.parse(); err != nil {
		utils.Errorln(err)
		return
	}

	if cp.sources[0].cpType == TransferObjectTypeLocal {
		cp.cpType = CpTypeUpload
		err = cp.upload()
	} else {
		cp.cpType = CpTypeDownload
		err = cp.download()
	}

	if err != nil {
		utils.Errorln(err)
		return
	}
}

// 解析参数
func (cp *Cp) parse() error {
	args := flag.Args()[1:]
	if len(args) == 0 {
		return errors.New("请输入完整参数")
	}

	if args[0] == "-r" {
		cp.isDir = true
		args = args[1:]
	}

	var err error
	switch len(args) {
	case 0:
		return errors.New("请输入完整参数")
	case 1:
		// 默认取temp目录作为target
		args = []string{args[0], os.TempDir()}
	}

	length := len(args)
	cp.target, err = newTransferObject(*cp.cfg, args[length-1])
	if err != nil {
		return err
	}

	cp.sources = make([]*TransferObject, 0)
	for i, arg := range args[:length-1] {
		s, err := newTransferObject(*cp.cfg, arg)
		if err != nil {
			return err
		}

		if s.cpType == TransferObjectTypeLocal && s.cpType == cp.target.cpType {
			return errors.New("源和目标不能同时为本地地址")
		}

		if i > 0 && s.cpType != cp.sources[i-1].cpType {
			return errors.New("source 类型不一致")
		}

		cp.sources = append(cp.sources, s)
	}

	return nil
}

// 上传
func (cp *Cp) upload() error {
	sftpClient, err := cp.target.server.GetSftpClient()
	if err != nil {
		return err
	}

	defer func() {
		_ = sftpClient.Close()
	}()

	var ioClient = IOClient{ClientType: IOClientSftp, SftpClient: sftpClient}

	for _, source := range cp.sources {
		if file, err := cp.transfer(&ioClient, source.path, cp.target.path, ""); err != nil {
			cp.printFileError(file, err)
		}
	}

	return nil
}

// IO复制 src -> dst
func (cp *Cp) ioCopy(client *IOClient, srcFile FileLike, dst string, fSize int64) (string, error) {
	var err error

	dst, err = cp.parseDstFilename(client, srcFile.Name(), dst)
	if err != nil {
		return dst, err
	}

	dstFile, err := client.Create(dst)
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
		process = float64(bytesCount) / float64(fSize) * 100
		speed = float64(bytesCount) / time.Now().Sub(startTime).Seconds()

		if eof {
			cp.printProcess(filename, 100.0, startTime, speed)
			break
		}
	}

	fmt.Println("")
	return "", nil
}

// 下载
func (cp *Cp) download() error {
	for _, source := range cp.sources {
		sftpClient, err := source.server.GetSftpClient()
		if err != nil {
			return err
		}

		func() {
			defer func() {
				_ = sftpClient.Close()
			}()

			var ioClient = IOClient{ClientType: IOClientLocal, SftpClient: sftpClient}
			if file, err := cp.transfer(&ioClient, source.path, cp.target.path, ""); err != nil {
				cp.printFileError(file, err)
			}
		}()
	}

	return nil
}

// 传输
// 上传时，src = 本地，dst = 远程
// 下载时，src = 远程，dst = 本地
func (cp *Cp) transfer(client *IOClient, src string, dst string, vPath string) (string, error) {
	srcFile, err := cp.openFile(client.SftpClient, src)
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

		childFiles, err := cp.readDir(client.SftpClient, srcFile.Name())
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
			if str, err := cp.transfer(client, childFilename, dst, vPath); err != nil {
				cp.printFileError(str, err)
			}
		}
	} else {
		newDst := path.Join(dst, vPath)

		if file, err := cp.ioCopy(client, srcFile, newDst, srcFileInfo.Size()); err != nil {
			return file, err
		}
	}

	return "", nil
}

// 解析dst文件名
// src = /root/example.txt dst = /root/ => /root/example.txt
// src = /root/example.txt dst = /root => /root/example.txt
// src = /root/example.txt dst = /root/new-name.txt => /root/new-name.txt
func (cp *Cp) parseDstFilename(client *IOClient, src string, dst string) (string, error) {
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

// 根据上传/下载打开相应位置的文件
func (cp *Cp) openFile(client *sftp.Client, file string) (FileLike, error) {
	if cp.cpType == CpTypeUpload {
		return os.Open(file)
	} else {
		return client.Open(file)
	}
}

// 根据上传/下载读取相应位置的目录，返回文件列表
func (cp *Cp) readDir(client *sftp.Client, name string) ([]os.FileInfo, error) {
	if cp.cpType == CpTypeUpload {
		return ioutil.ReadDir(name)
	} else {
		return client.ReadDir(name)
	}
}

// 创建传输对象
func newTransferObject(cfg Config, raw string) (*TransferObject, error) {
	obj := TransferObject{
		raw: raw,
	}

	args := strings.Split(raw, ":")
	switch len(args) {
	case 1:
		obj.cpType = TransferObjectTypeLocal
		obj.path = args[0]
	case 2:
		obj.path = strings.TrimSpace(args[1])
		serverIndex, exists := cfg.serverIndex[args[0]]
		if !exists {
			return nil, errors.New("服务器" + args[0] + "不存在")
		}
		obj.cpType = TransferObjectTypeRemote
		obj.server = *serverIndex.server

	default:
		return nil, errors.New(raw + " 格式错误")
	}

	return &obj, nil
}
