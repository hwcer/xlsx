package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
)

const (
	FlagsNameIn       string = "in"
	FlagsNameGo       string = "go"
	FlagsNameOut      string = "out"
	FlagsNameTag      string = "tag"
	FlagsNameJson     string = "json"
	FlagsNameIgnore   string = "ignore"   //忽略列表
	FlagsNameBranch   string = "branch"   //使用特定版本分支
	FlagsNameSummary  string = "summary"  //设置总表名称,设置为空时不输出总表
	FlagsNameLanguage string = "language" //多语言文件

)

var mod *Module

func init() {
	logger.SetCallDepth(0)
	logger.DelOutput(logger.DefaultConsoleName)
	cosgo.Config.Flags(FlagsNameIn, "", "", "需要解析的excel目录")
	cosgo.Config.Flags(FlagsNameOut, "", "", "输出文件目录")
	cosgo.Config.Flags(FlagsNameTag, "", "S", "字段标记，一般用来区分前后端字段,格式 客户端:C,服务器:S")
	cosgo.Config.Flags(FlagsNameGo, "", "", "生成的GO文件")
	cosgo.Config.Flags(FlagsNameJson, "", "", "是否导json格式")
	cosgo.Config.Flags(FlagsNameIgnore, "", "", "忽略的文件或者文件夹逗号分割多个")
	cosgo.Config.Flags(FlagsNameBranch, "", "", "使用特定版本分支")
	cosgo.Config.Flags(FlagsNameSummary, "", "", "设置总表名称,默认data,设置为空时不输出总表")
	cosgo.Config.Flags(FlagsNameLanguage, "", "", "生产的多语言EXCEL文件,默认不生成")
}

func New() *Module {
	if mod == nil {
		mod = &Module{}
		mod.Module = *cosgo.NewModule("xlsx")
	}
	return mod
}

type Module struct {
	cosgo.Module
}

func (this *Module) Start() error {
	_ = logger.SetOutput(logger.DefaultConsoleName, logger.Console)
	var enums = map[string]*enum{}
	if err := cosgo.Config.UnmarshalKey("enum", &enums); err != nil {
		return err
	}
	Config.enums = map[string]*enum{}
	for k, v := range enums {
		var pk string
		if v.Src != "" {
			pk = TrimProtoName(v.Src)
		} else {
			pk = TrimProtoName(k)
		}

		if v.Name == "" {
			v.Name = pk
		} else {
			v.Name = TrimProtoName(v.Name)
		}
		Config.enums[pk] = v
	}

	preparePath()
	LoadExcel(cosgo.Config.GetString(FlagsNameIn))
	logger.Trace("\n========================恭喜大表哥导表成功========================\n")
	logger.DelOutput(logger.DefaultConsoleName)
	return nil
}
