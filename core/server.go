package core

import (
	"os"
	"net"
	"strconv"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"time"
	"fmt"
	"strings"
)

type Server struct {
	Name     string                 `json:"name"`
	Ip       string                 `json:"ip"`
	Port     int                    `json:"port"`
	User     string                 `json:"user"`
	Password string                 `json:"password"`
	Method   string                 `json:"method"`
	Key      string                 `json:"key"`
	Options  map[string]interface{} `json:"options"`

	termWidth  int
	termHeight int
}

func (server *Server) Format() {
	if server.Port == 0 {
		server.Port = 22
	}

	if server.Method == "" {
		server.Method = "password"
	}
}

// 执行远程连接
func (server *Server) Connect() {
	auths, err := parseAuthMethods(server)

	if err != nil {
		Printer.Errorln("鉴权出错:", err)
		Log.Category("server").Error("auth fail", err)
		return
	}

	config := &ssh.ClientConfig{
		User: server.User,
		Auth: auths,
		Config: ssh.Config{
			Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
		},
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
		if ErrorAssert(err, "ssh: unable to authenticate") {
			Printer.Errorln("连接失败，请检查密码/密钥是否有误")
			return
		}

		Printer.Errorln("ssh dial fail:", err)
		Log.Category("server").Error("ssh dial fail", err)
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		Printer.Errorln("create session fail:", err)
		Log.Category("server").Error("create session fail", err)
		return
	}

	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		Printer.Errorln("创建文件描述符出错:", err)
		Log.Category("server").Error("创建文件描述符出错", err)
		return
	}

	stopKeepAliveLoop := server.startKeepAliveLoop(session)
	defer close(stopKeepAliveLoop)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	defer terminal.Restore(fd, oldState)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	server.termWidth, server.termHeight, _ = terminal.GetSize(fd)
	if err := session.RequestPty("xterm-256color", server.termHeight, server.termWidth, modes); err != nil {
		Printer.Errorln("创建终端出错:", err)
		Log.Category("server").Error("创建终端出错", err)
		return
	}

	winChange := server.listenWindowChange(session, fd)
	defer close(winChange)

	err = session.Shell()
	if err != nil {
		Printer.Errorln("执行Shell出错:", err)
		Log.Category("server").Error("执行Shell出错", err)
		return
	}

	err = session.Wait()
	if err != nil {
		//Printer.Errorln("执行Wait出错:", err)
		Log.Category("server").Error("执行Wait出错", err)
		return
	}
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
						Log.Category("server").Error("keepAliveLoop fail", err)
					}

					t := time.Duration(server.Options["ServerAliveInterval"].(float64))
					time.Sleep(time.Second * t)
				}
			}
		}
	}()
	return terminate
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

// 解析鉴权方式
func parseAuthMethods(server *Server) ([]ssh.AuthMethod, error) {
	sshs := []ssh.AuthMethod{}

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
	server.Key, _ = ParsePath(server.Key)

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

func (server *Server) Edit() {
	input := ""
	Printer.Info("Name(default=" + server.Name + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Name = input
		input = ""
	}

	Printer.Info("Ip(default=" + server.Ip + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Ip = input
		input = ""
	}

	Printer.Info("Port(default=" + strconv.Itoa(server.Port) + ")：")
	fmt.Scanln(&input)
	if input != "" {
		port, _ := strconv.Atoi(input)
		server.Port = port
		input = ""
	}

	Printer.Info("User(default=" + server.User + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.User = input
		input = ""
	}

	Printer.Info("Password(default=" + server.Password + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Password = input
		input = ""
	}

	Printer.Info("Method(default=" + server.Method + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Method = input
		input = ""
	}

	Printer.Info("Key(default=" + server.Key + ")：")
	fmt.Scanln(&input)
	if input != "" {
		server.Key = input
		input = ""
	}
}
