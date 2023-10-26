package main

import (
	"fmt"
	cosxls "github.com/hwcer/xlsx"
)

const (
	FieldTypeObject      cosxls.ProtoBuffType = "Object"
	FieldTypeArrayInt                         = "[]int"
	FieldTypeArrayString                      = "[]string"
	FieldTypeArrayObject                      = "[]object"
)

func init() {
	cosxls.Register(FieldTypeObject, &Object{})
	cosxls.Register(FieldTypeArrayInt, &ArrayInt{})
	cosxls.Register(FieldTypeArrayString, &ArrayString{})
	cosxls.Register(FieldTypeArrayObject, &ArrayObject{})
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

//-----------------------ArrayInt------------------------------

type ArrayInt struct {
}

func (this *ArrayInt) Type() string {
	return "int32"
}

func (this *ArrayInt) Value(vs ...string) (any, error) {
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

func (this *ArrayInt) Repeated() bool {
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
