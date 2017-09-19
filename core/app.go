package core

import (
	"fmt"
	"strconv"
	"io/ioutil"
	"encoding/json"
)

type App struct {
	ServerPath string
}

var (
	servers []Server
)

// 启动脚本
func (app *App) Start() {
	b, err := ioutil.ReadFile(app.ServerPath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, &servers)
	if err != nil {
		panic(err)
	}

	Printer.Infoln("========== 欢迎使用 Auto SSH ==========")
	for i, server := range servers {
		Printer.Logln(" ["+strconv.Itoa(i+1)+"]", server.Name)
	}
	Printer.Infoln("=======================================")

	server := inputSh()
	Printer.Infoln("你选择了: " + server.Name)
	server.Connection()
}

// 接收输入，获取对应sh脚本
func inputSh() Server {
	Printer.Info("请输入序号: ")
	input := ""
	fmt.Scanln(&input)
	num, err := strconv.Atoi(input)
	if err != nil {
		Printer.Errorln("输入有误，请重新输入")
		return inputSh()
	}

	if num <= 0 || num > len(servers) {
		Printer.Errorln("输入有误，请重新输入")
		return inputSh()
	}

	return servers[num-1]
}
