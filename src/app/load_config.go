package app

import (
	"autossh/src/utils"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
)

// 加载配置
func loadConfig(configFile string) (cfg *Config, err error) {
	configFile, err = utils.ParsePath(configFile)
	if err != nil {
		return cfg, err
	}

	if exists, _ := utils.FileIsExists(configFile); !exists {
		return cfg, errors.New("Can't read configFile file:" + configFile)
	}

	b, _ := ioutil.ReadFile(configFile)
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return cfg, err
	}

	cfg.file = configFile
	
	//加载外部文件
	for i := range cfg.ExtServerConfigs {
		b, err = ioutil.ReadFile(cfg.ExtServerConfigs[i])
		if err != nil {
			return cfg, errors.New("Can't read ext server config file: " + cfg.ExtServerConfigs[i] + ", err: " + err.Error())
		}
		server := new([]Server)
		err = json.Unmarshal(b, &server)
		if err != nil {
			return cfg, errors.New("Can't parse ext server config file: " + cfg.ExtServerConfigs[i] + ", err: " + err.Error())
		}

		for j := range *server {
			cfg.Servers = append(cfg.Servers, &((*server)[j]))
		}
	}

	cfg.createServerIndex()

	return cfg, nil
}
