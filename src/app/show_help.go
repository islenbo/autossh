package app

import (
	"autossh/src/utils"
	"flag"
)

func showHelp() {
	flag.Usage = usage
	flag.Usage()
}

func usage() {
	str :=
		`一个ssh远程客户端，可一键登录远程服务器，主要用来弥补Mac/Linux Terminal ssh无法保存密码的不足。
Usage:
  autossh [options] [commands]

Options:
  -c, -config	指定配置文件。
             	(default: ./config.json)
  -v, -version	显示版本信息。
  -h, -help	显示帮助信息。

Commands:
  upgrade    		检测升级。
  ${ServerNum}  	使用编号登录指定服务器。
  ${ServerAlias} 	使用别名登录指定服务器。
`
	utils.Logln(str)
}
