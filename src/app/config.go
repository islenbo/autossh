package app

import (
	"autossh/src/utils"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	ShowDetail bool                   `json:"show_detail"`
	Servers    []Server               `json:"servers"`
	Groups     []Group                `json:"groups"`
	Options    map[string]interface{} `json:"options"`

	// 服务器map索引，可通过编号、别名快速定位到某一个服务器
	serverIndex map[string]ServerIndex
	file        string
}

type Group struct {
	GroupName string   `json:"group_name"`
	Prefix    string   `json:"prefix"`
	Servers   []Server `json:"servers"`
	Collapse  bool     `json:"collapse"`
}

type LogMode string

const (
	LogModeCover  LogMode = "cover"
	LogModeAppend LogMode = "append"
)

type ServerLog struct {
	Enable   bool    `json:"enable"`
	Filename string  `json:"filename"`
	Mode     LogMode `json:"mode"`
}

const (
	IndexTypeServer IndexType = iota
	IndexTypeGroup
)

type IndexType int
type ServerIndex struct {
	indexType   IndexType
	groupIndex  int
	serverIndex int
	server      *Server
}

// 创建服务器索引
func (cfg *Config) createServerIndex() {
	cfg.serverIndex = make(map[string]ServerIndex)
	for i := range cfg.Servers {
		server := &cfg.Servers[i]
		server.Format()
		index := strconv.Itoa(i + 1)

		if _, ok := cfg.serverIndex[index]; ok {
			continue
		}

		server.MergeOptions(cfg.Options, false)
		cfg.serverIndex[index] = ServerIndex{
			indexType:   IndexTypeServer,
			groupIndex:  -1,
			serverIndex: i,
			server:      server,
		}
		if server.Alias != "" {
			cfg.serverIndex[server.Alias] = cfg.serverIndex[index]
		}
	}

	for i := range cfg.Groups {
		group := &cfg.Groups[i]
		for j := range group.Servers {
			server := &group.Servers[j]
			server.Format()
			server.groupName = group.GroupName
			index := group.Prefix + strconv.Itoa(j+1)

			if _, ok := cfg.serverIndex[index]; ok {
				continue
			}

			server.MergeOptions(cfg.Options, false)
			cfg.serverIndex[index] = ServerIndex{
				indexType:   IndexTypeGroup,
				groupIndex:  i,
				serverIndex: j,
				server:      server,
			}
			if server.Alias != "" {
				cfg.serverIndex[server.Alias] = cfg.serverIndex[index]
			}
		}
	}
}

// 保存配置文件
func (cfg *Config) saveConfig(backup bool) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	if backup {
		err = cfg.backup()
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(cfg.file, out.Bytes(), os.ModePerm)
}

// 备份配置文件
func (cfg *Config) backup() error {
	srcFile, err := os.Open(cfg.file)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	path, _ := filepath.Abs(filepath.Dir(cfg.file))
	backupFile := path + "/config-" + time.Now().Format("20060102150405") + ".json"
	desFile, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer desFile.Close()

	_, err = io.Copy(desFile, srcFile)
	if err != nil {
		return err
	}

	utils.Infoln("配置文件已备份：", backupFile)
	return nil
}
