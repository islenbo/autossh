package commands

import (
	"autossh/src/utils"
)

type VersionInfo struct {
	Version string
	Build   string
	Value   bool
}

func (versionInfo *VersionInfo) Process() bool {
	utils.Logln("autossh " + versionInfo.Version + " Build " + versionInfo.Build + "。")
	utils.Logln("由 Lenbo 编写，项目地址：https://github.com/islenbo/autossh。")

	return true
}
