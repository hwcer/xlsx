package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
)

const (
	FlagsNameIn       string = "in"
	FlagsNameGo       string = "go"
	FlagsNameOut      string = "out"
	FlagsNameJson     string = "json"
	FlagsNameIgnore   string = "ignore"
	FlagsNameLanguage string = "language" //多语言文件
)

var mod *module

func init() {
	logger.SetCallDepth(0)
	logger.Console.Disabled = true
	cosgo.Config.Flags(FlagsNameIn, "", "", "需要解析的excel目录")
	cosgo.Config.Flags(FlagsNameOut, "", "", "输出文件目录")
	cosgo.Config.Flags(FlagsNameGo, "", "", "生成的GO文件")
	cosgo.Config.Flags(FlagsNameJson, "", "", "是否导json格式")
	cosgo.Config.Flags(FlagsNameIgnore, "", "", "忽略的文件或者文件夹逗号分割多个")
	cosgo.Config.Flags(FlagsNameLanguage, "", "", "生产的多语言EXCEL文件,默认不生成")
}

func New() *module {
	if mod == nil {
		mod = &module{}
		mod.Id = "xlsx"
	}
	return mod
}

type module struct {
	cosgo.Module
}

func (this *module) Start() error {
	logger.Console.Disabled = false
	var enums = map[string]*enum{}
	if err := cosgo.Config.UnmarshalKey("enum", &enums); err != nil {
		return err
	}
	Config.enums = map[string]*enum{}
	for k, v := range enums {
		pk := TrimProtoName(k)
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
	logger.Console.Disabled = true
	return nil
}
