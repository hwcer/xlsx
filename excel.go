package xlsx

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"

	"github.com/xuri/excelize/v2"
)

func LoadExcel(dir string) {
	logger.Trace("====================开始解析静态数据====================")
	var sheets []*Sheet
	filter := map[string]*Sheet{}
	stat, err := os.Stat(dir)
	if err != nil {
		logger.Fatal(err)
	}

	var files []string

	if stat.IsDir() {
		files = GetFiles(dir, Ignore)
	} else {
		files = append(files, dir)
		dir = filepath.Dir(dir)
	}

	var protoIndex int
	var wb *excelize.File
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}
		wb, err = excelize.OpenFile(file)
		if err != nil {
			logger.Fatal("excel文件格式错误:%v\n%v", file, err)
			return
		}

		//wb, err := xlsx.OpenFile(file)
		logger.Trace("解析文件:%v", file)

		fileName := strings.TrimPrefix(file, dir)
		for _, sheetName := range wb.GetSheetList() {
			if strings.HasPrefix(sheetName, "~") {
				continue
			}

			for k, v := range parseSheet(wb, fileName, sheetName) {
				//lowerName := strings.ToLower(v.ProtoName)
				if i, ok := filter[k]; ok {
					logger.Alert("表格名字[%v]重复自动跳过", v.ProtoName)
					logger.Alert("----sheet:%v,file:%v", v.Name, v.FileName)
					logger.Alert("----sheet:%v,file:%v", i.Name, i.FileName)
				} else {
					protoIndex += 1
					//v.FileName = file
					v.ProtoIndex = protoIndex
					filter[k] = v
					sheets = append(sheets, v)
				}
			}
		}
		_ = wb.Close()
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

func parseSheet(wb *excelize.File, fileName string, sheetName string) (sheets map[string]*Sheet) {
	sheets = map[string]*Sheet{}
	logger.Trace("----开始读取表格[%v]", sheetName)
	sheet := &Sheet{excel: wb, SheetType: SheetTypeHash}
	sheet.Name = Convert(sheet.Name)
	sheet.FileName = fileName
	sheet.SheetName = sheetName
	sheet.Parser = Config.Parser(sheet)
	var ok bool
	if sheet.Skip, sheet.SheetName, ok = sheet.Parser.Verify(); !ok {
		return nil
	}
	if sheet.SheetName, ok = VerifyName(sheet.SheetName); !ok {
		return nil
	}

	sheet.ProtoName = TrimProtoName(sheet.SheetName)
	if sheet.ProtoName == "" {
		return nil
	}

	fields := sheet.Parser.Fields()
	if len(fields) == 0 {
		return nil
	}
	for pk, e := range Config.enums {
		if e.Src == sheet.ProtoName {
			if err := sheet.AddEnum(pk, e.Index); err != nil {
				logger.Trace("add enums:%v   error:%v", pk, err)
			}
		}
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
	if len(sheet.Fields) > 0 {
		sheets[strings.ToUpper(sheet.SheetName)] = sheet
	}
	for _, s := range sheet.sheetAttach {
		switch s.t {
		case SheetTypeEnum:
			if newSheet := sheet.reParseEnum(s); newSheet != nil && len(newSheet.Fields) > 0 {
				sheets[strings.ToUpper(newSheet.SheetName)] = newSheet
			}
		default:

		}
	}

	return
}
