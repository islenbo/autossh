package utils

import (
	"math"
	"strconv"
)

func SizeFormat(size float64) string {
	var k = 1024 // or 1024
	var sizes = []string{"B", "KB", "MB", "GB", "TB"}
	if size == 0 {
		return "0 B"
	}
	i := math.Floor(math.Log(size) / math.Log(float64(k)))
	r := size / math.Pow(float64(k), i)
	return strconv.FormatFloat(r, 'f', 2, 64) + " " + sizes[int(i)]
}
