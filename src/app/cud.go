package app

import (
	"autossh/src/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 编辑
func handleEdit(...interface{}) {
	utils.Info("请输入相应序号（exit退出当前操作）：")
	id := ""
	fmt.Scanln(&id)

	if strings.ToLower(id) == "exit" {
		return
	}

	serverIndex, ok := serverIndex[id]
	if !ok {
		utils.Errorln("序号不存在")
		handleEdit()
		return
	}

	serverIndex.server.Edit()
	_ = saveAndReload()
}

// 新增
func handleAdd(...interface{}) {
	groups := make(map[string]*Group)
	for i := range appConfig.Groups {
		group := &appConfig.Groups[i]
		groups[group.Prefix] = group
		utils.Info("["+group.Prefix+"]"+group.GroupName, "\t")
	}
	utils.Infoln("[其他值]默认组")
	utils.Info("请输入要插入的组：")
	g := ""
	fmt.Scanln(&g)

	server := Server{}
	server.Format()
	server.Edit()

	group, ok := groups[g]
	if ok {
		group.Servers = append(group.Servers, server)
	} else {
		appConfig.Servers = append(appConfig.Servers, server)
	}

	_ = saveAndReload()
}

// 移除
func handleRemove(...interface{}) {
	utils.Info("请输入相应序号（exit退出当前操作）：")
	id := ""
	fmt.Scanln(&id)

	if strings.ToLower(id) == "exit" {
		return
	}

	serverIndex, ok := serverIndex[id]
	if !ok {
		utils.Errorln("序号不存在")
		handleRemove()
		return
	}

	if serverIndex.indexType == IndexTypeServer {
		servers := appConfig.Servers
		appConfig.Servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
	} else {
		servers := appConfig.Groups[serverIndex.groupIndex].Servers
		servers = append(servers[:serverIndex.serverIndex], servers[serverIndex.serverIndex+1:]...)
		appConfig.Groups[serverIndex.groupIndex].Servers = servers
	}

	_ = saveAndReload()
}

func saveAndReload() (err error) {
	if err = saveConfig(true); err != nil {
		return errors.New("保存配置文件失败：" + err.Error())
	}

	return loadConfigAndShow()
}

// 保存配置文件
func saveConfig(backup bool) error {
	b, err := json.Marshal(appConfig)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	if backup {
		err = backConfig()
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(varConfig.Value, out.Bytes(), os.ModePerm)
}

func backConfig() error {
	srcFile, err := os.Open(varConfig.Value)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	path, _ := filepath.Abs(filepath.Dir(varConfig.Value))
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
