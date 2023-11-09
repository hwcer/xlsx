package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/utils"
	"github.com/hwcer/logger"
	"github.com/tealeg/xlsx/v3"
	"sort"
	"strings"
)

func writeLanguage(sheets []*Sheet) {
	logger.Trace("======================开始生成多语言文件======================")
	file := cosgo.Config.GetString(FlagsNameLanguage)
	wb, err := xlsx.OpenFile(file)
	logger.Trace("解析文件:%v", file)
	if err != nil {
		logger.Fatal("excel文件格式错误:%v\n%v", file, err)
	}
	exist := map[string]int32{}
	for _, sheet := range wb.Sheets {
		if sheet.Name == Config.LanguageNewSheetName {
			logger.Fatal("语言文件中上次新增[%v]没有处理,请先手动合并或者删除后再生成", Config.LanguageNewSheetName)
		}
		getExistLanguage(sheet, exist)
	}

	types := map[string]bool{}
	for _, k := range Config.Language {
		types[strings.ToLower(k)] = true
	}
	rows := map[string]string{}
	//遍历sheets
	for _, sheet := range sheets {
		sheet.Language(rows, types)
	}
	var keys []string
	for k, _ := range rows {
		if exist[k] == 0 {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		logger.Trace("本次没有新增文本,自动跳过")
		return
	}
	sort.Strings(keys)

	newSheet, err := wb.AddSheet(Config.LanguageNewSheetName)
	if err != nil {
		logger.Fatal("创建新页签失败")
	}
	for _, k := range keys {
		v := rows[k]
		row := newSheet.AddRow()
		for _, x := range []string{k, v} {
			row.AddCell().SetValue(x)
		}
	}
	if err = wb.Save(file); err != nil {
		logger.Fatal("保存文件失败:%v", err)
	}
}

func getExistLanguage(sheet *xlsx.Sheet, data map[string]int32) {
	maxRow := sheet.MaxRow
	for i := 0; i < maxRow; i++ {
		if row, err := sheet.Row(i); err == nil {
			id := strings.TrimSpace(row.GetCell(0).Value)
			if !utils.Empty(id) {
				data[id]++
			}
		}
	}
}
