package sample

import (
	"fmt"
	"strings"

	"github.com/hwcer/logger"
	cosxls "github.com/hwcer/xlsx"
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
// isEnd 检查字段解析是否结束，当所有嵌套标签都处理完成时返回true
func (this *Field) isEnd() bool {
	// 当flags数组为空时，表示所有嵌套结构（如[]、{}）都已处理完成
	return len(this.flags) == 0
}

// Compile 编译并判断字段是否合法
// 要求：1. 必须有字段名 2. 必须处理完所有嵌套标签 3. 对象类型必须有子对象且类型统一
func (this *Field) compile() bool {
	// 字段名不能为空
	if this.Name == "" {
		return false
	}
	// 非对象和数组对象类型只需检查标签是否处理完毕
	if !(this.ProtoType == FieldTypeObject || this.ProtoType == FieldTypeArrayObject) {
		return len(this.flags) == 0
	}
	// 对象类型必须有子对象
	if len(this.Dummy) == 0 {
		return false
	}

	var label string
	// 验证所有子对象类型是否统一
	for _, v := range this.Dummy {
		s := v.Compile()
		if label == "" {
			label = s
		} else if label != s {
			logger.Fatal("%v 子对象类型不统一:%v -- %v", this.Name, label, s)
		}
	}
	// 确保所有嵌套标签都已处理完毕
	return len(this.flags) == 0
}

// ending 处理字段结束符号，解析嵌套结构的结束标记
// 参数：
//
//	value: 单元格值
//	index: 单元格索引
//	suffix: 后缀字符串（包含结束符号）
//	protoType: 字段类型
//
// 返回：是否解析结束
func (this *Field) ending(value string, index int, suffix string, protoType cosxls.ProtoBuffType) bool {
	if suffix == "" {
		return this.isEnd()
	}
	var k []string // 收集子属性名
	var flag flags // 收集匹配的结束符号

	// 遍历后缀中的每个字符，处理结束符号
	for _, s := range suffix {
		c := fmt.Sprintf("%c", s)
		if v, has := this.flags.HasAndPop(c); !has {
			// 不是结束符号，作为子属性名的一部分
			k = append(k, v)
		} else {
			// 匹配到结束符号，添加到flag中
			flag = append(flag, v)
		}
	}

	// 非对象类型只需检查是否解析结束
	if !(this.ProtoType == FieldTypeObject || this.ProtoType == FieldTypeArrayObject) {
		return this.isEnd()
	}

	// 子对象属性不能为空
	if len(k) == 0 {
		logger.Fatal("子对象属性不能为空:%v", this.Name)
	}

	// 必须有子对象存在
	if len(this.Dummy) == 0 {
		logger.Fatal("错误的结束符号:%v", this.Name)
	}

	// 开始处理子属性
	id := strings.Join(k, "")
	dummy := this.Dummy[len(this.Dummy)-1] // 获取最后一个子对象
	if err := dummy.Add(id, protoType, index); err != nil {
		logger.Fatal(err)
	}

	// 返回是否完全解析结束
	return this.isEnd()
}

// Parse 解析字段值，处理各种嵌套结构类型
// 参数：
//
//	fieldType: 基础字段类型
//	value: 单元格值
//	index: 单元格索引
//
// 返回：是否解析结束
// 支持的嵌套结构：
//   - [{...}]: 数组对象类型
//   - [...]: 数组类型（根据基础类型确定具体数组类型）
//   - {...}: 对象类型
//   - <name>: 字段名声明格式
func (this *Field) parse(fieldType cosxls.ProtoBuffType, value string, index int) (end bool) {
	if fieldType == "" {
		return false
	}
	var protoType cosxls.ProtoBuffType
	//this.begin += 1
	this.Index = append(this.Index, index)
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
		} else if fieldType == cosxls.ProtoBuffTypeFloat {
			protoType = FieldTypeArrayFloat
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
	if !IsMultipleType(this.ProtoType) {
		return true
	}
	return this.ending(value, index, suffix, fieldType)
	//if !begin {
	//	return this.ending(cell)
	//}
	//fmt.Printf("发现ID:%v", suffix)
	//return false
}
