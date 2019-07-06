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
	"strings"
)

type CpType int

const (
	CpTypeLocal CpType = iota
	CpTypeRemote
)

type TransferObject struct {
	raw    string // 原始数据，如 vagrant:/root/example.txt
	cpType CpType
	server Server
	path   string // 从raw解析得到的文件路径，如 /root/example.txt
}

type Cp struct {
	isDir bool
	cfg   *Config

	sources []*TransferObject
	target  *TransferObject
}

type FileLike interface {
	Name() string
	Stat() (os.FileInfo, error)
	Read([]byte) (int, error)
	Close() error
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

	if cp.sources[0].cpType == CpTypeLocal {
		err = cp.upload()
	} else {
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

		if s.cpType == CpTypeLocal && s.cpType == cp.target.cpType {
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

	var funcUpload func(client *sftp.Client, src string, dst string, vPath string) (string, error)
	funcUpload = func(client *sftp.Client, src string, dst string, vPath string) (string, error) {
		srcFile, err := os.Open(src)
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

			childFiles, err := ioutil.ReadDir(srcFile.Name())
			if err != nil {
				return srcFile.Name(), err
			}

			if vPath == "" {
				vPath = "/"
			} else {
				vPath = path.Join(vPath, srcFileInfo.Name())
			}

			for _, childFile := range childFiles {
				childFilename := path.Join(src, childFile.Name())
				if str, err := funcUpload(client, childFilename, dst, vPath); err != nil {
					fmt.Println(str, ": ", err)
				}
			}
		} else {
			newDst := path.Join(dst, vPath)

			if file, err := cp.uploadFile(client, srcFile, newDst, srcFileInfo.Size()); err != nil {
				return file, err
			}
		}

		return "", nil
	}

	for _, source := range cp.sources {
		if file, err := funcUpload(sftpClient, source.path, cp.target.path, ""); err != nil {
			fmt.Println(file, ": ", err)
		}
	}

	return nil
}

// 上传文件
func (cp *Cp) uploadFile(client *sftp.Client, srcFile *os.File, remoteFile string, fSize int64) (string, error) {
	var err error

	remoteFile, err = cp.parseRemoteFilename(client, srcFile.Name(), remoteFile)
	if err != nil {
		return remoteFile, err
	}

	dstFile, err := client.Create(remoteFile)
	if err != nil {
		return remoteFile, err
	}

	defer func() {
		_ = dstFile.Close()
	}()

	return cp.showCopy(srcFile, dstFile, fSize)
}

// 显示复制进度
func (cp *Cp) showCopy(srcFile FileLike, dstFile io.Writer, fSize int64) (string, error) {
	bytes := [4096]byte{}
	bytesCount := 0

	for {
		n, err := srcFile.Read(bytes[:])
		eof := err == io.EOF
		if err != nil && err != io.EOF {
			return srcFile.Name(), err
		}

		bytesCount += n
		process := float64(bytesCount) / float64(fSize) * 100
		fmt.Print("\r" + srcFile.Name() + "\t\t" + fmt.Sprintf("%.2f", process) + "%")
		_, err = dstFile.Write(bytes[:n])
		if err != nil {
			return cp.target.path, err
		}

		if eof {
			break
		}
	}

	fmt.Print("\r"+srcFile.Name()+"\t\t"+"100%    ", "\n")
	return "", nil
}

// 下载
func (cp *Cp) download() error {
	filename := ""
	targetPath := cp.target.path
	targetFile, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			p := path.Dir(targetPath)
			if targetFile, err = os.Stat(p); err != nil {
				return err
			}

			filename = path.Base(targetPath)
			targetPath = p
		} else {
			return err
		}
	} else {
		if !targetFile.IsDir() {
			filename = path.Base(cp.target.path)
		}
	}

	target := path.Join(targetPath, filename)

	for _, source := range cp.sources {
		client, err := source.server.GetSftpClient()
		if err != nil {
			fmt.Println(err)
		}

		if file, err := cp.downloadFile(client, source.path, target); err != nil {
			fmt.Println(file, ": ", err)
		}

		_ = client.Close()
	}

	return nil
}

// 下载单个文件
func (cp *Cp) downloadFile(client *sftp.Client, src string, dst string) (string, error) {
	dstStat, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			if dstFile, err := os.Create(dst); err != nil {
				return dst, nil
			} else {
				dstStat, _ = dstFile.Stat()
			}

		} else {
			return dst, err
		}
	}

	var dstFile *os.File
	if dstStat.IsDir() {
		filename := path.Base(src)
		dstFile, err = os.Create(path.Join(dst, filename))
		if err != nil {
			return dst, err
		}
	} else {
		dstFile, err = os.Create(dst)
		if err != nil {
			return dst, err
		}
	}

	defer func() {
		_ = dstFile.Close()
	}()

	srcFile, err := client.Open(src)
	if err != nil {
		return src, err
	}

	s, err := srcFile.Stat()
	if err != nil {
		return src, err
	}

	if s.IsDir() {
		return src, errors.New("不是一个有效的文件")
	}

	defer func() {
		_ = dstFile.Sync()
	}()

	return cp.showCopy(srcFile, dstFile, s.Size())

}

// 解析远程文件名
// localFile = /root/example.txt remoteFile = /root/ => /root/example.txt
// localFile = /root/example.txt remoteFile = /root => /root/example.txt
// localFile = /root/example.txt remoteFile = /root/new-name.txt => /root/new-name.txt
func (cp *Cp) parseRemoteFilename(client *sftp.Client, localFile string, remoteFile string) (string, error) {
	dstFileInfo, err := client.Stat(remoteFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return remoteFile, err
		}

		if cp.isDir {
			if err := client.Mkdir(remoteFile); err != nil {
				return remoteFile, err
			}

			remoteFile = path.Join(remoteFile, path.Base(localFile))
		} else {
			var p = path.Dir(remoteFile)
			if _, err = client.Stat(p); err != nil {
				return remoteFile, err
			}

			remoteFile = path.Join(path.Dir(remoteFile), path.Base(remoteFile))
		}

	} else {
		if dstFileInfo.IsDir() {
			remoteFile = path.Join(remoteFile, path.Base(localFile))
		}
	}

	return remoteFile, nil
}

// 创建传输对象
func newTransferObject(cfg Config, raw string) (*TransferObject, error) {
	obj := TransferObject{
		raw: raw,
	}

	args := strings.Split(raw, ":")
	switch len(args) {
	case 1:
		obj.cpType = CpTypeLocal
		obj.path = args[0]
	case 2:
		obj.path = strings.TrimSpace(args[1])
		serverIndex, exists := cfg.serverIndex[args[0]]
		if !exists {
			return nil, errors.New("服务器" + args[0] + "不存在")
		}
		obj.cpType = CpTypeRemote
		obj.server = *serverIndex.server

	default:
		return nil, errors.New(raw + " 格式错误")
	}

	return &obj, nil
}
