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
		//writeExcelIndex(sheets)
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
	for _, out := range Config.Outputs {
		out.Writer(sheets)
	}
	globalObjects = map[string]*Dummy{}
}

func parseSheet(v *xlsx.Sheet) (sheets map[string]*Sheet) {
	//tag := strings.ToUpper(cosgo.Config.GetString(FlagsNameTag))

	sheets = map[string]*Sheet{}
	//countArr := []int{1, 101, 201, 301}
	maxRow := v.MaxRow
	logger.Trace("----开始读取表格[%v],共有%v行", v.Name, maxRow)
	sheet := &Sheet{Sheet: v, SheetType: SheetTypeHash}
	sheet.Name = Convert(sheet.Name)
	sheet.Parser = Config.Parser(sheet)
	var ok bool
	if sheet.Skip, sheet.SheetName, ok = sheet.Parser.Verify(); !ok {
		return nil
	}
	if sheet.ProtoName, ok = VerifyName(sheet.SheetName); !ok {
		return nil
	}
	//格式化ProtoName
	sheet.ProtoName = TrimProtoName(sheet.ProtoName)
	if sheet.ProtoName == "" || strings.HasPrefix(sheet.Name, "~") || strings.HasPrefix(sheet.ProtoName, "~") {
		return nil
	}

	//var pt ParserSheetType
	//if pt, ok = sheet.Parser.(ParserSheetType); ok {
	//	sheet.SheetType, sheet.SheetIndex = pt.SheetType()
	//}

	fields := sheet.Parser.Fields()
	if len(fields) == 0 {
		//logger.Debug("表[%v]字段为空已经跳过", sheet.SheetName)
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
	//if sheet.SheetType == SheetTypeEnum {
	//	sheet.reParseEnum()
	//}
	////ARRAY
	//if sheet.SheetType == SheetTypeArray {
	//	name := Config.ProtoNameFilter(SheetTypeHash, sheet.ProtoName)
	//	dummy := NewDummy(name)
	//	for _, field := range sheet.Fields {
	//		if len(field.Index) > 0 {
	//			_ = dummy.Add(field.Name, field.ProtoType, field.Index[0])
	//		}
	//	}
	//	globalObjects.Insert(sheet, dummy, true)
	//	sheet.DummyName = dummy.Name
	//	sheet.ProtoName = sheet.ProtoName + Config.ArraySuffix
	//}
	//protoName := sheet.ProtoName
	if len(sheet.Fields) > 0 {
		sheet.ProtoName = Config.ProtoNameFilter(sheet.SheetType, sheet.ProtoName)
		sheets[strings.ToUpper(sheet.SheetName)] = sheet
	}
	for _, s := range sheet.sheetAttach {
		switch s.t {
		case SheetTypeEnum:
			if newSheet := sheet.reParseEnum(s); newSheet != nil && len(newSheet.Fields) > 0 {
				sheets[strings.ToUpper(newSheet.SheetName)] = newSheet
			}
		//case SheetTypeArray:
		//	if newSheet := sheet.reParseArray(s); newSheet != nil && len(newSheet.Fields) > 0 {
		//		newSheet.ProtoName = Config.ProtoNameFilter(newSheet.SheetType, newSheet.ProtoName)
		//		sheets[s.k] = newSheet
		//	}
		default:

		}
	}

	//格外的Struct
	//if ev := Config.enums[sheet.ProtoName]; ev != nil {
	//	newSheet := *sheet
	//	newSheet.ProtoName = ev.Name
	//	newSheet.SheetType = SheetTypeEnum
	//	newSheet.SheetIndex = ev.Index
	//	newSheet.reParseEnum()
	//	if len(newSheet.Fields) > 0 {
	//		sheets[newSheet.ProtoName] = &newSheet
	//	}
	//}
	return
}
