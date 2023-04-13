package xlsx

import (
	"encoding/json"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
	"os"
	"path/filepath"
)

func writeValueJson(sheets []*Message) {
	logger.Trace("======================开始生成JSON数据======================")
	data := map[string]any{}
	var errs []error
	for _, sheet := range sheets {
		if v, e := sheet.Values(); len(e) == 0 {
			data[sheet.ProtoName] = v
		} else {
			errs = append(errs, e...)
		}
	}
	if len(errs) != 0 {
		logger.Trace("生成JSON数据失败")
		for _, err := range errs {
			logger.Trace(err)
		}
		//os.Exit(0)
	}

	b, err := json.Marshal(data)
	if err != nil {
		logger.Fatal(err)
	}

	file := filepath.Join(cosgo.Config.GetString(FlagsNameJson), "data.json")
	if err = os.WriteFile(file, b, os.ModePerm); err != nil {
		logger.Fatal(err)
	}
	logger.Trace("JSON Data File:%v", file)
}
