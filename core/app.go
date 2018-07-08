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

type IndexType int

const (
	IndexTypeServer IndexType = iota
	IndexTypeGroup
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

type ServerIndex struct {
	indexType   IndexType
	groupIndex  int
	serverIndex int
	server      *Server
}

type App struct {
	ConfigPath  string
	config      Config
	serverIndex map[string]ServerIndex
}

// 执行脚本
func (app *App) Init() {
	app.serverIndex = make(map[string]ServerIndex)

	// 解析配置
	app.loadConfig()

	app.loadServerMap(true)

	app.show()
}

func (app *App) saveAndReload() {
	app.saveConfig()
	app.loadConfig()
	app.loadServerMap(false)
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
		server := app.serverIndex[input].server
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
	case "add":
		app.handleAdd()
		return false
	case "remove":
		app.handleRemove()
		return false
	default:
		Printer.Errorln("指令无效")
		return false
	}
}

// 编辑
func (app *App) handleEdit() {
	Printer.Info("请输入相应序号（exit退出当前操作）：")
	id := ""
	fmt.Scanln(&id)

	if strings.ToLower(id) == "exit" {
		app.show()
		return
	}

	serverIndex, ok := app.serverIndex[id]
	if !ok {
		Printer.Errorln("序号不存在")
		app.handleEdit()
		return
	}

	serverIndex.server.Edit()
	app.saveAndReload()
}

// 移除
func (app *App) handleRemove() {
	Printer.Info("请输入相应序号（exit退出当前操作）：")
	id := ""
	fmt.Scanln(&id)

	if strings.ToLower(id) == "exit" {
		app.show()
		return
	}

	serverIndex, ok := app.serverIndex[id]
	if !ok {
		Printer.Errorln("序号不存在")
		app.handleEdit()
		return
	}

	if serverIndex.indexType == IndexTypeServer {
		servers := app.config.Servers
		app.config.Servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
	} else {
		servers := app.config.Groups[serverIndex.groupIndex].Servers
		servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
		app.config.Groups[serverIndex.groupIndex].Servers = servers
	}

	app.saveAndReload()
}

// 新增
func (app *App) handleAdd() {
	groups := make(map[string]*Group)
	for i := range app.config.Groups {
		group := &app.config.Groups[i]
		groups[group.Prefix] = group
		Printer.Info("["+group.Prefix+"]"+group.GroupName, "\t")
	}
	Printer.Infoln("[其他值]默认组")
	Printer.Info("请输入要插入的组：")
	g := ""
	fmt.Scanln(&g)

	server := Server{}
	server.Format()
	server.Edit()

	group, ok := groups[g]
	if ok {
		group.Servers = append(group.Servers, server)
	} else {
		app.config.Servers = append(app.config.Servers, server)
	}

	app.saveAndReload()
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

		if _, ok := app.serverIndex[flag]; !ok {
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
		fallthrough
	case "add":
		fallthrough
	case "remove":
		fallthrough
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
	maxlen := app.separatorLength()
	app.formatSeparator(" 欢迎使用 Auto SSH ", "=", maxlen)
	for i, server := range app.config.Servers {
		Printer.Logln(app.recordServer(strconv.Itoa(i+1), server))
	}

	for _, group := range app.config.Groups {
		if len(group.Servers) == 0 {
			continue
		}

		app.formatSeparator(" "+group.GroupName+" ", "_", maxlen)
		for i, server := range group.Servers {
			Printer.Logln(app.recordServer(group.Prefix+strconv.Itoa(i+1), server))
		}
	}

	app.formatSeparator("", "=", maxlen)
	Printer.Logln("", "[add]  添加", "    ", "[edit] 编辑", "    ", "[remove]删除")
	Printer.Logln("", "[exit]\t退出")
	app.formatSeparator("", "=", maxlen)
	Printer.Info("请输入序号或操作: ")
}

func (app *App) formatSeparator(title string, c string, maxlength float64) {

	charslen := int((maxlength - ZhLen(title)) / 2)
	chars := ""
	for i := 0; i < charslen; i ++ {
		chars += c
	}

	Printer.Infoln(chars + title + chars)
}

func (app *App) separatorLength() float64 {
	maxlength := 50.0
	for _, group := range app.config.Groups {
		length := ZhLen(group.GroupName)
		if length > maxlength {
			maxlength = length + 10
		}
	}

	return maxlength
}

// 加载
func (app *App) loadServerMap(check bool) {
	Log.Category("app").Info("server count", len(app.config.Servers), "group count", len(app.config.Groups))

	for i := range app.config.Servers {
		server := &app.config.Servers[i]
		server.Format()
		flag := strconv.Itoa(i + 1)

		if _, ok := app.serverIndex[flag]; ok && check {
			panic(errors.New("标识[" + flag + "]已存在，请检查您的配置文件"))
		}

		server.MergeOptions(app.config.Options, false)
		app.serverIndex[flag] = ServerIndex{
			indexType:   IndexTypeServer,
			groupIndex:  -1,
			serverIndex: i,
			server:      server,
		}
	}

	for i := range app.config.Groups {
		group := &app.config.Groups[i]
		for j := range group.Servers {
			server := &group.Servers[j]
			server.Format()
			flag := group.Prefix + strconv.Itoa(j+1)

			if _, ok := app.serverIndex[flag]; ok && check {
				panic(errors.New("标识[" + flag + "]已存在，请检查您的配置文件"))
			}

			server.MergeOptions(app.config.Options, false)
			app.serverIndex[flag] = ServerIndex{
				indexType:   IndexTypeGroup,
				groupIndex:  i,
				serverIndex: j,
				server:      server,
			}
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
