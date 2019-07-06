package app

import (
	"autossh/src/utils"
	"fmt"
	"io"
)

func handleRemove(cfg *Config, args []string) error {
	utils.Info("请输入相应序号：")

	id := ""
	_, err := fmt.Scanln(&id)
	if err == io.EOF {
		return nil
	}

	serverIndex, ok := cfg.serverIndex[id]
	if !ok {
		utils.Errorln("序号不存在")
		return handleRemove(cfg, args)
	}

	if serverIndex.indexType == IndexTypeServer {
		servers := cfg.Servers
		cfg.Servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
	} else {
		servers := cfg.Groups[serverIndex.groupIndex].Servers
		servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
		cfg.Groups[serverIndex.groupIndex].Servers = servers
	}

	return cfg.saveConfig(true)
}
