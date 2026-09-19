package xlsx

import (
	"path/filepath"
	"strings"

	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
)

// writeValueJson 生成 JSON 数据；返回值表示是否全部成功。
//
// 一张表只要有一行解析失败，整张表都不会写进 data——宁可整表缺失也不写半张表。
// 但这样一来"少了一整张表"必须让调用方知道：返回 false 由 Module.Start 转成非零退出码，
// 否则 CI 与 export.bat 会拿着缺表的 JSON 一路绿灯往下走。
func writeValueJson(sheets []*Sheet) (ok bool) {
	logger.Trace("======================开始生成JSON数据======================")
	data := map[string]any{}
	var errs []error
	var failed []string
	for _, sheet := range sheets {
		if v, e := sheet.Values(); len(e) == 0 {
			name := JsonNameFilterDefault(sheet)
			data[name] = v
		} else {
			errs = append(errs, e...)
			failed = append(failed, sheet.ProtoName)
		}
	}
	if len(errs) != 0 {
		logger.Alert("生成JSON数据失败,以下表格已被整表丢弃:%v", strings.Join(failed, ","))
		for _, err := range errs {
			logger.Alert(err)
		}
	}
	path := cosgo.Config.GetString(FlagsNameJson)
	//写盘失败必须计入 ok:磁盘满/目标只读时不能让 CI 拿旧产物绿灯通过
	if filepath.Ext(path) == ".json" {
		if err := WriteFile(path, data); err != nil {
			errs = append(errs, err)
		}
	} else {
		for k, v := range data {
			file := filepath.Join(path, k+".json")
			if err := WriteFile(file, v); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return len(errs) == 0
}
