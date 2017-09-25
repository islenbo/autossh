package core

import (
	"fmt"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"os"
	"bytes"
	"errors"
)

// 版本号
const VERSION = "0.2"

type App struct {
	ServersPath string
	servers     []Server
}

// 执行脚本
func (app *App) Exec() {
	b, err := ioutil.ReadFile(app.ServersPath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, &app.servers)
	if err != nil {
		panic(errors.New("解析servers.json失败：" + err.Error()))
	}

	if len(os.Args) > 1 {
		option := os.Args[1]
		switch option {
		case "list":
			app.list()
		case "add":
			app.add(app.getArg(2, ""))
		case "edit":
			app.edit(app.getArg(2, ""))
		case "remove":
			app.remove(app.getArg(2, ""))

		case "--version":
			fallthrough
		case "-v":
			app.version()

		case "--help":
			fallthrough
		case "-h":
			fallthrough
		default:
			app.help()
		}
	} else {
		app.start()
	}
}

// 启动脚本
func (app *App) start() {
	Printer.Infoln("========== 欢迎使用 Auto SSH ==========")
	for i, server := range app.servers {
		Printer.Logln(" ["+strconv.Itoa(i+1)+"]", server.Name)
	}
	Printer.Infoln("=======================================")

	server := app.inputSh()
	Printer.Infoln("你选择了: " + server.Name)
	server.Connection()
}

// 编辑
func (app *App) edit(name string) {
	exists, index := app.serverExists(name)
	if !exists {
		Printer.Errorln("Server", name, "不存在")
		return
	}

	server := &app.servers[index]
	var def string

	def = server.Ip
	Printer.Info("请输入IP(default: " + def + ")：")
	fmt.Scanln(&server.Ip)
	if server.Ip == "" {
		server.Ip = def
	}

	def = strconv.Itoa(server.Port)
	Printer.Info("请输入Port(default: " + def + ")：")
	fmt.Scanln(&server.Port)
	if server.Port == 0 {
		port, err := strconv.Atoi(def)

		if err != nil {
			Printer.Errorln("Port illegality")
			return
		} else {
			server.Port = port
		}
	}

	def = server.User
	Printer.Info("请输入User(default: " + def + ")：")
	fmt.Scanln(&server.User)
	if server.User == "" {
		server.User = def
	}

	def = server.Method
	Printer.Info("请输入Method[password/pem](default: " + def + ")：")
	fmt.Scanln(&server.Method)
	if server.Method == "" {
		server.Method = def
	}

	server.Password = ""
	if server.Method == "pem" {
		def = server.Key
		Printer.Info("请输入pem证书绝对目录(default: " + def + ")：")
		fmt.Scanln(&server.Key)
		if server.Key == "" {
			server.Key = def
		}

		Printer.Info("请输入pem证书密码（若无请留空）：")
		fmt.Scanln(&server.Password)
	} else {
		Printer.Info("请输入Password（若无请留空）：")
		fmt.Scanln(&server.Password)
	}

	err := app.saveServers()
	if err != nil {
		Printer.Errorln("保存到servers.json失败：", err)
	} else {
		Printer.Infoln("保存成功")
	}
}

// 删除
func (app *App) remove(name string) {
	exists, index := app.serverExists(name)

	if exists {
		app.servers = append(app.servers[:index], app.servers[index+1:]...)
		err := app.saveServers()
		if err != nil {
			Printer.Errorln("保存到servers.json失败：", err)
		} else {
			Printer.Infoln("删除成功")
		}
	} else {
		Printer.Errorln("Server", name, "不存在")
	}
}

// 添加
func (app *App) add(name string) {
	if name == "" {
		Printer.Errorln("server name 不能为空")
		return
	}

	if exists, _ := app.serverExists(name); exists {
		Printer.Errorln("server", name, "已存在")
		return
	}

	server := Server{Name: name}

	Printer.Info("请输入IP：")
	fmt.Scanln(&server.Ip)

	Printer.Info("请输入Port：")
	fmt.Scanln(&server.Port)

	Printer.Info("请输入User：")
	fmt.Scanln(&server.User)

	Printer.Info("请输入Method[password/pem]：")
	fmt.Scanln(&server.Method)

	if server.Method == "pem" {
		Printer.Info("请输入pem证书绝对目录：")
		fmt.Scanln(&server.Key)

		Printer.Info("请输入pem证书密码（若无请留空）：")
		fmt.Scanln(&server.Password)
	} else {
		Printer.Info("请输入Password（若无请留空）：")
		fmt.Scanln(&server.Password)
	}

	app.servers = append(app.servers, server)
	err := app.saveServers()
	if err != nil {
		Printer.Errorln("保存到servers.json失败：", err)
	} else {
		Printer.Infoln("添加成功")
	}
}

// 保存servers到servers.json文件
func (app *App) saveServers() error {
	b, err := json.Marshal(app.servers)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(app.ServersPath, out.Bytes(), os.ModePerm)
}

// 打印列表
func (app *App) list() {
	for _, server := range app.servers {
		Printer.Logln(server.Name)
	}
}

// 版本信息
func (app *App) version() {
	fmt.Println("Autossh", VERSION, "。")
	fmt.Println("由 Lenbo 编写，项目地址：https://github.com/islenbo/autossh。")
}

// 显示帮助信息
func (app *App) help() {
	fmt.Println("go写的一个ssh远程客户端。可一键登录远程服务器，主要用来弥补Mac/Linux Terminal ssh无法保存密码的不足。")
	fmt.Println("基本用法：")
	fmt.Println("  直接输入autossh不带任何参数，列出所有服务器，输入对应编号登录。")
	fmt.Println("参数：")
	fmt.Println("  -v, --version", "\t", "显示 autossh 的版本信息。")
	fmt.Println("  -h, --help   ", "\t", "显示帮助信息。")
	fmt.Println("操作：")
	fmt.Println("  list         ", "\t", "显示所有server。")
	fmt.Println("  add <name>   ", "\t", "添加一个 server。如：autossh add vagrant。")
	fmt.Println("  edit <name>  ", "\t", "编辑一个 server。如：autossh edit vagrant。")
	fmt.Println("  remove <name>", "\t", "删除一个 server。如：autossh remove vagrant。")
}

// 判断server是否存在
func (app *App) serverExists(name string) (bool, int) {
	for index, server := range app.servers {
		if server.Name == name {
			return true, index
		}
	}

	return false, -1
}

// 接收输入，获取对应sh脚本
func (app *App) inputSh() Server {
	Printer.Info("请输入序号: ")
	input := ""
	fmt.Scanln(&input)
	num, err := strconv.Atoi(input)
	if err != nil {
		Printer.Errorln("输入有误，请重新输入")
		return app.inputSh()
	}

	if num <= 0 || num > len(app.servers) {
		Printer.Errorln("输入有误，请重新输入")
		return app.inputSh()
	}

	return app.servers[num-1]
}

// 获取参数
func (app *App) getArg(index int, def string) string {
	max := len(os.Args) - 1
	if max >= index {
		return os.Args[index]
	}

	return def
}
