package app

import (
	"autossh/src/utils"
	"strings"
)

type Operation struct {
	Key     string
	Label   string
	End     bool
	Process func(cfg *Config, args []string) error
}

var menuMap [][]Operation

var operations = make(map[string]Operation)

func init() {
	menuMap = [][]Operation{
		{
			{Key: "add", Label: "添加", Process: handleAdd},
			{Key: "edit", Label: "编辑", Process: handleEdit},
			{Key: "remove", Label: "删除", Process: handleRemove},
		},
		{
			{Key: "exit", Label: "退出", End: true},
		},
	}
}

func showMenu() {
	var columnsMaxWidths = make(map[int]int)

	for i := 0; i < len(menuMap); i++ {
		for j := 0; j < len(menuMap[i]); j++ {
			operation := menuMap[i][j]

			// 计算每列最大长度
			maxLen := int(utils.ZhLen(operationFormat(operation)))
			if _, exists := columnsMaxWidths[j]; !exists {
				columnsMaxWidths[j] = maxLen
			}
			if columnsMaxWidths[j] < maxLen {
				columnsMaxWidths[j] = maxLen
			}

			operations[operation.Key] = operation
		}
	}

	for i := 0; i < len(menuMap); i++ {
		var output = ""
		for j := 0; j < len(menuMap[i]); j++ {
			operation := menuMap[i][j]
			output += stringPadding(operationFormat(operation), columnsMaxWidths[j]) + "\t"
		}

		utils.Logln(strings.TrimSpace(output))
		output = ""
	}
}

func operationFormat(operation Operation) string {
	return "[" + operation.Key + "] " + operation.Label
}

func stringPadding(str string, paddingLen int) string {
	if len(str) < paddingLen {
		return stringPadding(str+" ", paddingLen)
	} else {
		return str
	}
}
