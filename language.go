package xlsx

import (
	"sort"
	"strings"

	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/utils"
	"github.com/hwcer/logger"
	"github.com/xuri/excelize/v2"
)

func writeLanguage(sheets []*Sheet) {
	logger.Trace("======================开始生成多语言文件======================")
	file := cosgo.Config.GetString(FlagsNameLanguage)
	wb, err := excelize.OpenFile(file)
	logger.Trace("解析文件:%v", file)
	if err != nil {
		logger.Fatal("excel文件格式错误:%v\n%v", file, err)
	}
	defer wb.Close()

	rows := map[string]int{} // key -> row index
	sheetName := Config.LanguageNewSheetName

	// 检查工作表是否存在
	index, err := wb.GetSheetIndex(sheetName)
	if err != nil || index == -1 {
		// 创建工作表
		_, err = wb.NewSheet(sheetName)
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		// 获取已有数据
		getExistLanguage(wb, sheetName, rows)
	}

	types := map[string]bool{}
	for _, k := range Config.Language {
		types[strings.ToLower(k)] = true
	}
	text := map[string]string{}
	// 遍历sheets
	for _, s := range sheets {
		s.Language(text, types)
	}
	var keys []string
	var edit int32

	// 创建样式（黄色背景）
	styleID, err := wb.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FFFF00"},
			Pattern: 1,
		},
	})
	if err != nil {
		logger.Fatal("创建样式失败:%v", err)
	}

	for k, v := range text {
		if rowIdx, ok := rows[k]; !ok {
			keys = append(keys, k)
		} else {
			// 获取第二列的值
			cellValue, _ := wb.GetCellValue(sheetName, getCellName(2, rowIdx))
			if cellValue != v {
				edit++
				cellName := getCellName(2, rowIdx)
				wb.SetCellValue(sheetName, cellName, v)
				wb.SetCellStyle(sheetName, cellName, cellName, styleID)
			}
		}
	}
	if len(keys) == 0 && edit == 0 {
		logger.Trace("本次没有新增文本,自动跳过")
		return
	}
	if len(keys) > 0 {
		sort.Strings(keys)
		// 获取当前最大行号
		maxRow := len(rows) + 1
		for i, k := range keys {
			v := text[k]
			rowIdx := maxRow + i
			// 设置第一列（key）
			cellName1 := getCellName(1, rowIdx)
			wb.SetCellValue(sheetName, cellName1, k)
			wb.SetCellStyle(sheetName, cellName1, cellName1, styleID)
			// 设置第二列（value）
			cellName2 := getCellName(2, rowIdx)
			wb.SetCellValue(sheetName, cellName2, v)
			wb.SetCellStyle(sheetName, cellName2, cellName2, styleID)
		}
	}

	if err = wb.Save(); err != nil {
		logger.Fatal("保存文件失败:%v", err)
	}
}

func getExistLanguage(wb *excelize.File, sheetName string, rows map[string]int) {
	// 获取所有行
	allRows, err := wb.GetRows(sheetName)
	if err != nil {
		return
	}
	for i, row := range allRows {
		if len(row) > 0 {
			id := strings.TrimSpace(row[0])
			if !utils.Empty(id) {
				rows[id] = i + 1 // excelize 行号从1开始
			}
		}
	}
}

// getCellName 将列号和行号转换为单元格名称，如 (1, 1) -> "A1", (2, 3) -> "B3"
func getCellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
