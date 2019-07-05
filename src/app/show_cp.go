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

	source []*TransferObject
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

	if cp.source[0].cpType == CpTypeLocal {
		err = cp.upload()
	} else {
		//err = cp.download()
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

	cp.source = make([]*TransferObject, 0)
	for _, arg := range args[:length-1] {
		s, err := newTransferObject(*cp.cfg, arg)
		if err != nil {
			return err
		}

		if s.cpType == CpTypeLocal && s.cpType == cp.target.cpType {
			return errors.New("源和目标不能同时为本地地址")
		}

		cp.source = append(cp.source, s)
	}

	return nil
}

// 上传
func (cp *Cp) upload() error {
	client, err := cp.target.server.GetSftpClient()
	if err != nil {
		return err
	}

	defer func() {
		_ = client.Close()
	}()

	for _, source := range cp.source {
		file, err := cp.uploadFile(client, source.path)
		if err != nil {
			fmt.Println(file, ": ", err)
		}
	}

	return nil
}

// 上传单个文件
func (cp *Cp) uploadFile(client *sftp.Client, sourceFile string) (string, error) {
	if _, err := utils.FileIsExists(sourceFile); err != nil {
		return sourceFile, err
	}

	s, _ := os.Stat(sourceFile)
	if s.IsDir() {
		return sourceFile, errors.New("不是一个有效的文件")
	}

	srcFile, err := os.Open(sourceFile)
	if err != nil {
		return sourceFile, err
	}

	filename := path.Base(sourceFile)
	targetPath := cp.target.path

	targetFile, err := client.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			p := path.Dir(targetPath)
			if targetFile, err = client.Stat(p); err != nil {
				return cp.target.path, err
			}

			filename = path.Base(targetPath)
			targetPath = p
		} else {
			return cp.target.path, err
		}
	}

	t := targetPath
	if targetFile.IsDir() {
		t = path.Join(t, filename)
	}

	// create destination file
	dstFile, err := client.Create(t)
	if err != nil {
		return cp.target.path, err
	}

	defer func() {
		_ = dstFile.Close()
	}()

	return cp.showCopy(srcFile, dstFile, s.Size(), sourceFile)
}

// 显示复制进度
func (cp *Cp) showCopy(srcFile io.Reader, dstFile io.Writer, fSize int64, sourceFile string) (string, error) {
	// copy source file to destination file
	bytes := [4096]byte{}
	bytesCount := 0

	for {
		n, err := srcFile.Read(bytes[:])
		eof := err == io.EOF
		if err != nil && err != io.EOF {
			return sourceFile, err
		}

		bytesCount += n
		process := float64(bytesCount) / float64(fSize) * 100
		fmt.Print("\r" + sourceFile + "\t\t" + fmt.Sprintf("%.2f", process) + "%")
		_, err = dstFile.Write(bytes[:n])
		if err != nil {
			return cp.target.path, err
		}

		if eof {
			break
		}
	}

	fmt.Print("\r"+sourceFile+"\t\t"+"100%    ", "\n")
	return "", nil
}

// 下载
//func (cp *Cp) download() error {
//	sftpClient, err := cp.source.server.GetSftpClient()
//	if err != nil {
//		return err
//	}
//
//	defer sftpClient.Close()
//
//	// create destination file
//	filename := path.Base(cp.source.path)
//	dstFile, err := os.Create(path.Join(cp.target.path, filename))
//	if err != nil {
//		return err
//	}
//	defer dstFile.Close()
//
//	// open source file
//	srcFile, err := sftpClient.Open(cp.source.path)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	bytes := [4096]byte{}
//	s, err := srcFile.Stat()
//	if err != nil {
//		return err
//	}
//
//	fSize := s.Size()
//	bytesCount := 0
//
//	for {
//		n, err := srcFile.Read(bytes[:])
//		eof := err == io.EOF
//		if err != nil && err != io.EOF {
//			return err
//		}
//
//		bytesCount += n
//		process := float64(bytesCount) / float64(fSize) * 100
//		fmt.Print("\r" + filename + "\t" + fmt.Sprintf("%.2f", process) + "%")
//		_, err = dstFile.Write(bytes[:n])
//		if err != nil {
//			return err
//		}
//
//		if eof {
//			break
//		}
//	}
//	fmt.Print("\r" + filename + "\t" + "100%    \n")
//
//	// flush in-memory copy
//	return dstFile.Sync()
//}

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
