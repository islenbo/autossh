package main

import "autossh/core"

const SERVER_PATH = "./servers.json"

func main() {
	app := core.App{ServerPath: SERVER_PATH}
	app.Start()
}
