package xlsx

import (
	"github.com/tealeg/xlsx/v3"
	"strings"
)

type SheetType int8

const (
	TableTypeMap SheetType = iota //默认
	TableTypeObject
	TableTypeArray
)

type ProtoRequireHandle interface {
	Value(*Field, *xlsx.Row) (any, error)
	Require(*Field) string
}

type Parser interface {
	Verify() (skip int, name string, ok bool) //验证表格是否有效
	Fields() []*Field                         //表格字段
}
type ParserSheetType interface {
	SheetType() SheetType
}

type ParserStructType interface {
	StructType() [4]int
}
type config struct {
	Proto   string                   //proto 文件名
	Package string                   //包名
	Summary string                   //总表名,留空不生成总表
	Parser  func(*xlsx.Sheet) Parser //解析器
	Tables  map[string]SheetType     //表结构
	Message func() string            //可以加人proto全局对象
}

var Config = &config{
	Proto:   "message.proto",
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
	Config.SetTableType(TableTypeArray, "arr", "array", "slice")
	Config.SetTableType(TableTypeObject, "kv", "kvs", "obj", "object")
}
