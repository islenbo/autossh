package main

import (
	"autossh/core"
	"os"
	"path/filepath"
	"flag"
)

var (
	Version = "unknown"
	Build   = "unknown"
)

func main() {
	configPath := flag.String("c", "./config.json", "指定配置文件路径")
	flag.Parse()

	if *configPath == "" {
		*configPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		*configPath = *configPath + "/config.json"
	}

	_, err := os.Stat(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			core.Printer.Errorln("config file", *configPath+" not exists")
		} else {
			core.Printer.Errorln("unknown error", err)
		}

		return
	}

	app := core.App{
		ConfigPath: *configPath,
		Version:    Version,
		Build:      Build,
	}
	app.Exec()
}
