package core

import (
	"runtime"
	"os"
	"bytes"
	"os/exec"
	"strings"
	"os/user"
	"errors"
	"path/filepath"
	"unicode"
)

// 计算字符宽度（中文）
func ZhLen(str string) float64 {
	length := 0.0
	for _, c := range str {
		if unicode.Is(unicode.Scripts["Han"], c) {
			length += 2
		} else {
			length += 1
		}
	}

	return length
}

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
	} else if firstKey == "." {
		p, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return p + "/" + path, nil
	} else {
		return path, nil
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
