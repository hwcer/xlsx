package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
)

func writeValueInfo(sheets []*Sheet) {
	logger.Trace("======================开始生成INFO数据======================")
	info := map[string]any{}
	tableList := make([]string, 0, len(sheets))
	for _, sheet := range sheets {
		tableList = append(tableList, sheet.ProtoName)
		tableInfo := map[string]any{}
		if sheet.SheetType == SheetTypeStruct {
			tableInfo["type"] = "kv"
			tableInfo["tableType"] = fmt.Sprintf("%vTable", sheet.ProtoName)
		} else if sheet.SheetType == SheetTypeHash {
			tableInfo["type"] = "normal"
			tableInfo["rowType"] = fmt.Sprintf("%vRow", sheet.ProtoName)
		} else if sheet.SheetType == SheetTypeArray {
			tableInfo["type"] = "array"
			tableInfo["rowArrayType"] = fmt.Sprintf("%vRowArray", sheet.ProtoName)
		}
		tableInfo["file"] = sheet.FileName
		info[sheet.ProtoName] = tableInfo
	}

	data := map[string]any{}
	data["info"] = info
	data["tableList"] = tableList
	path := cosgo.Config.GetString(FlagsNameInfo)
	writeFile(path, data)
}
