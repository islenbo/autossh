package main

import (
	"autossh/src/app"
)

var (
	Version = "unknown"
	Build   = "unknown"
)

func main() {
	app.Version = Version
	app.Build = Build
	app.Run()
}
