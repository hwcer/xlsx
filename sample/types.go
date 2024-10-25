package sample

import (
	"fmt"
	cosxls "github.com/hwcer/xlsx"
	"strconv"
	"strings"
)

// 扩展类型
const (
	FieldTypeObject           cosxls.ProtoBuffType = "Object"
	FieldTypeArrayInt                              = "ArrInt"
	FieldTypeArrayInt32                            = "ArrInt32"
	FieldTypeArrayInt64                            = "ArrInt64"
	FieldTypeArrayString                           = "ArrString"
	FieldTypeArrayObject                           = "ArrObject"
	FieldTypeArrayIntSplit                         = "[]int"    //单元格切割成数组
	FieldTypeArrayInt32Split                       = "[]int32"  //单元格切割成数组
	FieldTypeArrayInt64Split                       = "[]int64"  //单元格切割成数组
	FieldTypeArrayStringSplit                      = "[]string" //单元格切割成数组
)

var multiple = map[cosxls.ProtoBuffType]struct{}{}

func init() {
	cosxls.Register(FieldTypeObject, &Object{})
	cosxls.Register(FieldTypeArrayInt, &ArrayFromMultipleCell{t: cosxls.ProtoBuffTypeInt32})
	cosxls.Register(FieldTypeArrayInt32, &ArrayFromMultipleCell{t: cosxls.ProtoBuffTypeInt32})
	cosxls.Register(FieldTypeArrayInt64, &ArrayFromMultipleCell{t: cosxls.ProtoBuffTypeInt64})
	cosxls.Register(FieldTypeArrayString, &ArrayFromMultipleCell{t: cosxls.ProtoBuffTypeString})
	cosxls.Register(FieldTypeArrayObject, &ArrayObject{})

	cosxls.Register(FieldTypeArrayIntSplit, &ArrayFromSplit{t: cosxls.ProtoBuffTypeInt32})
	cosxls.Register(FieldTypeArrayInt32Split, &ArrayFromSplit{t: cosxls.ProtoBuffTypeInt32})
	cosxls.Register(FieldTypeArrayInt64Split, &ArrayFromSplit{t: cosxls.ProtoBuffTypeInt64})
	cosxls.Register(FieldTypeArrayStringSplit, &ArrayFromSplit{t: cosxls.ProtoBuffTypeString})

	SetMultipleType(FieldTypeObject, FieldTypeArrayInt, FieldTypeArrayInt64, FieldTypeArrayString, FieldTypeArrayObject)
}

// IsMultipleType 取基础值类型
func IsMultipleType(k cosxls.ProtoBuffType) bool {
	_, ok := multiple[k]
	return ok
}

func SetMultipleType(keys ...cosxls.ProtoBuffType) {
	for _, k := range keys {
		multiple[k] = struct{}{}
	}
}

// -----------------------Object------------------------------

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
	if len(vs) == 0 || vs[0] == "" {
		return []any{}, nil
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

//-----------------------ArrayFromMultipleCell------------------------------

type ArrayFromMultipleCell struct {
	t cosxls.ProtoBuffType
}

func (this *ArrayFromMultipleCell) Type() string {
	return string(this.t)
}

func (this *ArrayFromMultipleCell) Value(vs ...string) (any, error) {
	var r []any
	if len(vs) == 0 {
		return []any{}, nil
	}
	parser := cosxls.Require(this.t)
	for _, i := range vs {
		if v, e := parser.Value(i); e != nil {
			return nil, e
		} else {
			r = append(r, v)
		}
	}
	return r, nil
}

func (this *ArrayFromMultipleCell) Repeated() bool {
	return true
}
