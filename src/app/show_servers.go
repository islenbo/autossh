package app

import (
	"autossh/src/utils"
	"strconv"
)

func showServers(configFile string) {
	// 清屏
	_ = utils.Clear()
	show(cfg)

	for {
		loop, clear, reload := scanInput(cfg)
		if !loop {
			break
		}

		if reload {
			cfg, _ = loadConfig(configFile)
		}

		if clear {
			_ = utils.Clear()
		}

		show(cfg)
	}
}

// 显示服务
func show(cfg *Config) {
	maxlen := separatorLength(*cfg)
	utils.Infoln(utils.FormatSeparator(" 欢迎使用 Auto SSH ", "=", maxlen))
	for i, server := range cfg.Servers {
		utils.Logln(server.FormatPrint(strconv.Itoa(i+1), cfg.ShowDetail))
	}

	for _, group := range cfg.Groups {
		if len(group.Servers) == 0 {
			continue
		}

		var collapseNotice = ""
		if group.Collapse {
			collapseNotice = "[" + group.Prefix + " ↓]"
		} else {
			collapseNotice = "[" + group.Prefix + " ↑]"
		}

		utils.Infoln(utils.FormatSeparator(" "+group.GroupName+" "+collapseNotice+" ", "_", maxlen))
		if !group.Collapse {
			for i, server := range group.Servers {
				utils.Logln(server.FormatPrint(group.Prefix+strconv.Itoa(i+1), cfg.ShowDetail))
			}
		}
	}

	utils.Infoln(utils.FormatSeparator("", "=", maxlen))

	showMenu()

	utils.Infoln(utils.FormatSeparator("", "=", maxlen))
	utils.Info("请输入序号或操作: ")
}

// 计算分隔符长度
func separatorLength(cfg Config) int {
	maxlength := 60
	for _, group := range cfg.Groups {
		length := utils.ZhLen(group.GroupName)
		if length > maxlength {
			maxlength = length + 10
		}
	}

	return maxlength
}
