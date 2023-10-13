package xlsx

import (
	"fmt"
	"github.com/tealeg/xlsx/v3"
)

// Field 基础字段
//
// Field.ProtoType 除proto基础数据类型外还可以自定义类型  array, arrInt,arrObj...
type Field struct {
	Name      string   //字段名字
	Index     []int    //字段关联的CELL索引
	Dummy     []*Dummy //子对象
	ProtoDesc string   //备注信息
	//SheetType  string        //表格中定义的原始字段类型
	ProtoType  ProtoBuffType //PROTO字段类型,和SheetType有一定的关联性
	ProtoIndex int           //proto index 自动生产
}

func (this *Field) Type() string {
	if handle := Require(this.ProtoType); handle != nil {
		return handle.Type()
	} else if len(this.Dummy) > 0 {
		return this.Dummy[0].Name
	} else {
		return string(this.ProtoType)
	}
}

// Value 根据一行表格获取值
func (this *Field) Value(row *xlsx.Row) (ret any, err error) {
	handle := Require(this.ProtoType)
	if len(this.Dummy) > 0 {
		ret, err = this.getDummyValue(row, handle)
	} else if handle != nil {
		ret, err = this.getProtoValue(row, handle)
	} else {
		err = fmt.Errorf("无法识别的类型(%v)", this.Name)
	}
	if err != nil {
		err = fmt.Errorf("字段名:%v,错误信息:%v", this.Name, err)
	}
	return
}

// getProtoValue 基础和预定义类型
func (this *Field) getProtoValue(row *xlsx.Row, handle ProtoBuffParse) (any, error) {
	if len(this.Index) == 0 {
		return nil, fmt.Errorf("字段名:%v,错误信息:%v", this.Name, "缺少有效的数据列")
	}
	var vs []string
	for _, i := range this.Index {
		if c := row.GetCell(i); c != nil {
			vs = append(vs, c.Value)
		}
	}
	return handle.Value(vs...)
}

// getDummyValue 内置对象
func (this *Field) getDummyValue(row *xlsx.Row, handle ProtoBuffParse) (any, error) {
	if !handle.Repeated() {
		return this.Dummy[0].Value(row)
	}
	var rs []any
	for _, c := range this.Dummy {
		if v, e := c.Value(row); e != nil {
			return nil, e
		} else {
			rs = append(rs, v)
		}
	}
	return rs, nil
}
