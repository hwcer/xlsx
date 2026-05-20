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

type Parser interface {
	Verify() (skip int, name string, ok bool) //验证表格是否有效
	Fields() []*Field                         //表格字段
}

type enum struct {
	Src   string `json:"src"`
	Index [4]int `json:"index"`
}

type config struct {
	enums                 map[string]*enum     //枚举配置,key:新枚举名,value.Src:源表ProtoName,value.Index:[key,val,type,desc]列索引
	Types                 map[string]SheetType //表结构
	Proto                 string               //proto 文件名
	Empty                 func(string) bool    //检查是否为空
	Package               string               //包名
	GOPackage             string               //Go包名,默认使用Package
	CSPackage             string               //C#包名,默认使用Package
	Parser                func(*Sheet) Parser  //解析器
	Summary               string               //总表名,留空不生成总表
	Message               func() string        //可以加人proto全局对象
	Language              []string             //多语言文本包含的类型
	Outputs               []Output             //附加输出插件
	ProtoHeader           string               //可选,指向一个现有proto文件,其内容将替代TemplateTitle作为文件头部
	JsonNameFilter        func(*Sheet) string  //JSON文件名字
	ProtoNameFilter       func(*Sheet) string  //过滤器
	LanguageNewSheetName  string               //多语言增量页签名
	EnableGlobalDummyName bool                 //允许未显式命名的子对象按签名自动生成名称,为false时必须通过.Name{}/<Name>显式命名
	NamedDummyInHeader    bool                 //显式命名的子对象假定已在ProtoHeader中声明,不注册到全局对象也不生成message定义
	JsonCompact           bool                 //JSON紧凑模式,为true时生成紧凑JSON,为false时生成格式化JSON
}

var Config = &config{
	enums:                 map[string]*enum{},
	Types:                 map[string]SheetType{},
	Proto:                 "configs.proto",
	Empty:                 func(s string) bool { return s == "" },
	Language:              []string{"text", "lang", "language"},
	LanguageNewSheetName:  "多语言文本",
	EnableGlobalDummyName: true,
	JsonCompact:           false,
}

func JsonNameFilterDefault(s *Sheet) string {
	if Config.JsonNameFilter != nil {
		return Config.JsonNameFilter(s)
	}
	return s.ProtoName
}

func ProtoNameFilterDefault(s *Sheet) string {
	if Config.ProtoNameFilter != nil {
		return Config.ProtoNameFilter(s)
	}
	return s.ProtoName
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

// SetEnum 注册一个枚举生成规则
// name 新生成的枚举名; src 源表ProtoName; index [key,val,type,desc]列索引,type/desc为-1表示省略
func (this *config) SetEnum(name, src string, index [4]int) {
	if this.enums == nil {
		this.enums = map[string]*enum{}
	}
	_, pk := TrimProtoName(name)
	_, sk := TrimProtoName(src)
	this.enums[pk] = &enum{Src: sk, Index: index}
}

func (this *config) SetJsonNameFilter(f func(*Sheet) string) {
	this.JsonNameFilter = f
}
func (this *config) SetProtoNameFilter(f func(*Sheet) string) {
	this.ProtoNameFilter = f
}
func init() {
	Config.SetType(SheetTypeHash, "map", "hash")
	Config.SetType(SheetTypeEnum, "kv", "kvs", "obj", "object", "struct")
}

type Output interface {
	Writer(sheets []*Sheet)
}
