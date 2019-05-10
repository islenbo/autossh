package commands

import (
	"autossh/src/utils"
	"flag"
)

type HelpInfo struct {
	Value bool
}

func (helInfo *HelpInfo) Process() {
	flag.Usage = usage
	flag.Usage()
}

func usage() {
	utils.Logln("一个ssh远程客户端，可一键登录远程服务器，主要用来弥补Mac/Linux Terminal ssh无法保存密码的不足。")
	utils.Logln("参数：")
	flag.PrintDefaults()
}
