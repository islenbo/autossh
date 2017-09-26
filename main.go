package main

import (
	"autossh/core"
	"path/filepath"
	"os"
	"fmt"
)

var (
	Version = "unknown"
	Build = "unknown"
)

func main() {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	app := core.App{
		ServersPath: path + "/servers.json",
		Version:     Version,
		Build: Build,
	}
	app.Exec()
}
