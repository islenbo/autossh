package app

import (
	"autossh/src/utils"
	"fmt"
	"io"
)

func handleEdit(cfg *Config, args []string) error {
	utils.Info("请输入相应序号：")
	id := ""
	if _, err := fmt.Scanln(&id); err == io.EOF {
		return nil
	}

	serverIndex, ok := cfg.serverIndex[id]
	if !ok {
		utils.Errorln("序号不存在")
		return handleEdit(cfg, args)
	}

	if err := serverIndex.server.Edit(); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return cfg.saveConfig(true)
}
