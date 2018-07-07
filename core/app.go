package core

import (
	"io/ioutil"
	"encoding/json"
	"errors"
	"strconv"
	"fmt"
	"strings"
)

type Config struct {
	LogPath    string                 `json:"log_path"`
	ShowDetail bool                   `json:"show_detail"`
	Servers    []Server               `json:"servers"`
	Groups     []Group                `json:"groups"`
	Options    map[string]interface{} `json:"options"`
}

type App struct {
	ConfigPath string
	Version    string
	Build      string
	config     Config
	serverMap  map[string]Server
}

// 执行脚本
func (app *App) Exec() {
	app.serverMap = make(map[string]Server)

	// 解析配置
	app.loadConfig()

	app.loadServerMap()

	//for {
		// 输出server
		app.showServers()

		// 监听输入
		input, isGlobal := app.checkInput()
		if isGlobal {
			end := app.handleGlobalCmd(input)
			if end {
				return
			}
		} else {
			server := app.serverMap[input]
			Printer.Infoln("你选择了", server.Name)
			server.Connect()
		}
	//}
}

func (app *App) handleGlobalCmd(cmd string) bool {
	switch strings.ToLower(cmd) {
	case "exit":
		return true
	default:
		Printer.Errorln("指令无效")
		return false
	}
}

// 检查输入
func (app *App) checkInput() (string, bool) {
	flag := ""
	for {
		fmt.Scanln(&flag)

		if app.isGlobalInput(flag) {
			return flag, true
		}

		if _, ok := app.serverMap[flag]; !ok {
			Printer.Errorln("输入有误，请重新输入")
		} else {
			return flag, false
		}
	}

	panic(errors.New("输入有误"))
}

// 判断是否全局输入
func (app *App) isGlobalInput(flag string) bool {
	switch flag {
	case "edit":
		return true
	case "exit":
		return true
	default:
		return false
	}
}

// 加载配置文件
func (app *App) loadConfig() {
	b, _ := ioutil.ReadFile(app.ConfigPath)
	err := json.Unmarshal(b, &app.config)
	if err != nil {
		panic(errors.New("加载配置文件失败：" + err.Error()))
	}
}

// 打印列表
func (app *App) showServers() {
	Printer.Infoln("========== 欢迎使用 Auto SSH ==========")
	for i, server := range app.config.Servers {
		Printer.Logln(app.recordServer(strconv.Itoa(i+1), server))
	}

	for _, group := range app.config.Groups {
		if len(group.Servers) == 0 {
			continue
		}

		Printer.Infoln("____________________ " + group.GroupName + " ____________________")
		for i, server := range group.Servers {
			Printer.Logln(app.recordServer(group.Prefix+strconv.Itoa(i+1), server))
		}
	}

	Printer.Infoln("\n=======================================")
	Printer.Logln(" [edit]", " ", "编辑")
	Printer.Logln(" [exit]", " ", "退出")
	Printer.Infoln("=======================================")
	Printer.Info("请输入序号或操作: ")
}

// 加载
func (app *App) loadServerMap() {
	for i, server := range app.config.Servers {
		flag := strconv.Itoa(i + 1)

		if _, ok := app.serverMap[flag]; ok {
			panic(errors.New("标识[" + flag + "]已存在，请检查您的配置文件"))
		}

		server.MergeOptions(app.config.Options, false)
		app.serverMap[flag] = server
	}

	for _, group := range app.config.Groups {
		for i, server := range group.Servers {
			flag := group.Prefix + strconv.Itoa(i+1)

			if _, ok := app.serverMap[flag]; ok {
				panic(errors.New("标识[" + flag + "]已存在，请检查您的配置文件"))
			}

			server.MergeOptions(app.config.Options, false)
			app.serverMap[flag] = server
		}
	}
}

func (app *App) recordServer(flag string, server Server) string {
	if app.config.ShowDetail {
		return " [" + flag + "]" + "\t" + server.Name + " [" + server.User + "@" + server.Ip + "]"
	} else {
		return " [" + flag + "]" + "\t" + server.Name
	}
}
