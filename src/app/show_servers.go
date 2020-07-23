package app

import (
	"autossh/src/utils"
	"strconv"
)

func showServers(configFile string) {
	cfg, err := loadConfig(configFile)
	if err != nil {
		utils.Errorln(err)
		return
	}

	// 清屏
	_ = utils.Clear()

	show(cfg)

	for {
		loop, clear, reload := scanInput(cfg)
		if !loop {
			break
		}

		if reload {
			cfg, err = loadConfig(configFile)
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

	indexMaxLen := 0;
	for i, server := range cfg.Servers {
		ilen := len(" [" + strconv.Itoa(i+1) + server.Alias + "]");
		if ilen > indexMaxLen {
			indexMaxLen = ilen;
		}
	}

	for i, server := range cfg.Servers {
		utils.Logln(server.FormatPrint(strconv.Itoa(i+1), indexMaxLen, cfg.ShowDetail))
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

		utils.Infoln("");
		utils.Infoln(utils.FormatSeparator(" "+group.GroupName+" "+collapseNotice+" ", "_", maxlen))
		if !group.Collapse {
			indexMaxLen := 0;
			for i, server := range group.Servers {
				ilen := len(" [" + group.Prefix + strconv.Itoa(i+1) + server.Alias + "]");
				if ilen > indexMaxLen {
					indexMaxLen = ilen;
				}
			}
			for i, server := range group.Servers {
				utils.Logln(server.FormatPrint(group.Prefix + strconv.Itoa(i+1), indexMaxLen, cfg.ShowDetail))
			}
		}
	}

	utils.Infoln("");
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
