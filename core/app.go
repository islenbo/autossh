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
type Group struct {
	Name string `json:"group"`
	Servers []Server
}

type App struct {
	ServersPath string
	Groups     []Group
}

// 执行脚本
func (app *App) Exec() {
	b, err := ioutil.ReadFile(app.ServersPath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, &app.Groups)
	if err != nil {
		panic(errors.New("解析servers.json失败：" + err.Error()))
	}

	if len(os.Args) > 1 {
		option := os.Args[1]
		switch option {
		case "list":
			app.list()
		case "add":
			app.add(app.getArg(2, 3, ""))
		case "edit":
			app.edit(app.getArg(2, 3, ""))
		case "remove":
			app.remove(app.getArg(2, 3, ""))

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
	for i, group := range app.Groups {
		Printer.Logln(" ["+strconv.Itoa(i+1)+"]", group.Name)
	}
	Printer.Infoln("=======================================")

	group := app.inputGroupSh()
	Printer.Infoln("你选择了: " + group.Name)

	Printer.Infoln("========== 请选择服务器 ==========")
	for i, server := range group.Servers {
		Printer.Logln(" ["+strconv.Itoa(i+1)+"]", server.Name)
	}
	Printer.Logln(" ["+"-1"+"]", "返回")
	Printer.Infoln("=======================================")
	server := app.inputServerSh(group)
	if server.Name == "" {
		app.start()
		return
	}
	server.Connection()
}

// 编辑
func (app *App) edit(groupName, serverName string) {
	exists, groupIndex, serverIndex := app.serverExists(groupName, serverName)
	if !exists {
		Printer.Errorln("Server", serverName, "不存在")
		return
	}

	server := &app.Groups[groupIndex].Servers[serverIndex]
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
func (app *App) remove(groupName, serverName string) {
	exists, groupIndex, serverIndex := app.serverExists(groupName, serverName)

	if exists {
		app.Groups[groupIndex].Servers = append(app.Groups[groupIndex].Servers[:serverIndex], app.Groups[groupIndex].Servers[serverIndex+1:]...)
		err := app.saveServers()
		if err != nil {
			Printer.Errorln("保存到servers.json失败：", err)
		} else {
			Printer.Infoln("删除成功")
		}
	} else {
		Printer.Errorln("Server", serverName, "不存在")
	}
}

// 添加
func (app *App) add(groupName, serverName string) {
	if groupName == "" {
		Printer.Errorln("group name 不能为空")
		return
	}

	if serverName == "" {
		Printer.Errorln("server name 不能为空")
		return
	}

	if exists, _, _ := app.serverExists(groupName, serverName); exists {
		Printer.Errorln("server", serverName, "已存在")
		return
	}
	group := Group{Name: groupName}
	server := Server{Name: serverName}

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
	group.Servers = append(group.Servers, server)
	app.Groups = append(app.Groups, group)
	err := app.saveServers()
	if err != nil {
		Printer.Errorln("保存到servers.json失败：", err)
	} else {
		Printer.Infoln("添加成功")
	}
}

// 保存servers到servers.json文件
func (app *App) saveServers() error {
	b, err := json.Marshal(app.Groups)
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
	for _, group := range app.Groups {
		Printer.Logln(group.Name, ":")
		for _, server := range group.Servers {
			Printer.Logln("   ", server.Name)
		}
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
func (app *App) serverExists(groupName, serverName string) (bool, int, int) {
	for index, group := range app.Groups {
		for i, server := range group.Servers {
			if server.Name == serverName && group.Name == groupName {
				return true, index, i
			}
		}
	}

	return false, -1, -1
}

// 接收输入，获取对应sh脚本
func (app *App) inputGroupSh() Group {
	num, err := app.inputInfo(len(app.Groups))
	if err != nil || num == -1 {
		return app.inputGroupSh()
	}
	return app.Groups[num-1]
}

func (app *App) inputServerSh(group Group) Server {
	servers := group.Servers
	num, err := app.inputInfo(len(servers))
	if err != nil {
		return app.inputServerSh(group)
	}
	if num == -1 {
		return Server{}
	}
	return servers[num-1]
}

func (app *App) inputInfo(max int) (int, error) {
	Printer.Info("请输入序号: ")
	input := ""
	fmt.Scanln(&input)
	num, err := strconv.Atoi(input)
	if err != nil {
		Printer.Errorln("输入有误，请重新输入")
		return 0, err
	}
	if num == -1 {
		return -1, nil
	}

	if num <= 0 || num > max {
		Printer.Errorln("输入有误，请重新输入")
		return 0, errors.New("输入有误，请重新输入")
	}
	return num, nil
}

// 获取参数
func (app *App) getArg(groupIndex, serverIndex int, def string) (string, string) {
	max := len(os.Args) - 1
	if max >= groupIndex || max >= serverIndex {
		return os.Args[groupIndex], os.Args[serverIndex]
	}

	return def, def
}
