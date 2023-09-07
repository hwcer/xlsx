package main

import (
	"github.com/hwcer/logger"
	xlsx2 "github.com/hwcer/xlsx"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

func init() {
	xlsx2.Config.Parser = func(sheet *xlsx.Sheet) xlsx2.Parser {
		return &header{sheet: sheet}
	}
}

type header struct {
	sheet *xlsx.Sheet
}

func (this *header) Verify() (skip int, name string, ok bool) {
	skip = 4
	//isok
	r, e := this.sheet.Row(0)
	if e != nil {
		return
		logger.Fatal("获取sheet行错误 name:%v,err:%v", this.sheet.Name, e)
	}
	cell := r.GetCell(0)
	ok = cell != nil && cell.Value != ""
	//name
	name = cell.Value
	return
}
func (this *header) Fields() (r []*xlsx2.Field) {
	var row *xlsx.Row
	var err error
	if row, err = this.sheet.Row(1); err != nil {
		return
	}
	sheetType := map[int]string{}
	for j := 0; j <= this.sheet.MaxCol; j++ {
		if c := row.GetCell(j); c != nil && c.Value != "" {
			sheetType[j] = xlsx2.FormatType(strings.TrimSpace(c.Value))
		}
	}

	if row, err = this.sheet.Row(3); err != nil {
		return
	}
	sheetDesc := map[int]string{}
	for j := 0; j <= this.sheet.MaxCol; j++ {
		if c := row.GetCell(j); c != nil {
			sheetDesc[j] = c.Value
		}
	}

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
			}
			field = &Field{}
		}
	}
	return
}

func (this *header) SheetType() (r xlsx2.SheetType) {
	row, err := this.sheet.Row(0)
	if err != nil {
		return
	}
	if c := row.GetCell(1); c != nil {
		r = xlsx2.Config.GetTableType(c.Value)
	}
	return
}
