package app

import (
	"github.com/pkg/sftp"
	"io/ioutil"
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

type IOClient interface {
	Stat(file string) (os.FileInfo, error)
	Mkdir(path string) error
	Create(file string) (FileLike, error)
	Open(file string) (FileLike, error)
	ReadDir(file string) ([]os.FileInfo, error)
}

// Local
type LocalIOClient struct {
}

func (client *LocalIOClient) Stat(file string) (os.FileInfo, error) {
	return os.Stat(file)
}

func (client *LocalIOClient) Mkdir(path string) error {
	return os.Mkdir(path, 0755)
}

func (client *LocalIOClient) Create(file string) (FileLike, error) {
	return os.Create(file)
}

func (client *LocalIOClient) Open(file string) (FileLike, error) {
	return os.Open(file)
}

func (client *LocalIOClient) ReadDir(file string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(file)
}

// SFTP(Remote)
type SftpIOClient struct {
	SftpClient *sftp.Client
}

func (client *SftpIOClient) Stat(file string) (os.FileInfo, error) {
	return client.SftpClient.Stat(file)
}

func (client *SftpIOClient) Mkdir(path string) error {
	return client.SftpClient.Mkdir(path)
}

func (client *SftpIOClient) Create(file string) (FileLike, error) {
	return client.SftpClient.Create(file)
}

func (client *SftpIOClient) Open(file string) (FileLike, error) {
	return client.SftpClient.Open(file)
}

func (client *SftpIOClient) ReadDir(file string) ([]os.FileInfo, error) {
	return client.SftpClient.ReadDir(file)
}
