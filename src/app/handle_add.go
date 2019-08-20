package app

import (
	"autossh/src/utils"
	"fmt"
	"io"
)

func handleAdd(cfg *Config, _ []string) error {
	groups := make(map[string]*Group)
	for i := range cfg.Groups {
		group := cfg.Groups[i]
		groups[group.Prefix] = group
		utils.Info("["+group.Prefix+"]"+group.GroupName, "\t")
	}
	utils.Infoln("[其他值]默认组")
	utils.Info("请输入要插入的组：")
	g := ""
	if _, err := fmt.Scanln(&g); err == io.EOF {
		return nil
	}

	server := Server{}
	server.Format()
	if err := server.Edit(); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	group, ok := groups[g]
	if ok {
		group.Servers = append(group.Servers, server)
		server.groupName = group.GroupName
	} else {
		cfg.Servers = append(cfg.Servers, &server)
	}

	return cfg.saveConfig(true)
}
