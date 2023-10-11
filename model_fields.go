package xlsx

import (
	"errors"
	"github.com/tealeg/xlsx/v3"
)

type Field struct {
<<<<<<< HEAD
	Name         string
	Index        []int
	Dummy        []*Dummy     //子对象
	ProtoDesc    string       //备注信息
	ProtoName    string       //数据集表格中自定义的子对象名字
	ProtoType    string       //字段类型
	ProtoIndex   int          //生成的pb索引，自动填充
	ProtoRequire ProtoRequire //复杂类型,array,object  []object...
=======
	nodes  []*Field                                              //子对象
	valueIndex  int                                              //表格中的值所在列
	protoIndex  int                                              //生成的索引，自动填充
	Parser func(v string, t string, repeated bool) (any, error)  //格式化数据结构
	ProtoName     string //数据集表格中自定义的子对象名字
	ProtoType     string //字段类型
	ProtoDesc     string //备注信息
	ProtoRepeated bool   //数组类型
}
//NewField 新字段,i:字段名或者数据值所在的列
func NewField(i int)*Field  {
	return &Field{valueIndex: i}
}
// AddNode 添加子对象,必须为数组或者对象
func (this *Field) AddNode(node *Field) error {
	if IsProtoValue(this.ProtoType) || !this.ProtoRepeated {
		return errors.New("基础数据类型无法添加子对象")
	}
	this.nodes = append(this.nodes, node)
	return nil
>>>>>>> cabfa43f3ff1057a9154cc80e61d02d81319fa71
}

// Value 根据一行表格获取值
func (this *Field) Value(row *xlsx.Row) (ret any, err error) {
	if len(this.nodes) > 0 {
		if this.ProtoRepeated {
			//使用子对象的数组
			var r []any
			var v any
			for _, f := range this.nodes {
				if v, err = f.Value(row); err == nil {
					r = append(r, v)
				} else {
					return
				}
			}
			return r, nil
		} else {
			//使用子对象的对象
			var v any
			r := map[string]any{}
			for _, f := range this.nodes {
				if v, err = f.Value(row); err == nil {
					r[f.ProtoName] = v
				} else {
					return
				}
			}
			return r, nil
		}

	} else {
		cell := row.GetCell(this.valueIndex)
		if this.Parser != nil {
			return this.Parser(cell.Value, this.ProtoType, this.ProtoRepeated)
		} else {
			return FormatValue(cell.Value, this.ProtoType, this.ProtoRepeated)
		}
	}
}
