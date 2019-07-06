package app

import (
	"autossh/src/utils"
	"strings"
)

const (
	InputCmdOpt int = iota
	InputCmdServer
	InputCmdGroupPrefix
)

var defaultServer = ""

// 获取输入
func scanInput(cfg *Config) (loop bool, clear bool, reload bool) {
	cmd, inputCmd, extInfo := checkInput(cfg)
	switch inputCmd {
	case InputCmdOpt:
		{
			operation := operations[cmd]
			if operation.Process != nil {
				if err := operation.Process(cfg, extInfo.([]string)); err != nil {
					utils.Errorln(err)
					loop = false
					return
				}

				if !operation.End {
					return true, true, true
				}
			}
			return
		}
	case InputCmdServer:
		{
			server := cfg.serverIndex[cmd].server
			utils.Infoln("你选择了", server.Name)
			err := server.Connect()
			if err != nil {
				utils.Logger.Error("server connect error ", err)
				utils.Errorln(err)
			}
			return true, true, false
		}
	case InputCmdGroupPrefix:
		{
			group := cfg.Groups[extInfo.(int)]
			group.Collapse = !group.Collapse
			err := cfg.saveConfig(false)
			if err != nil {
				utils.Errorln("备份失败", err)
				loop = false
				return
			} else {
				return true, true, true
			}
		}
	}

	loop = true
	return
}

// 检查输入
func checkInput(cfg *Config) (cmd string, inputCmd int, extInfo interface{}) {
	for {
		ipt := ""
		skipOpt := false
		if defaultServer == "" {
			utils.Scanln(&ipt)
		} else {
			ipt = defaultServer
			defaultServer = ""
			skipOpt = true
		}

		ipts := strings.Split(ipt, " ")
		cmd = ipts[0]

		if !skipOpt {
			if _, exists := operations[cmd]; exists {
				inputCmd = InputCmdOpt
				extInfo = ipts[1:]
				break
			}
		}

		if _, ok := cfg.serverIndex[cmd]; ok {
			inputCmd = InputCmdServer
			break
		}

		groupIndex := -1
		for index, group := range cfg.Groups {
			if group.Prefix == cmd {
				inputCmd = InputCmdGroupPrefix
				groupIndex = index
				extInfo = index
				break
			}
		}
		if groupIndex != -1 {
			break
		}

		utils.Errorln("输入有误，请重新输入")
	}

	return cmd, inputCmd, extInfo
}
