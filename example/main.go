package main

import (
	"strings"

	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
	"github.com/hwcer/xlsx"
	"github.com/hwcer/xlsx/sample"
)

func init() {
	xlsx.Config.Parser = sample.New
	logger.SetCallDepth(4)
	logger.Console.Sprintf = func(message *logger.Message) *strings.Builder {
		b := &strings.Builder{}
		b.WriteString(message.Content)
		return b
	}
}

func main() {
	cosgo.Start(false, xlsx.New())
}
