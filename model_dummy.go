package xlsx

import (
	"fmt"
	"github.com/tealeg/xlsx/v3"
	"sort"
	"strings"
)

type DummyField struct {
	Name       string
	label      string //NameType
	ProtoType  ProtoBuffType
	ProtoIndex int //最终索引ProtoIndex
	SheetIndex int //表格中的索引
}

func NewDummy() *Dummy {
	return &Dummy{}
}

type Dummy struct {
	Name  string
	Label string
	//Sheets []int
	Fields []*DummyField
}

func (this *Dummy) Get(name string) *DummyField {
	for _, v := range this.Fields {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (this *Dummy) Add(name string, protoType ProtoBuffType, sheetIndex int) error {
	name = strings.TrimSpace(name)
	if field := this.Get(name); field != nil {
		return fmt.Errorf("字段名重复:%v", name)
	}
	field := &DummyField{
		Name:       name,
		ProtoType:  protoType,
		SheetIndex: sheetIndex,
	}
	field.label = fmt.Sprintf("%v%v", FirstUpper(field.Name), FirstUpper(string(protoType)))
	this.Fields = append(this.Fields, field)
	//this.Sheets = append(this.Sheets, SheetIndex)
	return nil
}

// Compile 编译并返回全局唯一名字(标记)
func (this *Dummy) Compile() (string, error) {
	if this.Name != "" {
		return this.Name, nil
	}
	sort.Slice(this.Fields, this.less)
	var arr []string
	for i, v := range this.Fields {
		v.ProtoIndex = i + 1
		arr = append(arr, v.label)
	}
	this.Label = strings.Join(arr, "")
	this.Name = this.Label
	return this.Label, nil
}

// Value 填充对象值
func (this *Dummy) Value(row *xlsx.Row) (map[string]any, error) {
	r := map[string]any{}
	for _, field := range this.Fields {
		cell := row.GetCell(field.SheetIndex)
		handle := Require(field.ProtoType)
		if cell != nil && handle != nil {
			if v, err := handle.Value(cell.Value); err != nil {
				return nil, err
			} else {
				r[field.Name] = v
			}
		}
	}
	return r, nil
}

func (this *Dummy) less(i, j int) bool {
	return this.Fields[i].label < this.Fields[j].label
}
