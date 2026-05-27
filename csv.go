package xlsx

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// OpenFile 根据扩展名选择对应的解析方式，统一返回 *excelize.File
func OpenFile(file string) (*excelize.File, error) {
	ext := strings.ToLower(filepath.Ext(file))
	if ext == ".csv" {
		return openCSV(file)
	}
	return excelize.OpenFile(file)
}

func openCSV(file string) (*excelize.File, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("csv文件打开失败:%v, %w", file, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv文件解析失败:%v, %w", file, err)
	}

	base := filepath.Base(file)
	sheetName := strings.TrimSuffix(base, filepath.Ext(base))

	wb := excelize.NewFile()
	defaultSheet := wb.GetSheetName(0)
	if _, err = wb.NewSheet(sheetName); err != nil {
		return nil, err
	}
	if err = wb.DeleteSheet(defaultSheet); err != nil {
		return nil, err
	}

	for i, row := range rows {
		for j, cell := range row {
			cellName, _ := excelize.CoordinatesToCellName(j+1, i+1)
			_ = wb.SetCellValue(sheetName, cellName, cell)
		}
	}

	return wb, nil
}
