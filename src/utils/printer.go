package utils

import "fmt"

// 打印一行信息
// 字体颜色为默色
func Logln(a ...interface{}) {
	fmt.Println(a...)
}

// 打印一行信息
// 字体颜色为绿色
func Infoln(a ...interface{}) {
	fmt.Print("\033[32m")
	Logln(a...)
	fmt.Print("\033[0m")
}

// 打印信息（不换行）
// 字体颜色为绿色
func Info(a ...interface{}) {
	fmt.Print("\033[32m")
	Logln(a...)
	fmt.Print("\033[0m")
}

// 打印一行错误
// 字体颜色为红色
func Errorln(a ...interface{}) {
	fmt.Print("\033[31m")
	Logln(a...)
	fmt.Print("\033[0m")
}
