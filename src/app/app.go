package app

import (
	"autossh/src/app/commands"
	"autossh/src/utils"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	IndexTypeServer IndexType = iota
	IndexTypeGroup
)

var (
	Version string
	Build   string

	varVersion = commands.VersionInfo{Version: Version, Build: Build, Value: false}
	varHelp    = commands.HelpInfo{Value: false}
	varConfig  = commands.Config{Value: "./config.json"}

	appConfig   Config
	serverIndex = make(map[string]ServerIndex)
)

func init() {
	flag.BoolVar(&varVersion.Value, "v", varVersion.Value, "显示版本信息。")
	flag.BoolVar(&varHelp.Value, "h", varVersion.Value, "显示帮助信息。")
	flag.StringVar(&varConfig.Value, "c", varConfig.Value, "指定配置文件。")
	flag.Parse()
}

// 启动
func Run() {
	if varVersion.Value {
		varVersion.Process()
		return
	}

	if varHelp.Value {
		varHelp.Process()
		return
	}

	varConfig.Process()

	_, err := os.Stat(varConfig.Value)
	if err != nil {
		if os.IsNotExist(err) {
			utils.Errorln("config file", varConfig.Value+" not exists")
		} else {
			utils.Errorln("unknown error", err)
		}
		return
	}

	if err = loadConfigAndShow(); err != nil {
		utils.Errorln("加载配置失败：" + err.Error())
		return
	}
}

// 加载配置文件
func loadConfig() error {
	b, _ := ioutil.ReadFile(varConfig.Value)
	err := json.Unmarshal(b, &appConfig)
	return err
}

// 加载
func loadServerMap(check bool) error {
	for i := range appConfig.Servers {
		server := &appConfig.Servers[i]
		server.Format()
		index := strconv.Itoa(i + 1)

		if _, ok := serverIndex[index]; ok && check {
			return errors.New("标识[" + index + "]已存在，请检查您的配置文件")
		}

		server.MergeOptions(appConfig.Options, false)
		serverIndex[index] = ServerIndex{
			indexType:   IndexTypeServer,
			groupIndex:  -1,
			serverIndex: i,
			server:      server,
		}
	}

	for i := range appConfig.Groups {
		group := &appConfig.Groups[i]
		for j := range group.Servers {
			server := &group.Servers[j]
			server.Format()
			index := group.Prefix + strconv.Itoa(j+1)

			if _, ok := serverIndex[index]; ok && check {
				return errors.New("标识[" + index + "]已存在，请检查您的配置文件")
			}

			server.MergeOptions(appConfig.Options, false)
			serverIndex[index] = ServerIndex{
				indexType:   IndexTypeGroup,
				groupIndex:  i,
				serverIndex: j,
				server:      server,
			}
		}
	}

	return nil
}

func show() (err error) {
	//for {
	if err = utils.Clear(); err != nil {
		return err
	}

	// 输出server
	showServers()

	// 监听输入
	input, isGlobal := checkInput()
	if isGlobal {
		end := handleGlobalCmd(input)
		if end {
			return nil
		}
	} else {
		server := serverIndex[input].server
		utils.Infoln("你选择了", server.Name)
		server.Connect()
	}

	return nil
}

func showServers() {
	maxlen := separatorLength()
	formatSeparator(" 欢迎使用 Auto SSH ", "=", maxlen)
	for i, server := range appConfig.Servers {
		utils.Logln(recordServer(strconv.Itoa(i+1), server))
	}

	for _, group := range appConfig.Groups {
		if len(group.Servers) == 0 {
			continue
		}

		formatSeparator(" "+group.GroupName+" ", "_", maxlen)
		for i, server := range group.Servers {
			utils.Logln(recordServer(group.Prefix+strconv.Itoa(i+1), server))
		}
	}

	formatSeparator("", "=", maxlen)
	utils.Logln("", "[add]  添加", "    ", "[edit] 编辑", "    ", "[remove] 删除")
	utils.Logln("", "[exit]\t退出")
	formatSeparator("", "=", maxlen)
	utils.Info("请输入序号或操作: ")
}

