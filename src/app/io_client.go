package app

import (
	"github.com/pkg/sftp"
	"os"
)

type IOClientType int

type FileLike interface {
	Name() string
	Stat() (os.FileInfo, error)
	Read([]byte) (int, error)
	Close() error
	Write(p []byte) (n int, err error)
}

const (
	IOClientLocal IOClientType = iota
	IOClientSftp
)

type IOClient struct {
	ClientType IOClientType
	SftpClient *sftp.Client
}

// io stat
func (client *IOClient) Stat(file string) (os.FileInfo, error) {
	switch client.ClientType {
	case IOClientLocal:
		return os.Stat(file)
	case IOClientSftp:
		return client.SftpClient.Stat(file)
	default:
		return os.Stat(file)
	}
}

// io mkdir
func (client *IOClient) Mkdir(path string) error {
	switch client.ClientType {
	case IOClientLocal:
		return os.Mkdir(path, 0755)
	case IOClientSftp:
		return client.SftpClient.Mkdir(path)
	default:
		return os.Mkdir(path, 0755)
	}
}

// io create
func (client *IOClient) Create(file string) (FileLike, error) {
	switch client.ClientType {
	case IOClientLocal:
		return os.Create(file)
	case IOClientSftp:
		return client.SftpClient.Create(file)
	default:
		return os.Create(file)
	}
}
