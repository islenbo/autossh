package app

import (
	"autossh/src/utils"
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func (server *Server) Format() {
	if server.Port == 0 {
		server.Port = 22
	}

	if server.Method == "" {
		server.Method = "password"
	}
}

// 合并选项
func (server *Server) MergeOptions(options map[string]interface{}, overwrite bool) {
	if server.Options == nil {
		server.Options = make(map[string]interface{})
	}

	for k, v := range options {
		if overwrite {
			server.Options[k] = v
		} else {
			if _, ok := server.Options[k]; !ok {
				server.Options[k] = v
			}
		}

	}
}

func (server *Server) Edit() {
	input := ""
	utils.Info("Name(default=" + server.Name + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Name = input
		input = ""
	}

	utils.Info("Ip(default=" + server.Ip + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Ip = input
		input = ""
	}

	utils.Info("Port(default=" + strconv.Itoa(server.Port) + ")：")
	fmt.Scanln(&input)
	if input != "" {
		port, _ := strconv.Atoi(input)
		server.Port = port
		input = ""
	}

	utils.Info("User(default=" + server.User + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.User = input
		input = ""
	}

	utils.Info("Password(default=" + server.Password + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Password = input
		input = ""
	}

	utils.Info("Method(default=" + server.Method + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Method = input
		input = ""
	}

	utils.Info("Key(default=" + server.Key + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Key = input
		input = ""
	}
}

// 生成SSH Client
func (server *Server) GetSshClient() (*ssh.Client, error) {
	auth, err := parseAuthMethods(server)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: server.User,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// 默认端口为22
	if server.Port == 0 {
		server.Port = 22
	}

	addr := server.Ip + ":" + strconv.Itoa(server.Port)

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// 生成Sftp Client
func (server *Server) GetSftpClient() (*sftp.Client, error) {
	sshClient, err := server.GetSshClient()
	if err == nil {
		return sftp.NewClient(sshClient)
	} else {
		return nil, err
	}
}

// 执行远程连接
func (server *Server) Connect() error {
	client, err := server.GetSshClient()
	if err != nil {
		if utils.ErrorAssert(err, "ssh: unable to authenticate") {
			return errors.New("连接失败，请检查密码/密钥是否有误")
		}

		return errors.New("ssh dial fail:" + err.Error())
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return errors.New("create session fail:" + err.Error())
	}

	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return errors.New("创建文件描述符出错:" + err.Error())
	}

	stopKeepAliveLoop := server.startKeepAliveLoop(session)
	defer close(stopKeepAliveLoop)

	server.stdIO(session)

	defer terminal.Restore(fd, oldState)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	server.termWidth, server.termHeight, _ = terminal.GetSize(fd)
	if err := session.RequestPty("xterm-256color", server.termHeight, server.termWidth, modes); err != nil {
		return errors.New("创建终端出错:" + err.Error())
	}

	winChange := server.listenWindowChange(session, fd)
	defer close(winChange)

	err = session.Shell()
	if err != nil {
		return errors.New("执行Shell出错:" + err.Error())
	}

	_ = session.Wait()
	//if err != nil {
	//	return errors.New("执行Wait出错:" + err.Error())
	//}

	return nil
}

// 重定向标准输入输出
func (server *Server) stdIO(session *ssh.Session) {
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	if server.Log.Enable {
		ch, _ := session.StdoutPipe()

		go func() {
			flag := os.O_RDWR | os.O_CREATE
			switch server.Log.Mode {
			case LogModeAppend:
				flag = flag | os.O_APPEND
			case LogModeCover:
			}

			f, _ := os.OpenFile(server.formatLogFilename(server.Log.Filename), flag, 0644)

			for {
				buff := [128]byte{}
				n, _ := ch.Read(buff[:])
				if n > 0 {
					f.Write(buff[:n])
					os.Stdout.Write(buff[:n])
				}
			}
		}()
	} else {
		session.Stdout = os.Stdout
	}
}

// 格式化日志文件名
func (server *Server) formatLogFilename(filename string) string {
	kvs := map[string]string{
		"%g":  server.groupName,
		"%n":  server.Name,
		"%dt": time.Now().Format("2006-01-02-15-04-05"),
		"%d":  time.Now().Format("2006-01-02"),
		"%u":  server.User,
		"%a":  server.Alias,
	}

	for k, v := range kvs {
		filename = strings.ReplaceAll(filename, k, v)
	}

	return filename
}

// 解析鉴权方式
func parseAuthMethods(server *Server) ([]ssh.AuthMethod, error) {
	var sshs []ssh.AuthMethod

	switch strings.ToLower(server.Method) {
	case "password":
		sshs = append(sshs, ssh.Password(server.Password))
		break

	case "key":
		method, err := pemKey(server)
		if err != nil {
			return nil, err
		}
		sshs = append(sshs, method)
		break

		// 默认以password方式
	default:
		sshs = append(sshs, ssh.Password(server.Password))
	}

	return sshs, nil
}

// 解析密钥
func pemKey(server *Server) (ssh.AuthMethod, error) {
	if server.Key == "" {
		server.Key = "~/.ssh/id_rsa"
	}
	server.Key, _ = utils.ParsePath(server.Key)

	pemBytes, err := ioutil.ReadFile(server.Key)
	if err != nil {
		return nil, err
	}

	var signer ssh.Signer
	if server.Password == "" {
		signer, err = ssh.ParsePrivateKey(pemBytes)
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(server.Password))
	}

	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}

// 发送心跳包
func (server *Server) startKeepAliveLoop(session *ssh.Session) chan struct{} {
	terminate := make(chan struct{})
	go func() {
		for {
			select {
			case <-terminate:
				return
			default:
				if val, ok := server.Options["ServerAliveInterval"]; ok && val != nil {
					_, err := session.SendRequest("keepalive@bbr", true, nil)
					if err != nil {
						//Log.Category("server").Error("keepAliveLoop fail", err)
						// TODO 错误日志
					}

					t := time.Duration(server.Options["ServerAliveInterval"].(float64))
					time.Sleep(time.Second * t)
				} else {
					return
				}
			}
		}
	}()
	return terminate
}

// 监听终端窗口变化
func (server *Server) listenWindowChange(session *ssh.Session, fd int) chan struct{} {
	terminate := make(chan struct{})
	go func() {
		for {
			select {
			case <-terminate:
				return
			default:
				termWidth, termHeight, _ := terminal.GetSize(fd)

				if server.termWidth != termWidth || server.termHeight != termHeight {
					server.termHeight = termHeight
					server.termWidth = termWidth
					session.WindowChange(termHeight, termWidth)
				}

				time.Sleep(time.Millisecond * 3)
			}
		}
	}()

	return terminate
}
