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
			if v := parseSheet(sheet); v != nil {
				lowerName := strings.ToLower(v.SheetName)
				if i, ok := filter[lowerName]; ok {
					logger.Alert("表格名字[%v]重复自动跳过\n----FROM:%v\n----TODO:%v", v.ProtoName, i.FileName, file)
				} else {
					protoIndex += 1
					v.FileName = file
					v.ProtoIndex = protoIndex
					filter[lowerName] = v
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

func parseSheet(v *xlsx.Sheet) (sheet *Sheet) {
	//countArr := []int{1, 101, 201, 301}
	max := v.MaxRow
	logger.Trace("----开始读取表格[%v],共有%v行", v.Name, max)
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
	//格式化ProtoName
	sheet.ProtoName = strings.TrimSpace(sheet.ProtoName)
	sheet.ProtoName = strings.TrimPrefix(sheet.ProtoName, "_")
	sheet.ProtoName = FirstUpper(sheet.ProtoName)
	i := strings.Index(sheet.ProtoName, "_")
	for i > 0 {
		sheet.ProtoName = sheet.ProtoName[0:i] + FirstUpper(sheet.ProtoName[i+1:])
		i = strings.Index(sheet.ProtoName, "_")
	}

	if sheet.ProtoName == "" || strings.HasPrefix(sheet.SheetName, "~") || strings.HasPrefix(sheet.ProtoName, "~") {
		return nil
	}
<<<<<<< HEAD

=======
>>>>>>> cabfa43f3ff1057a9154cc80e61d02d81319fa71
	//sheet.LowerName = strings.ToLower(sheet.ProtoName)
	var index int
	for _, field := range sheet.Fields {
		index++
<<<<<<< HEAD
		field.ProtoIndex = index
		if field.ProtoRequire == FieldTypeNone {
=======
		field.protoIndex = index
		if field.ProtoDesc != "" {
>>>>>>> cabfa43f3ff1057a9154cc80e61d02d81319fa71
			field.ProtoDesc = strings.ReplaceAll(field.ProtoDesc, "\n", "")
		}
	}

	if sheet.SheetType == SheetTypeObj {
		sheet.reParseObjField()
	}

	return

}
