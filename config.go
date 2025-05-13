package xlsx

import (
	"strings"
)

type SheetType int8

const (
	SheetTypeHash SheetType = iota //默认 map
	SheetTypeEnum                  //struct
)

const (
	VersionTagChar = "#"
)

//type ProtoRequireHandle interface {
//	Value(*Field, *xlsx.Row) (any, error)
//	Require(*Field) string
//}

type Parser interface {
	Verify() (skip int, name string, ok bool) //验证表格是否有效
	Fields() []*Field                         //表格字段
}

//type ParserSheetType interface {
//	SheetType() (SheetType, [4]int)
//}

// ParserNewStruct 是否生成一个新对象
// name 如果与原始Sheet重名,将覆盖
// index 索引:  key,val,type,desc
// type,desc 为-1时将省力,类型一律为int32
//type ParserNewStruct interface {
//	NewStruct() map[string][4]int
//}

//type ParserStructIndex interface {
//	StructIndex() [4]int
//}

type enum struct {
	Src string `json:"src"`
	//Name  string `json:"name"`
	Index [4]int `json:"index"`
}

type config struct {
	enums                 map[string]*enum
	Types                 map[string]SheetType           //表结构
	Proto                 string                         //proto 文件名
	Empty                 func(string) bool              //检查是否为空
	Package               string                         //包名
	Parser                func(*Sheet) Parser            //解析器
	Summary               string                         //总表名,留空不生成总表
	Message               func() string                  //可以加人proto全局对象
	Language              []string                       //多语言文本包含的类型
	Outputs               []Output                       //附加输出插件
	ProtoNameFilter       func(SheetType, string) string //过滤器
	LanguageNewSheetName  string                         //多语言增量页签名
	EnableGlobalDummyName bool                           //允许自定义全局对象名
}

var Config = &config{
	enums:                map[string]*enum{},
	Types:                map[string]SheetType{},
	Proto:                "configs.proto",
	Empty:                func(s string) bool { return s == "" },
	Package:              "data",
	Summary:              "data",
	Language:             []string{"text", "lang", "language"},
	ProtoNameFilter:      func(sheetType SheetType, s string) string { return s },
	LanguageNewSheetName: "多语言文本",
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

func (this *config) SetOutput(o Output) {
	this.Outputs = append(this.Outputs, o)
}
func (this *config) SetProtoNameFilter(f func(SheetType, string) string) {
	this.ProtoNameFilter = f
}
func init() {
	Config.SetType(SheetTypeHash, "map", "hash")
	//Config.SetType(SheetTypeArray, "arr", "array", "slice")
	Config.SetType(SheetTypeEnum, "kv", "kvs", "obj", "object", "struct")
}

type Output interface {
	Writer(sheets []*Sheet)
}
