package main

import (
	"autossh/core"
	"path/filepath"
	"os"
	"fmt"
)

func main() {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	path = "/Users/linbo/develop/golang/src/autossh/"

	app := core.App{ServersPath: path + "/servers.json"}
	app.Exec()
}
