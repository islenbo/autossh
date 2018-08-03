package main

import (
	"autossh/core"
	"os"
	"path/filepath"
	"fmt"
	"strings"
)

var (
	Version = "unknown"
	Build   = "unknown"
)

func main() {
	configPath := ""
	if len(os.Args) > 1 {
		option := strings.Split(os.Args[1], "=")

		switch option[0] {
		case "--config":
			configPath = *core.Params.Get("--config")
		case "-c":
			configPath = *core.Params.Get("-c")

		case "--help":
			fallthrough
		case "-h":
			help()
			return

		case "--version":
			fallthrough
		case "-v":
			version()
			return
		}
	}

	defer func() {
		if err := recover(); err != nil {
			core.Log.Category("main").Error("recover", err)
		}
	}()

	if configPath == "" {
		configPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		configPath = configPath + "/config.json"
	} else {
		configPath, _ = core.ParsePath(configPath)
	}

	core.Log.Category("main").Info("config path=", configPath)

	_, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			core.Printer.Errorln("config file", configPath+" not exists")
			core.Log.Category("main").Error("config file not exists")
		} else {
			core.Printer.Errorln("unknown error", err)
			core.Log.Category("main").Error("unknown error", err)
		}

		return
	}

	app := core.App{
		ConfigPath: configPath,
	}
	app.Init()
}

// 版本信息
func version() {
	fmt.Println("autossh " + Version + " Build " + Build + "。")
	fmt.Println("由 Lenbo 编写，项目地址：https://github.com/islenbo/autossh。")
}

// 显示帮助信息
func help() {
	fmt.Println("一个ssh远程客户端，可一键登录远程服务器，主要用来弥补Mac/Linux Terminal ssh无法保存密码的不足。")
	fmt.Println("参数：")
	fmt.Println("  -c, --config ", "default=./config.json \t", "指定配置文件。")
	fmt.Println("  -h, --help   ", "                      \t", "显示帮助信息。")
	fmt.Println("  -v, --version", "                      \t", "显示 autossh 的版本信息。")
}
