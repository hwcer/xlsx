package main

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
	"github.com/hwcer/xlsx"
)

func init() {
	logger.SetCallDepth(4)
	logger.Console.Sprintf = func(message *logger.Message) string {
		fmt.Printf("日志路径:%v\n", message.Path)
		return message.Content
	}
}

func main() {
	cosgo.Start(false, xlsx.New())
}
