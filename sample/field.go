package sample

import (
	"fmt"
	"github.com/hwcer/logger"
	cosxls "github.com/hwcer/xlsx"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

type flags []string

func (this *flags) Has(s string) bool {
	for _, v := range *this {
		if s == v {
			return true
		}
	}
	return false
}

// HasAndPop 如果结尾存在s则将s弹出
func (this *flags) HasAndPop(s string) (r string, has bool) {
	r = s
	l := len(*this)
	if l == 0 {
		return
	}
	if v := (*this)[l-1]; v == s {
		has = true
		*this = (*this)[0 : l-1]
	}
	return
}

type Field struct {
	cosxls.Field
	flags flags
}

// ------------------------------------------------------------------
func (this *Field) isEnd() bool {
	return len(this.flags) == 0
}

// Compile 编译并判断是否合法,必须处理完所有标签和子对象
func (this *Field) compile() bool {
	if this.Name == "" {
		return false
	}
	if !(this.ProtoType == FieldTypeObject || this.ProtoType == FieldTypeArrayObject) {
		return len(this.flags) == 0
	}
	if len(this.Dummy) == 0 {
		return false
	}

	var label string
	for _, v := range this.Dummy {
		s := v.Compile()
		if label == "" {
			label = s
		} else if label != s {
			logger.Fatal("%v 子对象类型不统一:%v -- %v", this.Name, label, s)
		}
	}
	return len(this.flags) == 0
}

// 寻找结束符号
func (this *Field) ending(cell *xlsx.Cell, index int, suffix string, protoType cosxls.ProtoBuffType) bool {
	if suffix == "" {
		return this.isEnd()
	}
	var k []string
	var flag flags
	//var end bool
	for _, s := range suffix {
		c := fmt.Sprintf("%c", s)
		if v, has := this.flags.HasAndPop(c); !has {
			k = append(k, v)
		} else {
			flag = append(flag, v)
		}
	}
	if !(this.ProtoType == FieldTypeObject || this.ProtoType == FieldTypeArrayObject) {
		return this.isEnd()
	}
	if len(k) == 0 {
		logger.Fatal("子对象属性不能为空:%v", this.Name)
	}
	if len(this.Dummy) == 0 {
		logger.Fatal("错误的结束符号:%v", this.Name)
	}
	//开始子属性
	id := strings.Join(k, "")
	dummy := this.Dummy[len(this.Dummy)-1]
	if err := dummy.Add(id, protoType, index); err != nil {
		logger.Fatal(err)
	}
	//fmt.Printf("发现子属性:%v %v\n", this.Name, strings.Join(k, ""))
	//统计子对象属性

	return this.isEnd()
}

// Parse [{   [[  {  [
// protoType
func (this *Field) parse(fieldType cosxls.ProtoBuffType, cell *xlsx.Cell, index int) (end bool) {
	if fieldType == "" {
		return false
	}
	var protoType cosxls.ProtoBuffType
	//this.begin += 1
	this.Index = append(this.Index, index)
	value := cell.Value
	if value == "" {
		return len(this.flags) == 0 //TODO 只有ARRAY允许为空
	}
	//var protoName string
	var dummyName string
	if i, j := strings.Index(value, "<"), strings.Index(value, ">"); i >= 0 && j >= 0 {
		this.Name = cosxls.FirstUpper(value[i+1 : j])
		dummyName = value[i+1 : j]
		value = value[j+1:]

	}
	//begin := false //不能在同一个单元格内同时开始和结束
	name, suffix := "", ""
	if i := strings.Index(value, "[{"); i >= 0 {
		//begin = true
		name = value[0:i]
		suffix = value[i+2:]
		this.flags = append(this.flags, "]", "}")
		this.Dummy = append(this.Dummy, cosxls.NewDummy(dummyName))
		protoType = FieldTypeArrayObject
	} else if i = strings.Index(value, "["); i >= 0 {
		//begin = true
		name = value[0:i]
		//suffix = value[i:]
		this.flags = append(this.flags, "]")
		if fieldType == cosxls.ProtoBuffTypeString {
			protoType = FieldTypeArrayString
		} else if fieldType == cosxls.ProtoBuffTypeInt32 {
			protoType = FieldTypeArrayInt
		} else if fieldType == cosxls.ProtoBuffTypeInt64 {
			protoType = FieldTypeArrayInt64
		}

	} else if i = strings.Index(value, "{"); i >= 0 {
		//begin = true
		name = value[0:i]
		suffix = value[i+1:]
		this.flags = append(this.flags, "}")
		this.Dummy = append(this.Dummy, cosxls.NewDummy(dummyName))
		protoType = FieldTypeObject
	} else {
		name = value
		suffix = value
		protoType = fieldType
	}

	//第一个名字和类型为准
	if len(this.Index) == 1 {
		this.Name = name
		this.ProtoType = protoType
	}
	if !multiple[this.ProtoType] {
		return true
	}
	return this.ending(cell, index, suffix, fieldType)
	//if !begin {
	//	return this.ending(cell)
	//}
	//fmt.Printf("发现ID:%v", suffix)
	//return false
}
