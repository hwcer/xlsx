package main

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/xlsx"
)

func init() {
	xlsx.Config.SetOutput(&infoOutput{})
}

// 添加一个索引表
type infoOutput struct{}

func (i infoOutput) Writer(sheets []*xlsx.Sheet) {
	path := cosgo.Config.GetString("info")
	if path == "" {
		return
	}

	s := map[string]*Info{}
	var tableList []string
	for _, sheet := range sheets {
		v := &Info{
			File: sheet.FileName,
		}
		if sheet.SheetType == xlsx.SheetTypeHash {
			v.Type = "normal"
			//v.TableClass = fmt.Sprintf("%sTable", sheet.ProtoName)
			v.RowClass = fmt.Sprintf("%sRow", sheet.ProtoName)
			primary := sheet.Fields[0]
			if primary.ProtoType.IsNumber() {
				v.KeyType = "int"
			} else {
				v.KeyType = "string"
			}
		} else {
			v.Type = "kv"
			v.KeyType = "string"
			v.TableClass = fmt.Sprintf("%sTable", sheet.ProtoName)
		}

		s[sheet.ProtoName] = v
		tableList = append(tableList, sheet.SheetName)
	}

	data := map[string]any{}
	data["info"] = s
	data["tableList"] = tableList

	xlsx.WriteFile(cosgo.Abs(path), data)
}

type Info struct {
	File       string `json:"file"`
	RowClass   string `json:"rowClass,omitempty"`
	Type       string `json:"type"`
	TableClass string `json:"tableClass,omitempty"`
	KeyType    string `json:"keyType,omitempty"`
}
