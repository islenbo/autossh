package app

import (
	"autossh/src/utils"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// 编辑
func (server *Server) Edit() error {
	keys := []string{"Name", "Ip", "Port", "User", "Password", "Method", "Key", "Alias"}
	for _, key := range keys {
		if err := server.scanVal(key); err != nil {
			return err
		}
	}

	return nil
}

func deftVal(val string) string {
	if val != "" {
		return "(default=" + val + ")"
	} else {
		return ""
	}
}

func (server *Server) scanVal(fieldName string) (err error) {
	elem := reflect.ValueOf(server).Elem()
	field := elem.FieldByName(fieldName)
	switch field.Type().String() {
	case "int":
		utils.Info(fieldName + deftVal(strconv.FormatInt(field.Int(), 10)) + ":")
		var ipt int
		if _, err = fmt.Scanln(&ipt); err == nil {
			field.SetInt(int64(ipt))
		}
	case "string":
		utils.Info(fieldName + deftVal(field.String()) + ":")
		var ipt string
		if _, err = fmt.Scanln(&ipt); err == nil {
			field.SetString(ipt)
		}
	}

	if err != nil {
		if err == io.EOF {
			return err
		}

		// 允许输入空行
		if err.Error() == "unexpected newline" {
			return nil
		}
	}

	return nil
}
