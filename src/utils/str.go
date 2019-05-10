package utils

import "unicode"

// 计算字符宽度（中文）
func ZhLen(str string) float64 {
	length := 0.0
	for _, c := range str {
		if unicode.Is(unicode.Scripts["Han"], c) {
			length += 2
		} else {
			length += 1
		}
	}

	return length
}
