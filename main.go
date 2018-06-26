package main

import (
	"github.com/islenbo/autossh/core"
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

	app := core.App{ServersPath: path + "/servers.json"}
	app.Exec()
}
