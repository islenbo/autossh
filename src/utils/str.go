package utils

import (
	"unicode"
)

// 计算字符宽度（中文）
func ZhLen(str string) int {
	length := 0
	for _, c := range str {
		if unicode.Is(unicode.Scripts["Han"], c) {
			length += 2
		} else {
			length += 1
		}
	}

	return length
}

// 左右填充
// title 主体内容
// c 填充符号
// maxlength 总长度
// 如： title = 测试 c=* maxlength = 10 返回 ** 返回 **
func FormatSeparator(title string, c string, maxlength int) string {
	charslen := (maxlength - ZhLen(title)) / 2
	chars := ""
	for i := 0; i < charslen; i++ {
		chars += c
	}

	return chars + title + chars
}

// 右填充
//func AppendRight(body string, char string, maxlength int) string {
//	length := ZhLen(body)
//	if length >= maxlength {
//		return body
//	}
//
//	for i := 0; i < maxlength-length; i++ {
//		body = body + char
//	}
//
//	return body
//}
