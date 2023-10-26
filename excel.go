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
			for _, v := range parseSheet(sheet) {
				lowerName := strings.ToLower(v.ProtoName)
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

func parseSheet(v *xlsx.Sheet) (sheets []*Sheet) {
	//countArr := []int{1, 101, 201, 301}
	maxRow := v.MaxRow
	logger.Trace("----开始读取表格[%v],共有%v行", v.Name, maxRow)
	sheet := &Sheet{SheetName: v.Name, SheetRows: v}
	sheet.Parser = Config.Parser(v)
	var ok bool
	if sheet.SheetSkip, sheet.ProtoName, ok = sheet.Parser.Verify(); !ok {
		return nil
	}
	if i, ok := sheet.Parser.(ParserSheetType); ok {
		sheet.SheetType, sheet.Alias = i.SheetType()
	}
	if sheet.Fields = sheet.Parser.Fields(); len(sheet.Fields) == 0 {
		logger.Debug("表[%v]字段为空已经跳过", sheet.SheetName)
		return nil
	}
	//格式化ProtoName
	sheet.ProtoName = TrimProtoName(sheet.ProtoName)
	i := strings.Index(sheet.ProtoName, "_")
	for i > 0 {
		sheet.ProtoName = sheet.ProtoName[0:i] + FirstUpper(sheet.ProtoName[i+1:])
		i = strings.Index(sheet.ProtoName, "_")
	}

	if sheet.ProtoName == "" || strings.HasPrefix(sheet.SheetName, "~") || strings.HasPrefix(sheet.ProtoName, "~") {
		return nil
	}

	//sheet.LowerName = strings.ToLower(sheet.ProtoName)
	var index int
	var fields []*Field
	for _, field := range sheet.Fields {
		if h := Require(field.ProtoType); h == nil {
			logger.Alert("****************未知的数据类型,Sheet:%v ,Type:%v", sheet.SheetName, field.ProtoType)
			continue
		}
		index++
		field.ProtoIndex = index
		field.ProtoDesc = strings.ReplaceAll(field.ProtoDesc, "\n", "")
		fields = append(fields, field)
		//
		//if field.ProtoRequire == FieldTypeNone {
		//	field.ProtoDesc = strings.ReplaceAll(field.ProtoDesc, "\n", "")
		//} else {
		//	field.ProtoDesc = field.ProtoName
		//}
	}
	sheet.Fields = fields
	if sheet.SheetType == TableTypeObject {
		if sheet.Alias != "" {
			newSheet := *sheet
			alias := &newSheet
			alias.ProtoName = TrimProtoName(sheet.Alias)
			alias.reParseObjField()
			sheets = append(sheets, alias)
			sheet.SheetType = TableTypeMap
		} else {
			sheet.reParseObjField()
		}

	}
	sheets = append(sheets, sheet)
	return

}

func TrimProtoName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "_")
	s = FirstUpper(s)
	return s
}
