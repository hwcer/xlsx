package main

import (
	"fmt"
	cosxls "github.com/hwcer/xlsx"
)

const (
	FieldTypeArray  cosxls.ProtoBuffType = "Array"
	FieldTypeObject                      = "Object"
	FieldTypeArrObj                      = "ArrObj"
)

func init() {
	cosxls.Register(FieldTypeArray, &Array{})
	cosxls.Register(FieldTypeObject, &Object{})
	cosxls.Register(FieldTypeArrObj, &ArrObj{})
}

type Array struct {
}

func (this *Array) Type() string {
	return "int32"
}

func (this *Array) Value(vs ...string) (any, error) {
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

func (this *Array) Repeated() bool {
	return true
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

type ArrObj struct {
}

func (this *ArrObj) Type() string {
	return string(FieldTypeArrObj)
}

func (this *ArrObj) Value(...string) (any, error) {
	return nil, fmt.Errorf("对象无法直接获取值")
}
func (this *ArrObj) Repeated() bool {
	return true
}
