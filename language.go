package xlsx

import (
	"os"
	"path/filepath"
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
	logger.Trace("解析文件:%v", file)

	isCSV := strings.ToLower(filepath.Ext(file)) == ".csv"

	rows := map[string]int{} // key -> row index
	// CSV 只有单页,以文件名作为 sheet 名;xlsx 追加到独立的增量页签
	sheetName := Config.LanguageNewSheetName
	if isCSV {
		base := filepath.Base(file)
		sheetName = strings.TrimSuffix(base, filepath.Ext(base))
	}

	header := Config.LanguageHeader
	useHeader := len(header) > 0

	var wb *excelize.File
	freshSheet := false // 工作表是新建的(无已有数据,需要写入表头)
	if _, statErr := os.Stat(file); statErr != nil {
		// 文件不存在,根据扩展名自动生成
		wb = excelize.NewFile()
		defaultSheet := wb.GetSheetName(0)
		if defaultSheet != sheetName {
			if _, err := wb.NewSheet(sheetName); err != nil {
				logger.Fatal(err)
			}
			if err := wb.DeleteSheet(defaultSheet); err != nil {
				logger.Fatal(err)
			}
		}
		freshSheet = true
	} else {
		var err error
		wb, err = OpenFile(file)
		if err != nil {
			logger.Fatal("多语言文件格式错误:%v\n%v", file, err)
		}
		// 检查工作表是否存在
		if index, e := wb.GetSheetIndex(sheetName); e != nil || index == -1 {
			// 创建工作表
			if _, err = wb.NewSheet(sheetName); err != nil {
				logger.Fatal(err)
			}
			freshSheet = true
		} else {
			// 获取已有数据(固定跳过第一行表头)
			getExistLanguage(wb, sheetName, rows, useHeader)
		}
	}
	defer wb.Close()

	// lastRow 为当前工作表已占用的最大行号,新增文本从其后追加
	lastRow := maxUsedRow(wb, sheetName)
	// 新建工作表时写入表头(第一行),保证后续运行第一行始终被识别为表头
	if freshSheet && useHeader && lastRow == 0 {
		for j, h := range header {
			wb.SetCellValue(sheetName, getCellName(j+1, 1), h)
		}
		lastRow = 1
	}

	types := map[string]bool{}
	for _, k := range Config.LanguageTypes {
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
		// 从已占用的最大行号之后追加,避免覆盖表头与已有数据
		for i, k := range keys {
			v := text[k]
			rowIdx := lastRow + 1 + i
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

	if isCSV {
		if err = saveCSV(wb, sheetName, file); err != nil {
			logger.Fatal("保存文件失败:%v", err)
		}
	} else if err = wb.SaveAs(file); err != nil {
		logger.Fatal("保存文件失败:%v", err)
	}
}

// getExistLanguage 读取已有 key 到 rows。skipHeader 为 true 时固定跳过第一行表头,
// 表头不会被当作翻译 key,从而在写回时保持原样。
func getExistLanguage(wb *excelize.File, sheetName string, rows map[string]int, skipHeader bool) {
	// 获取所有行
	allRows, err := wb.GetRows(sheetName)
	if err != nil {
		return
	}
	for i, row := range allRows {
		if skipHeader && i == 0 {
			continue // 第一行为表头,不作为翻译 key
		}
		if len(row) > 0 {
			id := strings.TrimSpace(row[0])
			if !utils.Empty(id) {
				rows[id] = i + 1 // excelize 行号从1开始
			}
		}
	}
}

// maxUsedRow 返回工作表当前已占用的最大行号(含表头),空表返回 0
func maxUsedRow(wb *excelize.File, sheetName string) int {
	allRows, err := wb.GetRows(sheetName)
	if err != nil {
		return 0
	}
	return len(allRows)
}

// getCellName 将列号和行号转换为单元格名称，如 (1, 1) -> "A1", (2, 3) -> "B3"
func getCellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
