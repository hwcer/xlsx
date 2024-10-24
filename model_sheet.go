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

func (this GlobalDummy) Name(d *Dummy) string {
	label := d.Compile()
	if v, ok := this[label]; ok && v.Name != "" {
		return v.Name
	} else {
		return label
	}
}

// Exist 查找名字是否存在
func (this GlobalDummy) Exist(d *Dummy) *Dummy {
	if d.Name == "" {
		return nil
	}
	for _, v := range this {
		if d.Name == v.Name && d.Compile() != v.Compile() {
			return v
		}
	}
	return nil
}

func (this GlobalDummy) Insert(sheet *Sheet, d *Dummy, must ...bool) {
	label := d.Compile()
	if !Config.EnableGlobalDummyName && (len(must) == 0 || !must[0]) {
		d.Name = label
	}
	if e := globalObjects.Exist(d); e != nil {
		logger.Trace("子对象名重复并且属性不一样:%v.%v  Label:%v", sheet.ProtoName, d.Name, d.Compile())
		d.Name = label
	}
	if d.Name == "" {
		d.Name = Config.ProtoNameFilter(SheetTypeHash, label)
	} else {
		d.Name = Config.ProtoNameFilter(SheetTypeHash, d.Name)
	}
	v, ok := this[label]
	if !ok {
		this[label] = d
		return
	}

	if v.Name == "" {
		v.Name = d.Name
	} else if d.Name != "" && v.Name != d.Name {
		logger.Trace("冗余的对象名称%v,建议修改成%v", d.Name, v.Name)
	}
}

type Sheet struct {
	*xlsx.Sheet
	Skip       int       //数据表中数据部分需要跳过的行数
	Parser     Parser    //解析器
	Fields     []*Field  //字段列表
	FileName   string    //文件名
	ProtoName  string    // protoName 是pb.go中文件的名字，
	ProtoIndex int       //总表编号
	SheetType  SheetType //输出类型,kv arr map
	SheetIndex [4]int    //kv 模式下的字段
	DummyName  string    // array 模式下子对象名称
}

//const RowId = "id"

type rowArr struct {
	Coll []any
}

// 重新解析obj的字段
func (this *Sheet) reParseObjField() {
	maxRow := this.MaxRow
	var index int
	var fields []*Field
	indexes := this.SheetIndex
	//if p, ok := this.Parser.(ParserStructType); ok {
	//	indexes = p.StructType(this.ProtoName)
	//}
	for i := this.Skip; i <= maxRow; i++ {
		row, err := this.Sheet.Row(i)
		if err != nil {
			logger.Trace("%v,err:%v", i, err)
		}
		key := strings.TrimSpace(row.GetCell(indexes[0]).Value)
		if utils.Empty(key) {
			continue
		}

		index++
		field := &Field{}
		field.Name = key
		field.Index = []int{indexes[1]}
		//field.ProtoName = key
		//
		//
		field.ProtoIndex = index
		//field.ProtoRequire = FieldTypeNone
		if indexes[2] >= 0 {
			if v := strings.TrimSpace(row.GetCell(indexes[2]).Value); v != "" {
				field.ProtoType = ProtoBuffTypeFormat(v)
			}
		}
		if field.ProtoType == "" {
			field.ProtoType = ProtoBuffTypeFormat("int")
		}
		if indexes[3] >= 0 {
			if v := strings.TrimSpace(row.GetCell(indexes[3]).Value); v != "" {
				field.ProtoDesc = v
			}
		}
		fields = append(fields, field)
	}
	this.Fields = fields
}

