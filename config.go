package xlsx

import "strings"

type TableType int8

const (
	TableTypeMap TableType = iota //默认
	TableTypeObj
	TableTypeArr
)

type config struct {
	Suffix  string               //表名结尾
	Package string               //包名
	Summary string               //总表名,留空不生成总表
	Tables  map[string]TableType //表结构
}

var Config = &config{
	Suffix:  "",
	Package: "data",
	Summary: "data",
	Tables:  map[string]TableType{},
}

func (this *config) SetTableType(t TableType, names ...string) {
	for _, k := range names {
		this.Tables[strings.ToUpper(k)] = t
	}
}

func (this *config) GetTableType(name string) TableType {
	k := strings.ToUpper(name)
	return this.Tables[k]
}

func init() {
	Config.SetTableType(TableTypeMap, "map")
	Config.SetTableType(TableTypeObj, "kv", "kvs", "obj", "object")
	Config.SetTableType(TableTypeArr, "arr", "array", "slice")
}
