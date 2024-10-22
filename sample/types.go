package sample

import (
	"fmt"
	cosxls "github.com/hwcer/xlsx"
	"strconv"
	"strings"
)

const (
	FieldTypeObject           cosxls.ProtoBuffType = "Object"
	FieldTypeArrayInt                              = "ArrInt"
	FieldTypeArrayInt64                            = "ArrInt64"
	FieldTypeArrayString                           = "ArrString"
	FieldTypeArrayObject                           = "ArrObject"
	FieldTypeArrayIntSplit                         = "[]int" //单元格切割成数组
	FieldTypeArrayInt32Split                       = "[]int32"
	FieldTypeArrayInt64Split                       = "[]int64"
	FieldTypeArrayStringSplit                      = "[]string"
)

var multiple = map[cosxls.ProtoBuffType]bool{}

func init() {
	cosxls.Register(FieldTypeObject, &Object{})
	cosxls.Register(FieldTypeArrayInt, &ArrayInt{})
	cosxls.Register(FieldTypeArrayInt64, &ArrayInt64{})
	cosxls.Register(FieldTypeArrayString, &ArrayString{})
	cosxls.Register(FieldTypeArrayObject, &ArrayObject{})

	cosxls.Register(FieldTypeArrayIntSplit, &ArrayFromSplit{t: cosxls.ProtoBuffTypeInt32})
	cosxls.Register(FieldTypeArrayInt32Split, &ArrayFromSplit{t: cosxls.ProtoBuffTypeInt32})
	cosxls.Register(FieldTypeArrayInt64Split, &ArrayFromSplit{t: cosxls.ProtoBuffTypeInt64})
	cosxls.Register(FieldTypeArrayStringSplit, &ArrayFromSplit{t: cosxls.ProtoBuffTypeString})

	multiple[FieldTypeObject] = true
	multiple[FieldTypeArrayInt] = true
	multiple[FieldTypeArrayInt64] = true
	multiple[FieldTypeArrayString] = true
	multiple[FieldTypeArrayObject] = true
	//multiple[FieldTypeArrayIntSplit] = false
	//multiple[FieldTypeArrayInt32Split] = false
	//multiple[FieldTypeArrayInt64Split] = false
	//multiple[FieldTypeArrayStringSplit] = false
}

// IsMultipleType 取基础值类型
func IsMultipleType(k string) bool {
	t := cosxls.ProtoBuffType(strings.ToLower(k))
	if p := cosxls.Require(t); p != nil {
		return false
	}
	return multiple[t]
}

type Object struct {
}

func (this *Object) Type() string {
	return string(FieldTypeObject)
}

func (this *Object) Value(...string) (any, error) {
	return nil, fmt.Errorf("对象无法直接获取值")
}
func (this *Object) Repeated() bool {
	return false
}

//-----------------------ArrayFromSplit------------------------------

type ArrayFromSplit struct {
	t cosxls.ProtoBuffType
}

func (this *ArrayFromSplit) parse(v string) (any, error) {
	if this.t == cosxls.ProtoBuffTypeString {
		return v, nil
	}
	i, e := strconv.Atoi(v)
	if e != nil {
		return nil, e
	}
	switch this.t {
	case cosxls.ProtoBuffTypeInt32:
		return int32(i), nil
	case cosxls.ProtoBuffTypeInt64:
		return int64(i), nil
	default:
		return 0, fmt.Errorf("未知的类型:%v", v)
	}
}

func (this *ArrayFromSplit) Type() string {
	return string(this.t)
}

func (this *ArrayFromSplit) Value(vs ...string) (any, error) {
	var r []any
	if vs[0] == "" {
		return r, nil
	}
	for _, v := range strings.Split(vs[0], ",") {
		if i, e := this.parse(v); e != nil {
			return nil, e
		} else {
			r = append(r, i)
		}
	}
	return r, nil
}

func (this *ArrayFromSplit) Repeated() bool {
	return true
}

//-----------------------ArrayInt------------------------------

type ArrayInt struct {
}

func (this *ArrayInt) Type() string {
	return "int32"
}

func (this *ArrayInt) Value(vs ...string) (any, error) {
	var r []any
	parser := cosxls.Require(cosxls.ProtoBuffTypeInt64)
	for _, i := range vs {
		if v, e := parser.Value(i); e != nil {
			return nil, e
		} else {
			r = append(r, v)
		}
	}
	return r, nil
}

func (this *ArrayInt) Repeated() bool {
	return true
}

//-----------------------ArrayInt------------------------------

type ArrayInt64 struct {
}

func (this *ArrayInt64) Type() string {
	return "int64"
}

func (this *ArrayInt64) Value(vs ...string) (any, error) {
	var r []any
	parser := cosxls.Require(cosxls.ProtoBuffTypeInt32)
	for _, i := range vs {
		if v, e := parser.Value(i); e != nil {
			return nil, e
		} else {
			r = append(r, v)
		}
	}
	return r, nil
}

func (this *ArrayInt64) Repeated() bool {
	return true
}

//-----------------------ArrayString------------------------------

type ArrayString struct {
}

func (this *ArrayString) Type() string {
	return "string"
}

func (this *ArrayString) Value(vs ...string) (any, error) {
	return vs, nil
}

func (this *ArrayString) Repeated() bool {
	return true
}

//-----------------------ArrayObject------------------------------

type ArrayObject struct {
}

func (this *ArrayObject) Type() string {
	return string(FieldTypeArrayObject)
}

func (this *ArrayObject) Value(...string) (any, error) {
	return nil, fmt.Errorf("对象无法直接获取值")
}
func (this *ArrayObject) Repeated() bool {
	return true
}