func (this *Sheet) GetField(name string) *Field {
	for _, v := range this.Fields {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (this *Sheet) Values() (any, []error) {
	if this.SheetType == SheetTypeEnum {
		return this.kv()
	} else if this.SheetType == SheetTypeArray {
		return this.array()
	} else {
		return this.hash()
	}
}

// kv 模式
func (this *Sheet) kv() (any, []error) {
	r := map[string]any{}
	var errs []error
	var emptyCell []int
	maxRow := this.Sheet.MaxRow
	indexes := this.SheetIndex
	//if p, ok := this.Parser.(ParserStructType); ok {
	//	indexes = p.StructType(this.ProtoName)
	//}
	for i := this.Skip; i <= maxRow; i++ {
		row, err := this.Sheet.Row(i)
		if err != nil {
			logger.Trace("%v,err:%v", i, err)
		}

		id := strings.TrimSpace(row.GetCell(indexes[0]).Value)
		if utils.Empty(id) {
			emptyCell = append(emptyCell, row.GetCoordinate()+1)
			continue
		}
		if field := this.GetField(id); field != nil {
			var data any
			if data, err = field.Value(row); err == nil {
				r[id] = data
			} else {
				errs = append(errs, fmt.Errorf("解析错误:%v第%v行,%v", this.ProtoName, row.GetCoordinate()+1, err))
			}
		}
	}

	if len(emptyCell) > 10 {
		logger.Trace("%v共%v行ID为空已经忽略", this.ProtoName, len(emptyCell))
	}
	return r, errs
}

func (this *Sheet) hash() (any, []error) {
	r := map[string]any{}
	var errs []error
	var emptyCell []int
	maxRow := this.Sheet.MaxRow
	for i := this.Skip; i <= maxRow; i++ {
		row, err := this.Sheet.Row(i)
		if err != nil {
			logger.Trace("%v,err:%v", i, err)
		}

		id := strings.TrimSpace(row.GetCell(0).Value)
		if utils.Empty(id) {
			emptyCell = append(emptyCell, row.GetCoordinate()+1)
			continue
		}
		val, err := this.Value(row)
		if err != nil {
			errs = append(errs, fmt.Errorf("解析错误:%v第%v行,%v", this.ProtoName, row.GetCoordinate()+1, err))
			continue
		}
		r[id] = val
	}

	if len(emptyCell) > 10 {
		//logger.Trace("%v共%v行ID为空已经忽略:%v", this.ProtoName, len(emptyCell), emptyCell)
	}
	return r, errs
}

func (this *Sheet) array() (any, []error) {
	r := map[string]*rowArr{}
	var errs []error
	var emptyCell []int
	maxRow := this.Sheet.MaxRow
	for i := this.Skip; i <= maxRow; i++ {
		row, err := this.Sheet.Row(i)
		if err != nil {
			logger.Trace("%v,err:%v", i, err)
		}

		id := strings.TrimSpace(row.GetCell(0).Value)
		if utils.Empty(id) {
			emptyCell = append(emptyCell, row.GetCoordinate()+1)
			continue
		}
		//MAP ARRAY
		val, err := this.Value(row)
		if err != nil {
			errs = append(errs, fmt.Errorf("解析错误:%v第%v行,%v", this.ProtoName, row.GetCoordinate()+1, err))
			continue
		}
		if d, ok := r[id]; ok {
			d.Coll = append(d.Coll, val)
		} else {
			d = &rowArr{}
			d.Coll = append(d.Coll, val)
			r[id] = d
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
		v, e := field.GetBranch().Value(row)
		if e != nil {
			return nil, e
		} else {
			r[field.Name] = v
		}
	}
	return r, nil
}

// Language 找出所有多语言文本
func (this *Sheet) Language(r map[string]string, types map[string]bool) {
	var fields []*Field
	for _, v := range this.Fields {
		v = v.GetBranch()
		if h := Require(v.ProtoType); !h.Repeated() && len(v.Dummy) == 0 && len(v.Index) == 1 && types[v.FieldType] {
			fields = append(fields, v)
		}
	}
	maxRow := this.Sheet.MaxRow
	for i := this.Skip; i <= maxRow; i++ {
		row, err := this.Sheet.Row(i)
		if err != nil {
			logger.Trace("%v,err:%v", i, err)
		}
		id := strings.TrimSpace(row.GetCell(0).Value)
		if !utils.Empty(id) {
			for _, f := range fields {
				if c := row.GetCell(f.Index[0]); c != nil {
					k := fmt.Sprintf("%v_%v_%v", this.ProtoName, f.Name, id)
					r[k] = c.Value
				}
			}
		}
	}
}

// GlobalObjectsProtoName 通过ProtoName生成子对象
func (this *Sheet) GlobalObjectsProtoName() {
	for _, field := range this.Fields {
		if len(field.Dummy) > 0 {
			//t := field.Type()
			dummy := field.Dummy[0]
			globalObjects.Insert(this, dummy)
		}
	}
}

// GlobalObjectsAutoName 自动命名
//func (this *Sheet) GlobalObjectsAutoName() {
//	for _, field := range this.Fields {
//		if len(field.Dummy) > 0 {
//			dummy := field.Dummy[0]
//			if _, ok := globalObjects.Search(dummy); !ok {
//				globalObjects[dummy.Label] = dummy
//			}
//		}
//	}
//}
