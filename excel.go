package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

func LoadExcel(dir string) {
	logger.Trace("====================开始解析静态数据====================")
	var sheets []*Sheet
	filter := map[string]*Sheet{}
	files := GetFiles(dir, Ignore)
	var protoIndex int
	for _, file := range files {
		//wb, err := spreadsheet.Open(file)
		wb, err := xlsx.OpenFile(file)
		logger.Trace("解析文件:%v", file)
		if err != nil {
			logger.Fatal("excel文件格式错误:%v\n%v", file, err)
		}
		for _, sheet := range wb.Sheets {
			protoIndex += 1
			if v := parseSheet(sheet, protoIndex); v != nil {
				if i, ok := filter[v.LowerName]; ok {
					logger.Alert("表格名字[%v]重复自动跳过\n----FROM:%v\n----TODO:%v", v.ProtoName, i.FileName, file)
				} else {
					v.FileName = file
					filter[v.LowerName] = v
					sheets = append(sheets, v)
				}
			}
		}
	}
	writeExcelIndex(sheets)
	writeProtoMessage(sheets)
	if cosgo.Config.GetString(FlagsNameJson) != "" {
		writeValueJson(sheets)
	}
	if cosgo.Config.GetString(FlagsNameGo) != "" {
		ProtoGo()
	}

}

func parseSheet(sheet *xlsx.Sheet, index int) (r *Sheet) {
	//countArr := []int{1, 101, 201, 301}
	max := sheet.MaxRow
	logger.Trace("----开始读取表格[%v],共有%v行", sheet.Name, max)
	_ = parseHeader(sheet)
	//if r != nil {
	//	r.ProtoIndex = index
	//}
	return
}

func parseHeader(v *xlsx.Sheet) (sheet *Sheet) {
	sheet = &Sheet{SheetName: v.Name, SheetRows: v}
	parse := Config.Parser(v)
	var ok bool
	if sheet.SheetSkip, sheet.ProtoName, ok = parse.Verify(); !ok {
		return nil
	}
	if i, ok := parse.(ParserSheetType); ok {
		sheet.SheetType = i.SheetType()
	}
	if sheet.Fields = parse.Fields(); len(sheet.Fields) == 0 {
		logger.Debug("表[%v]字段为空已经跳过", sheet.SheetName)
		return nil
	}

	if sheet.ProtoName == "" || strings.HasPrefix(sheet.SheetName, "~") || strings.HasPrefix(sheet.ProtoName, "~") {
		return nil
	}
	sheet.LowerName = strings.ToLower(sheet.ProtoName)
	var index int = 1
	for _, field := range sheet.Fields {
		field.ProtoIndex = index
		index++
		if field.ProtoRequire == FieldTypeNone {
			field.ProtoDesc = strings.ReplaceAll(field.ProtoDesc, "\n", "")
		} else {
			field.ProtoDesc = field.ProtoName
		}
	}

	if sheet.SheetType == TableTypeObj {
		sheet.reParseObjField()
	}

	return
}
