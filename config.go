package xlsx

import (
	"strings"
)

type SheetType int8

const (
	SheetTypeHash SheetType = iota //默认
	SheetTypeStruct
	SheetTypeArray
)

//type ProtoRequireHandle interface {
//	Value(*Field, *xlsx.Row) (any, error)
//	Require(*Field) string
//}

type Parser interface {
	Verify() (skip int, name string, ok bool) //验证表格是否有效
	Fields() []*Field                         //表格字段
}

type ParserSheetType interface {
	SheetType() (SheetType, [4]int)
}

// ParserNewStruct 是否生成一个新对象
// name 如果与原始Sheet重名,将覆盖
// index 索引:  key,val,type,desc
// type,desc 为-1时将省力,类型一律为int32
type ParserNewStruct interface {
	NewStruct() map[string][4]int
}

//type ParserStructIndex interface {
//	StructIndex() [4]int
//}

type config struct {
	Types                map[string]SheetType //表结构
	Proto                string               //proto 文件名
	Package              string               //包名
	Summary              string               //总表名,留空不生成总表
	Parser               func(*Sheet) Parser  //解析器
	Message              func() string        //可以加人proto全局对象
	Language             []string             //多语言文本包含的类型
	LanguageNewSheetName string               //多语言增量页签名
}

var Config = &config{
	Types:                map[string]SheetType{},
	Proto:                "message.proto",
	Package:              "data",
	Summary:              "data",
	Language:             []string{"text", "lang", "language"},
	LanguageNewSheetName: "新增文本",
}

func (this *config) SetType(t SheetType, names ...string) {
	for _, k := range names {
		this.Types[strings.ToUpper(k)] = t
	}
}

func (this *config) GetType(name string) SheetType {
	k := strings.ToUpper(name)
	return this.Types[k]
}

func init() {
	Config.SetType(SheetTypeHash, "map")
	Config.SetType(SheetTypeArray, "arr", "array", "slice")
	Config.SetType(SheetTypeStruct, "kv", "kvs", "obj", "object")
}
