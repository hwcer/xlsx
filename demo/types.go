package main

import (
	"fmt"
	cosxls "github.com/hwcer/xlsx"
	"github.com/tealeg/xlsx/v3"
)

const (
	FieldTypeArray  cosxls.ProtoRequire = cosxls.FieldTypeArray
	FieldTypeObject                     = cosxls.FieldTypeObject
	FieldTypeArrObj                     = cosxls.FieldTypeArrObj
)

func init() {
	cosxls.Register(FieldTypeArray, &Array{})
	cosxls.Register(FieldTypeObject, &Object{})
	cosxls.Register(FieldTypeArrObj, &ArrObj{})
}

type Array struct {
}

func (this *Array) Value(field *cosxls.Field, row *xlsx.Row) (any, error) {
	var r []any
	for _, i := range field.Index {
		if v, err := cosxls.FormatValue(row, i, field.ProtoType); err == nil {
			r = append(r, v)
		} else {
			return nil, err
		}
	}
	return r, nil
}

func (this *Array) Require(field *cosxls.Field) string {
	return fmt.Sprintf("repeated %v", field.ProtoType)
}

type Object struct {
}

func (this *Object) Value(field *cosxls.Field, row *xlsx.Row) (any, error) {
	return field.Dummy[0].Value(row)
}

func (this *Object) Require(field *cosxls.Field) string {
	return fmt.Sprintf("%v", field.ProtoType)
}

type ArrObj struct {
}

func (this *ArrObj) Value(field *cosxls.Field, row *xlsx.Row) (any, error) {
	var r []any
	var v map[string]any
	var err error
	for _, dummy := range field.Dummy {
		if v, err = dummy.Value(row); err == nil && len(v) > 0 {
			r = append(r, v)
		} else if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (this *ArrObj) Require(field *cosxls.Field) string {
	return fmt.Sprintf("%v", field.ProtoType)
}