// 检查输入
func checkInput() (ipt string, isGlobal bool) {
	for {
		fmt.Scanln(&ipt)

		if isGlobalInput(ipt) {
			isGlobal = true
			break
		}

		if _, ok := serverIndex[ipt]; !ok {
			utils.Errorln("输入有误，请重新输入")
			continue
		} else {
			isGlobal = false
			break
		}
	}

	return ipt, isGlobal
}

func separatorLength() float64 {
	maxlength := 50.0
	for _, group := range appConfig.Groups {
		length := utils.ZhLen(group.GroupName)
		if length > maxlength {
			maxlength = length + 10
		}
	}

	return maxlength
}

// 判断是否全局输入
func isGlobalInput(flag string) bool {
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

func formatSeparator(title string, c string, maxlength float64) {

	charslen := int((maxlength - utils.ZhLen(title)) / 2)
	chars := ""
	for i := 0; i < charslen; i++ {
		chars += c
	}

	utils.Infoln(chars + title + chars)
}

func recordServer(flag string, server Server) string {
	if appConfig.ShowDetail {
		return " [" + flag + "]" + "\t" + server.Name + " [" + server.User + "@" + server.Ip + "]"
	} else {
		return " [" + flag + "]" + "\t" + server.Name
	}
}

// TODO 将全局处理抽取重构
func handleGlobalCmd(cmd string) bool {
	switch strings.ToLower(cmd) {
	case "exit":
		return true
	case "edit":
		handleEdit()
		return false
	case "add":
		handleAdd()
		return false
	case "remove":
		handleRemove()
		return false
	default:
		utils.Errorln("指令无效")
		return false
	}
}

// 编辑
func handleEdit() {
	utils.Info("请输入相应序号（exit退出当前操作）：")
	id := ""
	fmt.Scanln(&id)

	if strings.ToLower(id) == "exit" {
		show()
		return
	}

	serverIndex, ok := serverIndex[id]
	if !ok {
		utils.Errorln("序号不存在")
		handleEdit()
		return
	}

	serverIndex.server.Edit()
	saveAndReload()
}

// 新增
func handleAdd() {
	groups := make(map[string]*Group)
	for i := range appConfig.Groups {
		group := &appConfig.Groups[i]
		groups[group.Prefix] = group
		utils.Info("["+group.Prefix+"]"+group.GroupName, "\t")
	}
	utils.Infoln("[其他值]默认组")
	utils.Info("请输入要插入的组：")
	g := ""
	fmt.Scanln(&g)

	server := Server{}
	server.Format()
	server.Edit()

	group, ok := groups[g]
	if ok {
		group.Servers = append(group.Servers, server)
	} else {
		appConfig.Servers = append(appConfig.Servers, server)
	}

	saveAndReload()
}

// 移除
func handleRemove() {
	utils.Info("请输入相应序号（exit退出当前操作）：")
	id := ""
	fmt.Scanln(&id)

	if strings.ToLower(id) == "exit" {
		show()
		return
	}

	serverIndex, ok := serverIndex[id]
	if !ok {
		utils.Errorln("序号不存在")
		handleEdit()
		return
	}

	if serverIndex.indexType == IndexTypeServer {
		servers := appConfig.Servers
		appConfig.Servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
	} else {
		servers := appConfig.Groups[serverIndex.groupIndex].Servers
		servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
		appConfig.Groups[serverIndex.groupIndex].Servers = servers
	}

	saveAndReload()
}

func saveAndReload() (err error) {
	if err = saveConfig(); err != nil {
		return errors.New("保存配置文件失败：" + err.Error())
	}

	return loadConfigAndShow()
}

func loadConfigAndShow() (err error) {
	if err = loadConfig(); err != nil {
		return err
	}

	if err = loadServerMap(true); err != nil {
		return err
	}

	if err = show(); err != nil {
		return err
	}

	return nil
}

// 保存配置文件
func saveConfig() error {
	b, err := json.Marshal(appConfig)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	err = backConfig()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(varConfig.Value, out.Bytes(), os.ModePerm)
}

func backConfig() error {
	srcFile, err := os.Open(varConfig.Value)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	path, _ := filepath.Abs(filepath.Dir(varConfig.Value))
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

	utils.Infoln("配置文件已备份：", backupFile)
	return nil
}
