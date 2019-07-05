package app

import (
	"autossh/src/utils"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
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
	raw    string
	cpType CpType
	server Server
	path   string
}

type Cp struct {
	isDir bool
	cfg   *Config

	source *TransferObject
	target *TransferObject
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

	if cp.source.cpType == CpTypeLocal {
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

	cp.source, err = newTransferObject(*cp.cfg, args[0])
	if err != nil {
		return err
	}

	cp.target, err = newTransferObject(*cp.cfg, args[1])
	if err != nil {
		return err
	}

	if cp.source.cpType == CpTypeLocal && cp.source.cpType == cp.target.cpType {
		return errors.New("源和目标不能同时为本地地址")
	}

	return nil
}

// 上传
func (cp *Cp) upload() error {
	if exists, err := utils.FileIsExists(cp.source.path); !exists {
		return err
	}

	s, _ := os.Stat(cp.source.path)
	if s.IsDir() && !cp.isDir {
		return errors.New("源文件是一个目录")
	}
	srcFile, err := os.Open(cp.source.path)

	sftpClient, err := cp.target.server.GetSftpClient()
	if err != nil {
		return err
	}

	defer sftpClient.Close()

	// create destination file
	filename := path.Base(cp.source.path)
	dstFile, err := sftpClient.Create(path.Join(cp.target.path, filename))
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	// copy source file to destination file
	bytes := [4096]byte{}
	fSize := s.Size()
	bytesCount := 0

	for {
		n, err := srcFile.Read(bytes[:])
		eof := err == io.EOF
		if err != nil && err != io.EOF {
			return err
		}

		bytesCount += n
		process := float64(bytesCount) / float64(fSize) * 100
		fmt.Print("\r" + filename + "\t" + fmt.Sprintf("%.2f", process) + "%")
		_, err = dstFile.Write(bytes[:n])
		if err != nil {
			return err
		}

		if eof {
			break
		}
	}
	fmt.Print("\r" + filename + "\t" + "100%    \n")

	return nil
}

// 下载
func (cp *Cp) download() error {
	sftpClient, err := cp.source.server.GetSftpClient()
	if err != nil {
		return err
	}

	defer sftpClient.Close()

	// create destination file
	filename := path.Base(cp.source.path)
	dstFile, err := os.Create(path.Join(cp.target.path, filename))
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// open source file
	srcFile, err := sftpClient.Open(cp.source.path)
	if err != nil {
		log.Fatal(err)
	}

	bytes := [4096]byte{}
	s, err := srcFile.Stat()
	if err != nil {
		return err
	}

	fSize := s.Size()
	bytesCount := 0

	for {
		n, err := srcFile.Read(bytes[:])
		eof := err == io.EOF
		if err != nil && err != io.EOF {
			return err
		}

		bytesCount += n
		process := float64(bytesCount) / float64(fSize) * 100
		fmt.Print("\r" + filename + "\t" + fmt.Sprintf("%.2f", process) + "%")
		_, err = dstFile.Write(bytes[:n])
		if err != nil {
			return err
		}

		if eof {
			break
		}
	}
	fmt.Print("\r" + filename + "\t" + "100%    \n")

	// flush in-memory copy
	return dstFile.Sync()
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
