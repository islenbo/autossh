package app

import (
	"autossh/src/utils"
	"fmt"
	"os"
	"strings"
)

func handleCp(params ...interface{}) {
	var args []string
	if len(params) == 1 {
		args = params[0].([]string)
	}
	r, direction, server, serverPath, localPath := parseArgs(args)

	if &server == nil || serverPath == "" || localPath == "" {
		utils.Errorln("命令有误，请重新输入")
		scanInput()
		return
	}

	if direction == "up" {
		copyUpload(server, serverPath, localPath)
	} else {
		copyDownload(server, serverPath, localPath)
	}

	fmt.Println(r, direction, server, serverPath, localPath)
}

func copyUpload(server Server, serverPath string, localPath string) error {
	client, err := server.GetSftpClient()
	if err != nil {
		return err
	}

	// TODO 上传
	if client == nil {

	}

	return nil
}

func copyDownload(server Server, serverPath string, localPath string) {

}

// 解析参数
func parseArgs(args []string) (r bool, direction string, server Server, serverPath string, localPath string) {
	for _, arg := range args {
		if isR(arg) {
			r = true
			continue
		}

		if s, sp, ok := isServer(arg); ok {
			if direction == "" {
				direction = "down"
			}

			server = *s
			serverPath = sp
			continue
		}

		if lp, ok := isLocalPath(arg); ok {
			if direction == "" {
				direction = "up"
			}

			localPath = lp
			continue
		}
	}

	return r, direction, server, serverPath, localPath
}

// 是否为一个合法的本地文件/路径
func isLocalPath(arg string) (path string, ok bool) {
	ok, err := utils.FileIsExists(arg)
	// 如果文件不存在，也认为是合法的，因为可能是下载
	if os.IsNotExist(err) {
		ok = true
	}

	return arg, ok
}

// 是否是一个合法的server
// 支持的格式包括：$serverNum:$serverPath、$serverNum
func isServer(arg string) (server *Server, path string, ok bool) {
	args := strings.Split(arg, ":")

	if len(args) > 0 {
		if len(args) > 1 {
			path = args[1]
		}

		serIdx, exists := serverIndex[args[0]]
		if exists {
			server = serIdx.server
		}

		ok = server != nil && path != ""
	}

	return server, path, ok
}

func isR(arg string) bool {
	return arg == "-r"
}
