package main

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
	"github.com/hwcer/xlsx"
)

func init() {
	xlsx.Config.Package = "config"
	xlsx.Config.Summary = "static"
	logger.SetCallDepth(4)
	logger.Console.Sprintf = func(message *logger.Message) string {
		return message.Content
	}
}

func main() {
	cosgo.Start(false, xlsx.New())
}
