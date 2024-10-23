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
			for k, v := range parseSheet(sheet) {
				//lowerName := strings.ToLower(v.ProtoName)
				if i, ok := filter[k]; ok {
					logger.Alert("表格名字[%v]重复自动跳过", v.ProtoName)
					logger.Alert("----sheet:%v,file:%v", v.Name, v.FileName)
					logger.Alert("----sheet:%v,file:%v", i.Name, i.FileName)
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
	if cosgo.Config.GetString(FlagsNameInfo) != "" {
		writeValueInfo(sheets)
	}
	if cosgo.Config.GetString(FlagsNameGo) != "" {
		ProtoGo()
	}
	if p := cosgo.Config.GetString(FlagsNameLanguage); p != "" {
		writeLanguage(sheets)
	}
}

func parseSheet(v *xlsx.Sheet) (sheets map[string]*Sheet) {
	//tag := strings.ToUpper(cosgo.Config.GetString(FlagsNameTag))
	sheets = map[string]*Sheet{}
	//countArr := []int{1, 101, 201, 301}
	maxRow := v.MaxRow
	logger.Trace("----开始读取表格[%v],共有%v行", v.Name, maxRow)
	sheet := &Sheet{Sheet: v}
	sheet.Name = Convert(sheet.Name)
	sheet.Parser = Config.Parser(sheet)
	var ok bool
	if sheet.Skip, sheet.ProtoName, ok = sheet.Parser.Verify(); !ok {
		return nil
	}
	if sheet.ProtoName, ok = VerifyName(sheet.ProtoName); !ok {
		return nil
	}

	var pt ParserSheetType
	if pt, ok = sheet.Parser.(ParserSheetType); ok {
		sheet.SheetType, sheet.SheetIndex = pt.SheetType()
	}

	fields := sheet.Parser.Fields()
	if len(fields) == 0 {
		//logger.Debug("表[%v]字段为空已经跳过", sheet.SheetName)
		return nil
	}

	//格式化ProtoName
	sheet.ProtoName = TrimProtoName(sheet.ProtoName)

	if sheet.ProtoName == "" || strings.HasPrefix(sheet.Name, "~") || strings.HasPrefix(sheet.ProtoName, "~") {
		return nil
	}
	var index int
	fieldsMap := map[string]*Field{}

	for _, field := range fields {
		if h := Require(field.ProtoType); h == nil {
			logger.Alert("****************未知的数据类型,Sheet:%v ,Type:%v", sheet.Name, field.ProtoType)
			continue
		}
		if field.Name, ok = VerifyName(field.Name); !ok {
			continue
		}
		if i := strings.Index(field.Name, VersionTagChar); i > 0 {
			branch := field.Name[i+1:]
			field.Name = field.Name[:i]
			fm := fieldsMap[field.Name]
			if fm == nil {
				fm = &Field{}
				index++
				fm.Name = field.Name
				fm.FieldType = field.FieldType
				fm.ProtoType = field.ProtoType
				fm.ProtoIndex = index
				fm.ProtoDesc = strings.ReplaceAll(field.ProtoDesc, "\n", "")
				sheet.Fields = append(sheet.Fields, fm)
				fieldsMap[field.Name] = fm
			}
			fm.SetBranch(branch, field)
		} else if fm := fieldsMap[field.Name]; fm != nil {
			fm.Name = field.Name
			fm.Dummy = field.Dummy
			fm.Index = field.Index
			fm.FieldType = field.FieldType
			fm.ProtoType = field.ProtoType
			field.ProtoDesc = strings.ReplaceAll(field.ProtoDesc, "\n", "")
		} else {
			index++
			field.ProtoIndex = index
			field.ProtoDesc = strings.ReplaceAll(field.ProtoDesc, "\n", "")
			sheet.Fields = append(sheet.Fields, field)
			fieldsMap[field.Name] = field
		}

	}
	//sheet.Fields = fields
	if sheet.SheetType == SheetTypeStruct {
		sheet.reParseObjField()
	}
	if len(sheet.Fields) > 0 {
		sheets[sheet.ProtoName] = sheet
	}
	//格外的Struct
	if ev := Config.enums[sheet.ProtoName]; ev != nil {
		newSheet := *sheet
		newSheet.ProtoName = ev.Name
		newSheet.SheetType = SheetTypeStruct
		newSheet.SheetIndex = ev.Index
		newSheet.reParseObjField()
		if len(newSheet.Fields) > 0 {
			sheets[newSheet.ProtoName] = &newSheet
		}
	}
	return
}
