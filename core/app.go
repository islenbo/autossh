package core

import (
	"io/ioutil"
	"encoding/json"
	"errors"
	"strconv"
	"fmt"
	"strings"
	"bytes"
	"os"
	"io"
	"path/filepath"
	"time"
)

type Group struct {
	GroupName string   `json:"group_name"`
	Prefix    string   `json:"prefix"`
	Servers   []Server `json:"servers"`
}

type Config struct {
	ShowDetail bool                   `json:"show_detail"`
	Servers    []Server               `json:"servers"`
	Groups     []Group                `json:"groups"`
	Options    map[string]interface{} `json:"options"`
}

type App struct {
	ConfigPath string
	config     Config
	serverMap  map[string]*Server
}

// 执行脚本
func (app *App) Init() {
	app.serverMap = make(map[string]*Server)

	// 解析配置
	app.loadConfig()

	app.loadServerMap()

	app.show()
}

func (app *App) show() {
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
		Log.Category("app").Info("select server", server.Name)
		server.Connect()
	}
	//}
}

func (app *App) handleGlobalCmd(cmd string) bool {
	switch strings.ToLower(cmd) {
	case "exit":
		return true
	case "edit":
		app.handleEdit()
		return false
	default:
		Printer.Errorln("指令无效")
		return false
	}
}

// 编辑
func (app *App) handleEdit() {
	Printer.Info("请输入相应序号（exit退出编辑）：")
	id := ""
	fmt.Scanln(&id)

	if strings.ToLower(id) == "exit" {
		app.show()
		return
	}

	server, ok := app.serverMap[id]
	if !ok {
		Printer.Errorln("序号不存在")
		app.handleEdit()
		return
	}

	server.Edit()
	app.saveConfig()
	app.show()
}

// 保存配置文件
func (app *App) saveConfig() error {
	b, err := json.Marshal(app.config)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	err = app.backConfig()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(app.ConfigPath, out.Bytes(), os.ModePerm)
}

func (app *App) backConfig() error {
	srcFile, err := os.Open(app.ConfigPath)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	path, _ := filepath.Abs(filepath.Dir(app.ConfigPath))
	backupFile := path + "/config-" + time.Now().Format("20060102150405") + ".json"
	desFile, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer desFile.Close()

	_, err = io.Copy(desFile, srcFile)
	if err != nil {
		return err
	}

	Printer.Infoln("配置文件已备份：", backupFile)
	return nil
}

// 检查输入
func (app *App) checkInput() (string, bool) {
	flag := ""
	for {
		fmt.Scanln(&flag)
		Log.Category("app").Info("input scan:", flag)

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
		Printer.Errorln("加载配置文件失败", err)
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

	Printer.Logln("")
	Printer.Infoln("=======================================")
	Printer.Logln(" [edit]", " ", "编辑")
	Printer.Logln(" [exit]", " ", "退出")
	Printer.Infoln("=======================================")
	Printer.Info("请输入序号或操作: ")
}

// 加载
func (app *App) loadServerMap() {
	Log.Category("app").Info("server count", len(app.config.Servers), "group count", len(app.config.Groups))

	for i := range app.config.Servers {
		server := &app.config.Servers[i]
		flag := strconv.Itoa(i + 1)

		if _, ok := app.serverMap[flag]; ok {
			panic(errors.New("标识[" + flag + "]已存在，请检查您的配置文件"))
		}

		server.MergeOptions(app.config.Options, false)
		app.serverMap[flag] = server
	}

	for i := range app.config.Groups {
		group := &app.config.Groups[i]
		for j := range group.Servers {
			server := &group.Servers[j]
			flag := group.Prefix + strconv.Itoa(j+1)

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
