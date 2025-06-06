package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
	"os"
	"path/filepath"
	"strings"
)

// strlen 字符串宽度,多字节按2输出
func strlen(s string) (r int) {
	for _, v := range s {
		if len(string(v)) > 2 {
			r += 2
		} else {
			r += 1
		}
	}
	return
}

func repeat(s string, n int) string {
	l := strlen(s)
	if l >= n {
		return s
	}
	s = s + strings.Repeat(" ", n-l)
	return s
}

func writeExcelIndex(sheets []*Sheet) {
	logger.Trace("======================开始生成配置索引======================")
	//输出所有标签
	b := &strings.Builder{}
	//t.WriteString("\n//配置索引......\n")
	in := cosgo.Config.GetString(FlagsNameIn) + "/"
	for _, s := range sheets {
		b.WriteString(repeat(s.ProtoName, 30))
		b.WriteString(repeat(s.Name, 40))
		b.WriteString(fmt.Sprintf("%v\n", strings.TrimPrefix(s.FileName, in)))
	}
	f := filepath.Join(cosgo.Config.GetString(FlagsNameOut), "配置索引.txt")
	if err := os.WriteFile(f, []byte(b.String()), os.ModePerm); err != nil {
		logger.Fatal(err)
	}
	logger.Trace("配置索引文件:%v", f)
}

func writeProtoMessage(sheets []*Sheet) {
	logger.Trace("======================开始生成PROTO MESSAGE======================")
	//输出配置
	b := &strings.Builder{}
	ProtoTitle(b)
	b.WriteString("\n//全局对象......\n")
	if Config.Message != nil {
		b.WriteString(Config.Message())
		b.WriteString("\n")
	}
	buildGlobalObjects(b, sheets)

	b.WriteString("\n//数据对象......\n")
	ProtoMessage(sheets, b)
	file := filepath.Join(cosgo.Config.GetString(FlagsNameOut), Config.Proto)
	if err := os.WriteFile(file, []byte(b.String()), os.ModePerm); err != nil {
		logger.Fatal(err)
	}
	logger.Trace("Proto Message File:%v", file)
}

func buildGlobalObjects(b *strings.Builder, sheets []*Sheet) {
	for _, s := range sheets {
		s.GlobalObjectsProtoName()
	}
	//for _, s := range sheets {
	//	s.GlobalObjectsAutoName()
	//}
	for _, dummy := range globalObjects {
		//if dummy.Name == "" {
		//	dummy.Name = globalObjects.Name(dummy)
		//}
		ProtoDummy(dummy, b)
	}

}
