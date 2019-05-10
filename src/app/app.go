package app

import (
	"autossh/src/app/commands"
	"autossh/src/utils"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	IndexTypeServer IndexType = iota
	IndexTypeGroup
)

const (
	InputCmdOpt int = iota
	InputCmdServer
	InputCmdGroupPrefix
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
	if err = utils.Clear(); err != nil {
		return err
	}

	// 输出server
	showServers()

	// 监听输入
	input, inputCmd, groupIndex := checkInput()
	switch inputCmd {
	case InputCmdOpt:
		{
			operation := operations[input]
			operation.Process()
			if operation.End {
				return nil
			}
		}
	case InputCmdServer:
		{
			server := serverIndex[input].server
			utils.Infoln("你选择了", server.Name)
			server.Connect()
		}
	case InputCmdGroupPrefix:
		{
			group := &appConfig.Groups[groupIndex]
			group.Collapse = !group.Collapse
			_ = saveConfig(false)
			show()
		}
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

		var collapseNotice = ""
		if group.Collapse {
			collapseNotice = "[" + group.Prefix + " ↓]"
		} else {
			collapseNotice = "[" + group.Prefix + " ↑]"
		}

		formatSeparator(" "+group.GroupName+" "+collapseNotice+" ", "_", maxlen)
		if !group.Collapse {
			for i, server := range group.Servers {
				utils.Logln(recordServer(group.Prefix+strconv.Itoa(i+1), server))
			}
		}
	}

	formatSeparator("", "=", maxlen)

	showMenu()

	formatSeparator("", "=", maxlen)
	utils.Info("请输入序号或操作: ")
}

// 检查输入
func checkInput() (ipt string, inputCmd int, groupIndex int) {
	for {
		fmt.Scanln(&ipt)

		if _, exists := operations[ipt]; exists {
			inputCmd = InputCmdOpt
			break
		}

		if _, ok := serverIndex[ipt]; ok {
			inputCmd = InputCmdServer
			break
		}

		groupIndex = -1
		for index, group := range appConfig.Groups {
			if group.Prefix == ipt {
				inputCmd = InputCmdGroupPrefix
				groupIndex = index
				break
			}
		}
		if groupIndex != -1 {
			break
		}

		utils.Errorln("输入有误，请重新输入")
	}

	return ipt, inputCmd, groupIndex
}

func separatorLength() float64 {
	maxlength := 60.0
	for _, group := range appConfig.Groups {
		length := utils.ZhLen(group.GroupName)
		if length > maxlength {
			maxlength = length + 10
		}
	}

	return maxlength
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
