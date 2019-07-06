package utils

import "strings"

// 错误断言
func ErrorAssert(err error, assert string) bool {
	return strings.Contains(err.Error(), assert)
}
