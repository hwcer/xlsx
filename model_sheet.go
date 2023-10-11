package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo/utils"
	"github.com/hwcer/logger"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

type GlobalDummy map[string]*Dummy

var ignoreFiles []string
var globalObjects = GlobalDummy{}

// Search 查找可能兼容的对象
func (this *GlobalDummy) Search(d *Dummy) (r string, ok bool) {
	dict := *this
	for k, v := range dict {
		if v.Label == d.Label {
			return k, true
		}
	}
	return
}

type Sheet struct {
	Fields     []*Field    //字段列表
	FileName   string      //文件名
	SheetName  string      //表格名称
	SheetRows  *xlsx.Sheet //sheets
	SheetSkip  int         //数据表中数据部分需要跳过的行数
	SheetType  SheetType   //输出类型,kv arr map
	ProtoName  string      // protoName 是pb.go中文件的名字，
	ProtoIndex int         //总表编号
}

//const RowId = "id"

type rowArr struct {
	Coll []any
}

// 重新解析obj的字段
func (this *Sheet) reParseObjField() {
	max := this.SheetRows.MaxRow
	var index int
	var fields []*Field
	for i := this.SheetSkip; i <= max; i++ {
		row, err := this.SheetRows.Row(i)
		if err != nil {
			logger.Trace("%v,err:%v", i, err)
		}
		key := strings.TrimSpace(row.GetCell(0).Value)
		if utils.Empty(key) {
			continue
		}

		index++
		field := &Field{}
		field.Name = key
		field.Index = []int{1}
		field.ProtoName = key
		field.ProtoIndex = index
		field.ProtoRequire = FieldTypeNone
		if v := strings.TrimSpace(row.GetCell(2).Value); v != "" {
			field.ProtoType = FormatType(v)
		} else {
			field.ProtoType = FormatType("int")
		}
		if Config.Require != nil {
			field.ProtoRequire = Config.Require(field.ProtoType)
		}

		if v := strings.TrimSpace(row.GetCell(3).Value); v != "" {
			field.ProtoDesc = v
		}
		fields = append(fields, field)
	}
	this.Fields = fields
}

func (this *Sheet) GetField(name string) *Field {
	for _, v := range this.Fields {
		if v.ProtoName == name {
			return v
		}
	}
	return nil
}

func (this *Sheet) Values() (any, []error) {
	r := map[string]any{}
	var errs []error
	var emptyCell []int
	max := this.SheetRows.MaxRow
	for i := this.SheetSkip; i <= max; i++ {
		row, err := this.SheetRows.Row(i)
		if err != nil {
			logger.Trace("%v,err:%v", i, err)
		}

		id := strings.TrimSpace(row.GetCell(0).Value)
		if utils.Empty(id) {
			emptyCell = append(emptyCell, row.GetCoordinate()+1)
			continue
		}
		//KV 模式直接定位 0,1 列
		if this.SheetType == SheetTypeObj {
			if field := this.GetField(id); field != nil {
				var data any
				if data, err = field.Value(row); err == nil {
					r[id] = data
				} else {
					errs = append(errs, fmt.Errorf("解析错误:%v第%v行,%v", this.ProtoName, row.GetCoordinate()+1, err))
				}
			}
			continue
		}
		//MAP ARRAY
		val, err := this.Value(row)
		if err != nil {
			errs = append(errs, fmt.Errorf("解析错误:%v第%v行,%v", this.ProtoName, row.GetCoordinate()+1, err))
			continue
		}
		//TODO
		if this.SheetType == SheetTypeArr {
			if d, ok := r[id]; !ok {
				d2 := &rowArr{}
				d2.Coll = append(d2.Coll, val)
				r[id] = d2
			} else {
				d2, _ := d.(*rowArr)
				d2.Coll = append(d2.Coll, val)
			}
		} else {
			r[id] = val
		}
	}

	if len(emptyCell) > 10 {
		logger.Trace("%v共%v行ID为空已经忽略:%v", this.ProtoName, len(emptyCell), emptyCell)
	}
	return r, errs
}

func (this *Sheet) Value(row *xlsx.Row) (map[string]any, error) {
	r := map[string]any{}
	for _, field := range this.Fields {
		v, e := field.Value(row)
		if e != nil {
			return nil, e
		} else {
			r[field.ProtoName] = v
		}
	}
	return r, nil
}
//
//// GlobalObjectsProtoName 通过ProtoName生成对象
//func (this *Sheet) GlobalObjectsProtoName() {
//	for _, field := range this.Fields {
//		if (field.ProtoRequire == FieldTypeObject || field.ProtoRequire == FieldTypeArrObj) && field.ProtoName != "" {
//			name := field.ProtoName
//			dummy := field.Dummy[0]
//			if k, ok := globalObjects.Search(dummy); ok {
//				field.ProtoType = k
//				if name != k {
//					logger.Trace("冗余的对象名称%v.%v,建议修改成%v", this.ProtoName, name, k)
//				}
//			} else {
//				field.ProtoType = name
//				globalObjects[name] = dummy
//			}
//		}
//	}
//}
//
//// GlobalObjectsAutoName 自动命名
//func (this *Sheet) GlobalObjectsAutoName() {
//	for _, field := range this.Fields {
//		if (field.ProtoRequire == FieldTypeObject || field.ProtoRequire == FieldTypeArrObj) && field.ProtoName == "" {
//			dummy := field.Dummy[0]
//			if k, ok := globalObjects.Search(dummy); ok {
//				field.ProtoType = k
//			} else {
//				field.ProtoType = dummy.Label
//				globalObjects[dummy.Label] = dummy
//			}
//		}
//	}
//}
//
//func buildGlobalObjects(b *strings.Builder, sheets []*Sheet) {
//	for _, s := range sheets {
//		s.GlobalObjectsProtoName()
//	}
//	for _, s := range sheets {
//		s.GlobalObjectsAutoName()
//	}
//	for k, dummy := range globalObjects {
//		dummy.Name = k
//		ProtoDummy(dummy, b)
//	}
//
//	globalObjects = map[string]*Dummy{}
//}
