package app

import (
	"autossh/src/utils"
	"errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
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
	Alias    string                 `json:"alias"`
	Log      ServerLog              `json:"log"`

	termWidth  int
	termHeight int
	groupName  string
}

// 格式化，赋予默认值
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

// 格式化输出，用于打印
func (server *Server) FormatPrint(flag string, ShowDetail bool) string {
	alias := ""
	if server.Alias != "" {
		alias = "|" + server.Alias
	}

	if ShowDetail {
		return " [" + flag + alias + "]" + "\t" + server.Name + " [" + server.User + "@" + server.Ip + "]"
	} else {
		return " [" + flag + alias + "]" + "\t" + server.Name
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
	defer terminal.Restore(fd, oldState)

	stopKeepAliveLoop := server.startKeepAliveLoop(session)
	defer close(stopKeepAliveLoop)

	err = server.stdIO(session)
	if err != nil {
		return err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	server.termWidth, server.termHeight, _ = terminal.GetSize(fd)
	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}
	if err := session.RequestPty(termType, server.termHeight, server.termWidth, modes); err != nil {
		return errors.New("创建终端出错:" + err.Error())
	}

	server.listenWindowChange(session, fd)

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
func (server *Server) stdIO(session *ssh.Session) error {
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	if server.Log.Enable {
		ch, err := session.StdoutPipe()
		if err != nil {
			return err
		}

		go func() {
			flag := os.O_RDWR | os.O_CREATE
			switch server.Log.Mode {
			case LogModeAppend:
				flag = flag | os.O_APPEND
			case LogModeCover:
			}

			f, err := os.OpenFile(server.formatLogFilename(server.Log.Filename), flag, 0644)
			if err != nil {
				utils.Logger.Error("Open file fail ", err)
				return
			}

			for {
				buff := [4096]byte{}
				n, _ := ch.Read(buff[:])
				if n > 0 {
					if _, err := f.Write(buff[:n]); err != nil {
						utils.Logger.Error("Write file buffer fail ", err)
					}

					if _, err := os.Stdout.Write(buff[:n]); err != nil {
						utils.Logger.Error("Write stdout buffer fail ", err)
					}
				}
			}
		}()
	} else {
		session.Stdout = os.Stdout
	}

	return nil
}

// 格式化日志文件名
func (server *Server) formatLogFilename(filename string) string {
	kvs := map[string]string{
		"%g":  server.groupName,
		"%n":  server.Name,
		"%dt": time.Now().Format("20060102-150405"),
		"%d":  time.Now().Format("20060102"),
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
						utils.Logger.Category("server").Error("keepAliveLoop fail", err)
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
func (server *Server) listenWindowChange(session *ssh.Session, fd int) {
	go func() {
		sigwinchCh := make(chan os.Signal, 1)
		defer close(sigwinchCh)

		signal.Notify(sigwinchCh, syscall.SIGWINCH)
		termWidth, termHeight, err := terminal.GetSize(fd)
		if err != nil {
			utils.Logger.Error(err)
		}

		for {
			select {
			// 阻塞读取
			case sigwinch := <-sigwinchCh:
				if sigwinch == nil {
					return
				}
				currTermWidth, currTermHeight, err := terminal.GetSize(fd)

				// 判断一下窗口尺寸是否有改变
				if currTermHeight == termHeight && currTermWidth == termWidth {
					continue
				}

				// 更新远端大小
				session.WindowChange(currTermHeight, currTermWidth)
				if err != nil {
					utils.Logger.Error(err)
					continue
				}

				termWidth, termHeight = currTermWidth, currTermHeight
			}
		}
	}()
}
