package xlsx

import (
	"encoding/json"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
	"os"
	"path/filepath"
)

func writeValueJson(sheets []*Sheet) {
	logger.Trace("======================开始生成JSON数据======================")
	data := map[string]any{}
	var errs []error
	for _, sheet := range sheets {
		if v, e := sheet.Values(); len(e) == 0 {
			data[sheet.ProtoName] = v
			//if e2 := writeFile(sheet.ProtoName, v); e2 != nil {
			//	errs = append(errs, e2)
			//}
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
	path := cosgo.Config.GetString(FlagsNameJson)
	if filepath.Ext(path) == ".json" {
		writeFile(path, data)
	} else {
		for k, v := range data {
			file := filepath.Join(path, k+".json")
			writeFile(file, v)
		}
	}
}

func writeFile(file string, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		logger.Error("writeFile:%v", err)
		return
	}

	//file := filepath.Join(cosgo.Config.GetString(FlagsNameJson), name+".json")
	if err = os.WriteFile(file, b, os.ModePerm); err != nil {
		logger.Error("writeFile:%v", err)
	}
}
