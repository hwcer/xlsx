package xlsx

import (
	"fmt"
	"github.com/tealeg/xlsx/v3"
)

type Field struct {
	Name         string
	Index        []int
	Dummy        []*Dummy     //子对象
	ProtoDesc    string       //备注信息
	ProtoName    string       //数据集表格中自定义的子对象名字
	ProtoType    string       //字段类型
	ProtoIndex   int          //生成的pb索引，自动填充
	ProtoRequire ProtoRequire //是否pb容器类型
}

// Value 根据一行表格获取值
func (this *Field) Value(row *xlsx.Row) (ret any, err error) {
	if len(this.Index) == 0 {
		return nil, fmt.Errorf("字段名:%v,错误信息:%v", this.Name, "缺少有效的数据列")
	}
	if handle, ok := protoRequireHandles[this.ProtoRequire]; ok {
		ret, err = handle.Value(this, row)
	} else {
		err = fmt.Errorf("字段名:%v,错误信息:%v", this.Name, err)
	}
	return
}
