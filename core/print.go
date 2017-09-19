package core

import "fmt"

type Print struct {
}

var Printer Print

// 打印一行信息
// 字体颜色为绿色
func (print Print) Logln(a ...interface{}) {
	fmt.Println(a...)
}

// 打印信息（不换行）
// 字体颜色为绿色
func (print Print) Log(a ...interface{}) {
	fmt.Print(a...)
}

// 打印一行信息
// 字体颜色为绿色
func (print Print) Infoln(a ...interface{}) {
	fmt.Print("\033[32m")
	fmt.Println(a...)
	fmt.Print("\033[0m")
}

// 打印信息（不换行）
// 字体颜色为绿色
func (print Print) Info(a ...interface{}) {
	fmt.Print("\033[32m")
	fmt.Print(a...)
	fmt.Print("\033[0m")
}

// 打印一行错误
// 字体颜色为红色
func (print Print) Errorln(a ...interface{}) {
	fmt.Print("\033[31m")
	fmt.Println(a...)
	fmt.Print("\033[0m")
}

// 打印信息（不换行）
// 字体颜色为红色
func (print Print) Error(a ...interface{}) {
	fmt.Print("\033[31m")
	fmt.Print(a...)
	fmt.Print("\033[0m")
}
