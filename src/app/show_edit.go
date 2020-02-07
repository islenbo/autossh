package app

import (
	"autossh/src/utils"
	"encoding/json"
	"net/http"
)

type ConfigController struct {
}

func (ctl *ConfigController) restful(w http.ResponseWriter, r *http.Request) {
	var apiResp = func(data interface{}, code int, msg string) {
		var resp = make(map[string]interface{})
		resp["data"] = data
		resp["code"] = code
		resp["msg"] = msg

		buffer, _ := json.Marshal(resp)

		_, _ = w.Write(buffer)
	}

	defer func() {
		if err := recover(); err != nil {
			apiResp(nil, -1, err.(string))
		}
	}()

	switch r.Method {
	case "GET":
		apiResp(ctl.show(r), 0, "success")
	case "POST":
		apiResp(ctl.store(r), 0, "success")
	}
}

func (*ConfigController) show(r *http.Request) interface{} {
	return cfg
}

func (ctl *ConfigController) store(r *http.Request) interface{} {
	return nil
}

func showEdit() {
	var err error
	utils.Infoln("http edit server startup.")
	utils.Infoln("listening", cfg.HttpAddr)

	cfgCtl := ConfigController{}
	http.HandleFunc("/config", cfgCtl.restful)

	fsh := http.FileServer(http.Dir(cfg.HttpPublic))
	http.Handle("/", http.StripPrefix("/", fsh))

	//h := http.FileServer(http.Dir(cfg.HttpPublic))
	err = http.ListenAndServe(cfg.HttpAddr, nil)
	if err != nil {
		utils.Errorln("listen http server fail: ", err)
	}

}
