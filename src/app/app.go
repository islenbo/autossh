package app

import (
	"autossh/src/utils"
	"flag"
)

var (
	Version string
	Build   string

	varVersion bool
	varHelp    bool
	varUpgrade bool
	varCp      bool
	varConfig  = "./config.json"
)

func init() {
	flag.BoolVar(&varVersion, "v", varVersion, "版本信息")
	flag.BoolVar(&varVersion, "version", varVersion, "版本信息")
	flag.BoolVar(&varHelp, "h", varHelp, "帮助信息")
	flag.BoolVar(&varHelp, "help", varHelp, "帮助信息")
	flag.StringVar(&varConfig, "c", varConfig, "指定配置文件路径")
	flag.StringVar(&varConfig, "config", varConfig, "指定配置文件路径")

	flag.Parse()

	if len(flag.Args()) > 0 {
		arg := flag.Arg(0)
		switch arg {
		case "upgrade":
			varUpgrade = true
		case "cp":
			varCp = true
		}
	}
}

func Run() {
	if exists, _ := utils.FileIsExists(varConfig); !exists {
		utils.Errorln("Can't read config file", varConfig)
		return
	}

	if varVersion {
		showVersion()
	} else if varHelp {
		showHelp()
	} else if varUpgrade {
		showUpgrade()
	} else if varCp {
		showCp(varConfig)
	} else {
		showServers(varConfig)
	}
}
