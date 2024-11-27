package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"path/filepath"
)

func writeValueJson(sheets []*Sheet) {
	logger.Trace("======================开始生成JSON数据======================")
	data := map[string]any{}
	var errs []error
	for _, sheet := range sheets {
		if v, e := sheet.Values(); len(e) == 0 {
			data[sheet.ProtoName] = v
			//if e2 := WriteFile(sheet.ProtoName, v); e2 != nil {
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
		WriteFile(path, data)
	} else {
		for k, v := range data {
			file := filepath.Join(path, k+".json")
			WriteFile(file, v)
		}
	}
}
