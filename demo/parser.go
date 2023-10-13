package main

import (
	"fmt"
	cosxls "github.com/hwcer/xlsx"
	"github.com/tealeg/xlsx/v3"
)

func init() {
	cosxls.Config.Parser = func(sheet *xlsx.Sheet) cosxls.Parser {
		return &parser{sheet: sheet}
	}
}

type parser struct {
	sheet *xlsx.Sheet
}

func (this *parser) Verify() (skip int, name string, ok bool) {
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
	name = cell.Value
	return
}
func (this *parser) Fields() (r []*cosxls.Field) {
	var row *xlsx.Row
	var err error

	//proto buff type
	if row, err = this.sheet.Row(1); err != nil {
		return
	}
	sheetType := map[int]cosxls.ProtoBuffType{}
	for j := 0; j <= this.sheet.MaxCol; j++ {
		if c := row.GetCell(j); c != nil && c.Value != "" {
			sheetType[j] = cosxls.ProtoBuffTypeFormat(c.Value)
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
		if end = field.parse(protoType, row.GetCell(j), j); end {
			if field.compile() {
				r = append(r, &field.Field)
				field = &Field{}
			}
		}
	}
	return
}

func (this *parser) SheetType() (r cosxls.SheetType) {
	row, err := this.sheet.Row(0)
	if err != nil {
		return
	}
	if c := row.GetCell(1); c != nil {
		r = cosxls.Config.GetTableType(c.Value)
	}
	return
}

// StructType Struct表(kv模式)下解析方式
// key index
// val index
// type index type默认为int32
func (this *parser) StructType() [4]int {
	return [4]int{0, 1, 2, 3} //默认值,仅仅演示
}
