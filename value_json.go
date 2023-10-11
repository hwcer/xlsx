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
	//data := map[string]any{}
	var errs []error
	for _, sheet := range sheets {
		if v, e := sheet.Values(); len(e) == 0 {
			if e2 := writeFile(sheet.ProtoName, v); e2 != nil {
				errs = append(errs, e2)
			}
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

}

func writeFile(name string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	file := filepath.Join(cosgo.Config.GetString(FlagsNameJson), name+".json")
	if err = os.WriteFile(file, b, os.ModePerm); err != nil {
		return err
	}
	//logger.Trace("JSON Data File:%v", file)
	return nil
}
