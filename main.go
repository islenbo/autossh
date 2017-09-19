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

	app := core.App{ServerPath: path + "/servers.json"}
	app.Start()
}
