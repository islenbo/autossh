package core

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type Server struct {
	Name     string `json:"name"`
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Method   string `json:"method"`
	Key      string `json:"key"`
}

// 执行远程连接
func (server *Server) Connection(appConfig *Config) {
	auths, err := parseAuthMethods(server, appConfig)

	if err != nil {
		Printer.Errorln("鉴权出错:", err)
		return
	}

	config := &ssh.ClientConfig{
		User: server.User,
		Auth: auths,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr := server.Ip + ":" + strconv.Itoa(server.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		Printer.Errorln("建立连接出错:", err)
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		Printer.Errorln("创建Session出错:", err)
		return
	}

	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		Printer.Errorln("创建文件描述符出错:", err)
		return
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		Printer.Errorln("获取窗口宽高出错:", err)
		return
	}

	defer terminal.Restore(fd, oldState)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		Printer.Errorln("创建终端出错:", err)
		return
	}

	err = session.Shell()
	if err != nil {
		Printer.Errorln("执行Shell出错:", err)
		return
	}

	err = session.Wait()
	if err != nil {
		//Printer.Errorln("执行Wait出错:", err)
		return
	}
}

// 解析鉴权方式
func parseAuthMethods(server *Server, config *Config) ([]ssh.AuthMethod, error) {
	sshs := []ssh.AuthMethod{}

	switch server.Method {
	case "password":
		sshs = append(sshs, ssh.Password(server.Password))
		break

	case "pem":
		method, err := pemKey(server.Key, server.Password)
		if err != nil {
			return nil, err
		}
		sshs = append(sshs, method)
		break

	default:
		switch config.Method {
		case "password":
			sshs = append(sshs, ssh.Password(config.Password))
			break

		case "pem":
			method, err := pemKey(config.Key, config.Password)
			if err != nil {
				return nil, err
			}
			sshs = append(sshs, method)
			break

		default:
			return nil, errors.New("无效的密码方式: " + server.Method)
		}
	}

	return sshs, nil
}

// 解析pem密钥
func pemKey(key, password string) (ssh.AuthMethod, error) {
	pemBytes, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, err
	}

	var signer ssh.Signer
	if password == "" {
		signer, err = ssh.ParsePrivateKey(pemBytes)
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password))
	}

	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}
