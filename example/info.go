package main

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/xlsx"
	"strings"
)

const FlagsNameInfo string = "info"

func init() {
	i := &infoOutput{}
	xlsx.Config.SetOutput(i)
	xlsx.Config.SetJsonNameFilter(i.JsonNameFilter)
	xlsx.Config.SetProtoNameFilter(i.ProtoNameFilter)
	cosgo.Config.Flags(FlagsNameInfo, "", "", "生成索引文件路径")
}

// 添加一个索引表
type infoOutput struct{}

func (i infoOutput) JsonNameFilter(sheet *xlsx.Sheet) string {
	tag := strings.ToUpper(cosgo.Config.GetString(xlsx.FlagsNameTag))
	if tag != "C" {
		return sheet.ProtoName
	}
	return sheet.SheetName
}

func (i infoOutput) ProtoNameFilter(sheet *xlsx.Sheet) string {
	tag := strings.ToUpper(cosgo.Config.GetString(xlsx.FlagsNameTag))
	if tag != "C" {
		return sheet.ProtoName
	}
	if sheet.SheetType == xlsx.SheetTypeHash {
		return fmt.Sprintf("%sRow", sheet.ProtoName)
	} else {
		return fmt.Sprintf("%sTable", sheet.ProtoName)
	}
}

func (i infoOutput) Writer(sheets []*xlsx.Sheet) {
	path := cosgo.Config.GetString(FlagsNameInfo)
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

		s[sheet.SheetName] = v
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
