package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/utils"
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

	rows := map[string]*xlsx.Row{}
	var sheet *xlsx.Sheet
	for _, s := range wb.Sheets {
		if s.Name == Config.LanguageNewSheetName {
			sheet = s
			getExistLanguage(sheet, rows)
			break
		}
	}
	if sheet == nil {
		sheet, err = wb.AddSheet(Config.LanguageNewSheetName)
		if err != nil {
			logger.Fatal(err)
		}
	}

	types := map[string]bool{}
	for _, k := range Config.Language {
		types[strings.ToLower(k)] = true
	}
	text := map[string]string{}
	//遍历sheets
	for _, s := range sheets {
		s.Language(text, types)
	}
	var keys []string
	var edit int32

	style := xlsx.NewStyle()
	style.Fill = xlsx.Fill{}
	style.Fill.PatternType = "solid"
	style.Fill.FgColor = "FFFF00"

	for k, v := range text {
		if r, ok := rows[k]; !ok {
			keys = append(keys, k)
		} else if c := r.GetCell(1); c != nil && c.Value != v {
			edit++
			c.SetString(v)
			c.SetStyle(style)
		}
	}
	if len(keys) == 0 && edit == 0 {
		logger.Trace("本次没有新增文本,自动跳过")
		return
	}
	if len(keys) > 0 {
		sort.Strings(keys)
		for _, k := range keys {
			v := text[k]
			row := sheet.AddRow()
			for _, x := range []string{k, v} {
				c := row.AddCell()
				c.SetValue(x)
				c.SetStyle(style)
			}
		}
	}
	//
	//newSheet, err := wb.AddSheet(Config.LanguageNewSheetName)
	//if err != nil {
	//	logger.Fatal("创建新页签失败")
	//}
	//for _, k := range keys {
	//	v := rows[k]
	//	newSheet.Add
	//	row := newSheet.AddRow()
	//	for _, x := range []string{k, v} {
	//		row.AddCell().SetValue(x)
	//	}
	//}
	if err = wb.Save(file); err != nil {
		logger.Fatal("保存文件失败:%v", err)
	}
}

func getExistLanguage(sheet *xlsx.Sheet, rows map[string]*xlsx.Row) {
	maxRow := sheet.MaxRow
	for i := 0; i < maxRow; i++ {
		if row, err := sheet.Row(i); err == nil {
			id := strings.TrimSpace(row.GetCell(0).Value)
			if !utils.Empty(id) {
				rows[id] = row
			}
		}
	}
}
