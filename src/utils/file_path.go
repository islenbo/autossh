package utils

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// 解析路径
func ParsePath(path string) (string, error) {
	str := []rune(path)
	firstKey := string(str[:1])

	if firstKey == "~" {
		home, err := home()
		if err != nil {
			return "", err
		}

		return home + string(str[1:]), nil
	} else if firstKey == "/" {
		return path, nil
	} else {
		p, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return p + "/" + path, nil
	}
}

func home() (string, error) {
	u, err := user.Current()
	if nil == err {
		return u.HomeDir, nil
	}

	// cross compile support

	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

// 判断文件是否存在
func FileIsExists(file string) (bool, error) {
	file, err := ParsePath(file)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(file)
	if err != nil {
		return false, err
		//if os.IsNotExist(err) {
		//	return false, err
		//} else {
		//	// unknown error
		//}
	}

	return true, nil
}
