package xlsx

import (
	"github.com/tealeg/xlsx/v3"
	"strings"
)

type SheetType int8

const (
	TableTypeMap SheetType = iota //默认
	TableTypeObj
	//TableTypeArr
)

//type ProtoRequire int8
//
//const (
//	FieldTypeNone   ProtoRequire = 0
//	FieldTypeArray               = -1
//	FieldTypeObject              = -2
//	FieldTypeArrObj              = -3
//)

type ProtoRequireHandle interface {
	Value(*Field, *xlsx.Row) (any, error)
	Require(*Field) string
}

//var protoRequireHandles = map[ProtoRequire]ProtoRequireHandle{}
//
//func init() {
//	//Register(ProtoRequire(0), &defaultProtoRequireHandle{})
//}

//func Register(t ProtoRequire, h ProtoRequireHandle) {
//	protoRequireHandles[t] = h
//}

//type defaultProtoRequireHandle struct {
//}
//
//func (this *defaultProtoRequireHandle) Value(field *Field, rows *xlsx.Row) (any, error) {
//	i := field.Index[0]
//	return FormatValue(rows, i, field.ProtoType)
//}
//func (this *defaultProtoRequireHandle) Require(field *Field) string {
//	return field.ProtoType
//}

type Parser interface {
	Verify() (skip int, name string, ok bool) //验证表格是否有效
	Fields() []*Field                         //表格字段
}
type ParserSheetType interface {
	SheetType() SheetType
}

// StructType Struct表(kv模式)下解析方式
// key index
// val index
// type index type默认为int32
// desc 描述

type ParserStructType interface {
	StructType() [4]int
}
type config struct {
	//Suffix  string                   //表名结尾
	Package string                   //包名
	Summary string                   //总表名,留空不生成总表
	Parser  func(*xlsx.Sheet) Parser //解析器
	Tables  map[string]SheetType     //表结构
	Message func() string            //可以加人proto全局对象
	//Require func(string) ProtoRequire //格式化类型
}

var Config = &config{
	//Suffix:  "",
	Package: "data",
	Summary: "data",
	Tables:  map[string]SheetType{},
}

func (this *config) SetTableType(t SheetType, names ...string) {
	for _, k := range names {
		this.Tables[strings.ToUpper(k)] = t
	}
}

func (this *config) GetTableType(name string) SheetType {
	k := strings.ToUpper(name)
	return this.Tables[k]
}

func init() {
	Config.SetTableType(TableTypeMap, "map")
	Config.SetTableType(TableTypeObj, "kv", "kvs", "obj", "object")
	//Config.SetTableType(TableTypeArr, "arr", "array", "slice")
}
