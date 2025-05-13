package sample

import (
	"fmt"
	"github.com/hwcer/logger"
	cosxls "github.com/hwcer/xlsx"
	"github.com/tealeg/xlsx/v3"
	"strconv"
	"strings"
)

func New(sheet *cosxls.Sheet) cosxls.Parser {
	return &Parser{sheet: sheet}
}

type Parser struct {
	sheet *cosxls.Sheet
}

func s2a(s string) [4]int {
	r := [4]int{-1, -1, -1}
	arr := strings.Split(s, ",")
	for i, v := range arr {
		if i < 4 {
			n, _ := strconv.Atoi(v)
			r[i] = n
		}
	}
	return r
}

// kv:name:2,0
func (this *Parser) Verify() (skip int, name string, ok bool) {
	skip = 4
	//isok
	r, e := this.sheet.Row(0)
	if e != nil {
		fmt.Printf("获取sheet行错误 name:%v,err:%v", this.sheet.Name, e)
		return
	}
	cell := r.GetCell(0)
	ok = cell != nil && cell.Value != ""
	//name
	if ok {
		name = cell.Value
	}
	//attach
	m := this.sheet.MaxCol
	var err error
	for i := 1; i <= m; i++ {
		if cell = r.GetCell(i); cell != nil && cell.Value != "" {
			arr := strings.Split(cell.Value, ":")
			if strings.ToLower(arr[0]) == "kv" {
				if len(arr) == 1 {
					//默认kv 模式
					err = this.sheet.AddEnum(name, [4]int{0, 1, 2, 3})
				} else if len(arr) == 3 {
					err = this.sheet.AddEnum(arr[1], s2a(arr[2]))
				} else {
					err = fmt.Errorf("attach error,sheet:%v,value:%v", this.sheet.Name, cell.Value)
				}
			}
			if err != nil {
				logger.Fatal(err)
			}
		}
	}
	return
}
func (this *Parser) Fields() (r []*cosxls.Field) {
	var row *xlsx.Row
	var err error

	//proto buff type
	if row, err = this.sheet.Row(1); err != nil {
		return
	}
	sheetType := map[int]cosxls.ProtoBuffType{}
	fieldType := map[int]string{}
	for j := 0; j <= this.sheet.MaxCol; j++ {
		if c := row.GetCell(j); c != nil && c.Value != "" {
			sheetType[j] = cosxls.ProtoBuffTypeFormat(c.Value)
			fieldType[j] = strings.ToLower(strings.TrimSpace(c.Value))
		}
	}
	//描述
	if row, err = this.sheet.Row(3); err != nil {
		return
	}
	sheetDesc := map[int]string{}
	for j := 0; j <= this.sheet.MaxCol; j++ {
		if c := row.GetCell(j); c != nil {
			sheetDesc[j] = c.Value
		}
	}
	//字段
	var end bool
	var field = &Field{}
	if row, err = this.sheet.Row(2); err != nil {
		return
	}
	for j := 0; j <= this.sheet.MaxCol; j++ {
		protoType := sheetType[j]
		if field.ProtoDesc == "" {
			field.ProtoDesc = sheetDesc[j]
		}
		if field.FieldType == "" {
			field.FieldType = fieldType[j]
		}
		if end = field.parse(protoType, row.GetCell(j), j); end {
			if field.compile() {
				r = append(r, &field.Field)
				field = &Field{}
			}
		}
	}
	return
}

func (this *Parser) SheetType() (r cosxls.SheetType, index [4]int) {
	row, err := this.sheet.Row(0)
	if err != nil {
		return
	}
	if c := row.GetCell(1); c != nil {
		r = cosxls.Config.GetType(c.Value)
	}
	index = [4]int{0, 1, 2, 3}
	return
}

func (this *Parser) NewStruct() (r map[string][4]int) {
	r = map[string][4]int{}
	if strings.ToUpper(this.sheet.ProtoName) == "EMITTER" {
		r["events"] = [4]int{1, 0, 2, 7}
	}
	return
}

// StructType Struct表(kv模式)下解析方式
// key index
// val index
// type index type默认为int32
//func (this *Parser) StructType(protoName string) [4]int {
//	if name := strings.ToUpper(protoName); name == "EVENTS" {
//		return [4]int{1, 0, 2, 7}
//	}
//	return [4]int{0, 1, 2, 3} //默认值,仅仅演示
//}
