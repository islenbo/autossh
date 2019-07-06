package utils

import (
	"os"
	"os/exec"
	"runtime"
)

// 清屏
func Clear() error {
	var cmd exec.Cmd
	if "windows" == runtime.GOOS {
		cmd = *exec.Command("cmd", "/c", "cls")
	} else {
		cmd = *exec.Command("clear")
	}

	cmd.Stdout = os.Stdout

	return cmd.Run()
}
