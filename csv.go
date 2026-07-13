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
	// 去除 UTF-8 BOM,避免首列内容附带 BOM 导致 key 不一致
	if len(rows) > 0 && len(rows[0]) > 0 {
		rows[0][0] = strings.TrimPrefix(rows[0][0], "\xEF\xBB\xBF")
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

// saveCSV 将 excelize 内存工作簿的指定 sheet 导出为 CSV 文件,写入 UTF-8 BOM 兼容 Excel 打开
func saveCSV(wb *excelize.File, sheetName, file string) error {
	rows, err := wb.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取工作表失败:%v, %w", sheetName, err)
	}
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("csv文件创建失败:%v, %w", file, err)
	}
	defer f.Close()

	if _, err = f.WriteString("\xEF\xBB\xBF"); err != nil {
		return err
	}
	w := csv.NewWriter(f)
	if err = w.WriteAll(rows); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
