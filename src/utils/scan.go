package utils

import (
	"bufio"
	"os"
)

// GO自带的fmt.Scanln将空格也当作结束符，若需要读取含有空格的句子请使用该方法
func Scanln(a *string) {
	reader := bufio.NewReader(os.Stdin)
	data, _, _ := reader.ReadLine()
	*a = string(data)
}
