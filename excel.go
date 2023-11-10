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
		//logger.Trace("解析文件:%v", file)
		if err != nil {
			logger.Fatal("excel文件格式错误:%v\n%v", file, err)
		}
		for _, sheet := range wb.Sheets {
			for k, v := range parseSheet(sheet) {
				//lowerName := strings.ToLower(v.ProtoName)
				if i, ok := filter[k]; ok {
					logger.Alert("表格名字[%v]重复自动跳过\n----FROM:%v\n----TODO:%v", v.Name, i.Name, file)
				} else {
					protoIndex += 1
					v.FileName = file
					v.ProtoIndex = protoIndex
					filter[k] = v
					sheets = append(sheets, v)
				}
			}
		}
	}
	if cosgo.Config.GetString(FlagsNameOut) != "" {
		writeExcelIndex(sheets)
		writeProtoMessage(sheets)
	}
	if cosgo.Config.GetString(FlagsNameJson) != "" {
		writeValueJson(sheets)
	}
	if cosgo.Config.GetString(FlagsNameGo) != "" {
		ProtoGo()
	}
	if p := cosgo.Config.GetString(FlagsNameLanguage); p != "" {
		writeLanguage(sheets)
	}
}

func parseSheet(v *xlsx.Sheet) (sheets map[string]*Sheet) {
	sheets = map[string]*Sheet{}
	//countArr := []int{1, 101, 201, 301}
	//maxRow := v.MaxRow
	//logger.Trace("----开始读取表格[%v],共有%v行", v.Name, maxRow)
	sheet := &Sheet{Sheet: v}
	sheet.Parser = Config.Parser(sheet)
	var ok bool
	if sheet.Skip, sheet.ProtoName, ok = sheet.Parser.Verify(); !ok {
		return nil
	}
	var pt ParserSheetType
	if pt, ok = sheet.Parser.(ParserSheetType); ok {
		sheet.SheetType, sheet.SheetIndex = pt.SheetType()
	}
	if sheet.Fields = sheet.Parser.Fields(); len(sheet.Fields) == 0 {
		//logger.Debug("表[%v]字段为空已经跳过", sheet.SheetName)
		return nil
	}
	//格式化ProtoName
	sheet.ProtoName = TrimProtoName(sheet.ProtoName)

	if sheet.ProtoName == "" || strings.HasPrefix(sheet.Name, "~") || strings.HasPrefix(sheet.ProtoName, "~") {
		return nil
	}

	//sheet.LowerName = strings.ToLower(sheet.ProtoName)
	var index int
	var fields []*Field
	for _, field := range sheet.Fields {
		if h := Require(field.ProtoType); h == nil {
			logger.Alert("****************未知的数据类型,Sheet:%v ,Type:%v", sheet.Name, field.ProtoType)
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
	if sheet.SheetType == SheetTypeStruct {
		sheet.reParseObjField()
	}
	sheets[sheet.ProtoName] = sheet
	//格外的Struct
	var ps ParserNewStruct
	if ps, ok = sheet.Parser.(ParserNewStruct); ok {
		if attach := ps.NewStruct(); len(attach) > 0 {
			for name, sheetIndex := range attach {
				newSheet := *sheet
				newSheet.ProtoName = TrimProtoName(name)
				newSheet.SheetType = SheetTypeStruct
				newSheet.SheetIndex = sheetIndex
				newSheet.reParseObjField()
				sheets[newSheet.ProtoName] = &newSheet
			}
		}
	}
	return
}
