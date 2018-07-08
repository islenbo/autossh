package core

import "os"

var paramsMap map[string]string

var Params params

type params struct {
}

func init() {
	paramsMap = make(map[string]string)
}

func (p params) Get(key string) *string {
	if v, ok := paramsMap[key]; ok {
		return &v
	}

	for _, param := range os.Args {
		if param[:len(key)] == key {
			val := param[len(key)+1:]
			paramsMap[key] = val
			break
		}
	}

	v, ok := paramsMap[key]
	if !ok {
		return nil
	} else {
		return &v
	}
}
