package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

// Field 基础字段
//
// Field.ProtoType 除proto基础数据类型外还可以自定义类型  array, arrInt,arrObj...
type Field struct {
	Name       string            //字段名字
	Index      []int             //字段关联的CELL索引
	Dummy      []*Dummy          //子对象
	FieldType  string            //表格中定义的原始字段类型
	ProtoDesc  string            //备注信息
	ProtoType  ProtoBuffType     //PROTO字段类型,和SheetType有一定的关联性
	ProtoIndex int               //proto index 自动生产
	Branch     map[string]*Field //版本分支,仅影响数据，不影响结构,不支持子对象
}

func (this *Field) Type() string {
	if len(this.Dummy) > 0 {
		return this.Dummy[0].Name
	} else if handle := Require(this.ProtoType); handle != nil {
		return handle.Type()
	} else {
		return string(this.ProtoType)
	}
}

func (this *Field) SetBranch(k string, v *Field) {
	if this.Branch == nil {
		this.Branch = make(map[string]*Field)
	}
	k = strings.ToUpper(k)
	this.Branch[k] = v
}
func (this *Field) GetBranch() *Field {
	f := this
	if len(f.Branch) == 0 {
		return f
	}
	if branch := strings.ToUpper(cosgo.Config.GetString(FlagsNameBranch)); branch != "" {
		if i, ok := this.Branch[branch]; ok {
			f = i
		}
	}
	return f
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
	index := this.Index
	if len(index) == 0 {
		return nil, fmt.Errorf("字段名:%v,错误信息:%v", this.Name, "缺少有效的数据列")
	}
	var vs []string
	for _, i := range index {
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
