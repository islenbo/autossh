package utils

import (
	"log"
	"os"
	"path/filepath"
)

type logger struct {
	File     string
	category string
	level    string
}

var Logger logger

func init() {
	dir, _ := os.Executable()
	logFile, _ := ParsePath(filepath.Dir(dir) + "/app.log")
	Logger = logger{
		File: logFile,
	}
}

func (logger *logger) write(msg ...interface{}) {
	if _, err := os.Stat(logger.File); err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(logger.File)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	logFile, err := os.OpenFile(logger.File, os.O_RDWR|os.O_APPEND, 0666)
	defer logFile.Close()
	if err != nil {
		panic(err)
	}

	// 创建一个日志对象
	l := log.New(logFile, logger.level, log.LstdFlags)

	s := make([]interface{}, 1)
	s[0] = "[" + logger.category + "]"
	msg = append(s, msg)

	l.Println(msg...)
	logger.category = ""
}

func (logger *logger) Category(category string) *logger {
	logger.category = category
	return logger
}

func (logger *logger) Info(msg ...interface{}) {
	logger.level = "[Info]"
	logger.write(msg...)
}

func (logger *logger) Error(msg ...interface{}) {
	logger.level = "[Error]"
	logger.write(msg...)
}
