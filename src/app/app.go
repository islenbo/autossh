package app

import (
	"autossh/src/utils"
	"flag"
	"os"
	"path/filepath"
)

var (
	Version string
	Build   string

	c       string
	v       bool
	h       bool
	upgrade bool
	cp      bool
	edit    bool

	cfg *Config
)

func init() {
	// 取执行文件所在目录下的config.json
	dir, _ := os.Executable()
	c = filepath.Dir(dir) + "/config.json"

	flag.StringVar(&c, "c", c, "指定配置文件路径")
	flag.StringVar(&c, "config", c, "指定配置文件路径")

	flag.BoolVar(&v, "v", v, "版本信息")
	flag.BoolVar(&v, "version", v, "版本信息")

	flag.BoolVar(&h, "h", h, "帮助信息")
	flag.BoolVar(&h, "help", h, "帮助信息")

	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) > 0 {
		arg := flag.Arg(0)
		switch arg {
		case "upgrade":
			upgrade = true
		case "cp":
			cp = true
		case "edit":
			edit = true
		default:
			defaultServer = arg
		}
	}
}

func Run() {
	if v {
		showVersion()
	} else if h {
		showHelp()
	} else if upgrade {
		showUpgrade()
	} else {
		var err error
		cfg, err = loadConfig(c)
		if err != nil {
			utils.Errorln(err)
			return
		}

		if cp {
			showCp()
		} else if edit {
			showEdit()
		} else {
			showServers(c)
		}
	}
}
