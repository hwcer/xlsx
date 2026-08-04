package xlsx

import (
	"fmt"
	"strings"

	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
)

// Field 基础字段
//
// Field.ProtoType 除proto基础数据类型外还可以自定义类型  array, arrInt,arrObj...
type Field struct {
	Name       string            //字段名字
	side       string            //字段归属标记(S:服务器,C:客户端),用于区分前后端字段
	Index      []int             //字段关联的CELL索引
	Dummy      []*Dummy          //子对象
	FieldType  string            //表格中定义的原始字段类型
	ProtoDesc  string            //备注信息
	ProtoType  ProtoBuffType     //PROTO字段类型,和SheetType有一定的关联性
	ProtoIndex int               //proto index 自动生产
	Branch     map[string]*Field //版本分支,仅影响数据，不影响结构,不支持子对象
	kvRow      int               //kv 模式下第几行产生的
}

func NewField(name, side string) *Field {
	return &Field{Name: name, side: side}
}

func (this *Field) Side(side ...string) string {
	if len(side) > 0 {
		this.side = side[0]
	}
	return this.side
}

func (this *Field) Type() string {
	if len(this.Dummy) > 0 {
		return this.Dummy[0].Name
	}
	if handle := Require(this.ProtoType); handle != nil {
		return handle.Type()
	}
	return string(this.ProtoType)
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
func (this *Field) Value(shell *Sheet, row []string) (ret any, err error) {
	handle := Require(this.ProtoType)
	if len(this.Dummy) > 0 {
		ret, err = this.getDummyValue(shell, row, handle)
	} else if handle != nil {
		ret, err = this.getProtoValue(shell, row, handle)
	} else {
		err = fmt.Errorf("无法识别的类型(%v)", this.Name)
	}
	if err != nil {
		//必须用 %w:Sheet 层要靠 errors.As 从这里取回 CellError 的列坐标
		err = fmt.Errorf("字段名:%v,错误信息:%w", this.Name, err)
	}
	return
}

// getProtoValue 基础和预定义类型
func (this *Field) getProtoValue(shell *Sheet, row []string, handle ProtoBuffParse) (any, error) {
	index := this.Index
	if len(index) == 0 {
		return nil, fmt.Errorf("字段名:%v,错误信息:%v", this.Name, "缺少有效的数据列")
	}
	var vs []string
	var cols []int //vs 各值来自哪一列,报错时用来定位
	for _, i := range index {
		if i < len(row) && !Config.Empty(row[i]) {
			vs = append(vs, row[i])
			cols = append(cols, i)
		} else {
			this.hasEmptyValue(shell, row)
		}
	}
	if len(vs) > 0 {
		//多列合并成一个值(数组类字段)时只能指到首列,已足够定位到出错的字段块
		v, err := handle.Value(vs...)
		return v, NewCellError(cols[0], err)
	} else if handle.Repeated() {
		return []any{}, nil //空数组
	} else {
		v, err := handle.Value("") //填充零值
		return v, NewCellError(index[0], err)
	}
}

// getDummyValue 内置对象
func (this *Field) getDummyValue(shell *Sheet, row []string, handle ProtoBuffParse) (any, error) {
	var rs []any
	for _, c := range this.Dummy {
		if v, e := c.Value(row); e != nil {
			return nil, e
		} else if v != nil {
			rs = append(rs, v)
		} else {
			this.hasEmptyValue(shell, row)
		}
	}
	if len(rs) > 0 {
		if handle.Repeated() {
			return rs, nil
		} else {
			return rs[0], nil
		}
	} else {
		if handle.Repeated() {
			return []any{}, nil
		} else {
			return nil, nil
		}
	}
}

func (this *Field) hasEmptyValue(shell *Sheet, row []string) {
	if cosgo.Config.GetBool(FlagsNameVerify) {
		id := ""
		if len(row) > 0 {
			id = row[0]
		}
		if id != "" {
			logger.Alert("空值警告 sheet:%s id:%s , field:%s", shell.SheetName, id, this.Name)
		}
	}
}
