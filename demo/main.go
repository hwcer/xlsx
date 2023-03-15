package main

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/xlsx"
)

func init() {
	cosgo.Console.Close()
}

func main() {
	cosgo.SetLoggerFormat(loggerFormat)
	cosgo.Start(false, xlsx.New())
}

func loggerFormat(msg *logger.Message) string {
	return msg.Content
}
