package xlsx

import (
	"fmt"

	"github.com/hwcer/cosgo/utils"
	"github.com/hwcer/logger"
	"github.com/xuri/excelize/v2"

	"strings"
)

type GlobalDummy map[string]*Dummy

var ignoreFiles []string
var globalObjects = GlobalDummy{}

func (this GlobalDummy) Insert(sheet *Sheet, d *Dummy, must ...bool) {
	label := d.Compile()
	if !Config.EnableGlobalDummyName && (len(must) == 0 || !must[0]) {
		d.Name = ""
	}
	if d.Name != "" {
		if v, ok := this[d.Name]; ok {
			if v.Compile() != label {
				logger.Trace("对象名重复:%v.%v  Label:%v  Target:%v", sheet.ProtoName, d.Name, label, v.Compile())
				d.Name = label
			}
		}
		this[d.Name] = d
	} else {
		d.Name = label
		if _, ok := this[label]; !ok {
			this[label] = d
		}
	}
}

type Sheet struct {
	//*xlsx.Sheet
	//Rows         [][]string

	Name         string                  //
	Skip         int                     //数据表中数据部分需要跳过的行数
	Parser       Parser                  //解析器
	Fields       []*Field                //字段列表
	FileName     string                  //文件名
	ProtoName    string                  // protoName 是pb.go中文件的名字，
	SheetType    SheetType               //输出类型,kv map
	SheetName    string                  //原名
	ProtoIndex   int                     //总表编号
	sheetAttach  map[string]*SheetAttach //枚举和索引
	sheetIndexes [4]int                  //kv 索引
	rows         [][]string              //所有行数据
	excel        *excelize.File
}

type SheetAttach struct {
	t SheetType
	k string
	v [4]int
}

//const RowId = "id"

//	type rowArr struct {
//		Coll []any
//	}

// GetRows 获取工作表的所有行数据
// 该方法会缓存行数据，避免重复读取Excel文件
// 返回：包含所有行的二维字符串数组，如果获取失败则返回nil
func (this *Sheet) GetRows() [][]string {
	var err error
	if this.rows == nil {
		if this.rows, err = this.excel.GetRows(this.SheetName); err != nil {
			logger.Trace("获取行数据失败:%v,err:%v", this.SheetName, err)
			return nil
		}
	}
	return this.rows
}

// GetRow 获取指定索引的行数据
// 参数：
//
//	index: 行索引（从0开始）
//
// 返回：指定行的字符串数组，如果索引超出范围则返回nil
func (this *Sheet) GetRow(index int) []string {
	rows := this.GetRows()
	if index >= len(rows) {
		return nil
	}
	return rows[index]
}

// MaxRow 获取工作表的最大行数
// 返回：工作表的总行数，如果没有行数据则返回0
func (this *Sheet) MaxRow() int {
	rows := this.GetRows()
	if len(rows) == 0 {
		return 0
	}
	return len(rows)
}

func (this *Sheet) Clone() *Sheet {
	r := *this
	r.Fields = nil
	r.sheetAttach = nil
	return &r
}
func (this *Sheet) SearchByTag(i int) *Field {
	for _, f := range this.Fields {
		if f.tag == i {
			return f
		}
	}
	return nil
}

func (this *Sheet) SearchByIndex(i int) *Field {
	for _, f := range this.Fields {
		for _, k := range f.Index {
			if k == i {
				return f
			}
		}
	}
	return nil
}

// AddEnum 创建枚举
func (this *Sheet) AddEnum(k string, v [4]int) error {
	if this.sheetAttach == nil {
		this.sheetAttach = map[string]*SheetAttach{}
	}
	k = TrimProtoName(k)
	if _, ok := this.sheetAttach[k]; ok {
		return fmt.Errorf("attach已经存在,sheet:%v,k:%v", this.ProtoName, k)
	}

	this.sheetAttach[k] = &SheetAttach{k: k, v: v, t: SheetTypeEnum}
	return nil
}

// AddIndex 创建索引 [2]int{"索引字段","索引值"}    map[i][]int32{id,id,id}
//func (this *Sheet) AddIndex(k string, v [2]int) error {
//	if _, ok := this.sheetAttach[k]; ok {
//		return fmt.Errorf("attach已经存在,sheet:%v,k:%v", this.ProtoName, k)
//	}
//	av := [4]int{v[0], v[1], -1, -1}
//	this.sheetAttach[k] = &SheetAttach{t: SheetTypeArray, k: k, v: av}
//	return nil
//}

// 重新解析obj的字段
func (this *Sheet) reParseEnum(attach *SheetAttach) *Sheet {
	rows := this.GetRows()
	if rows == nil {
		return nil
	}
	maxRow := this.MaxRow() - 1
	var index int
	var fields []*Field
	//indexes := this.SheetIndex
	//if p, ok := this.Parser.(ParserStructType); ok {
	//	indexes = p.StructType(this.ProtoName)
	//}
	attach.k = Convert(attach.k)
	newSheet := this.Clone()
	newSheet.ProtoName = attach.k
	newSheet.SheetName = attach.k
	newSheet.SheetType = SheetTypeEnum
	newSheet.sheetIndexes = attach.v
	indexes := attach.v

	var ok bool
	if newSheet.ProtoName, ok = VerifyName(newSheet.ProtoName); !ok {
		return nil
	}
	newSheet.ProtoName = TrimProtoName(newSheet.ProtoName)
	//newSheet.ProtoName = Config.ProtoNameFilter(newSheet, newSheet.ProtoName)

	for i := this.Skip; i <= maxRow; i++ {
		row := this.GetRow(i)
		if row == nil {
			break
		}
		key := ""
		if indexes[0] < len(row) {
			key = strings.TrimSpace(row[indexes[0]])
		}
		if utils.Empty(key) {
			continue
		}

		index++
		field := &Field{}
		field.tag = i
		field.Name = key
		field.Index = []int{indexes[1]}
		//field.ProtoName = key
		//
		//
		field.ProtoIndex = index
		//field.ProtoRequire = FieldTypeNone
		if indexes[2] >= 0 && indexes[2] < len(row) {
			if v := strings.TrimSpace(row[indexes[2]]); v != "" {
				field.ProtoType = ProtoBuffTypeFormat(v)
			}
		}
		if field.ProtoType == "" {
			field.ProtoType = ProtoBuffTypeFormat("int")
		}
		if indexes[3] >= 0 && indexes[3] < len(row) {
			if v := strings.TrimSpace(row[indexes[3]]); v != "" {
				field.ProtoDesc = v
			}
		}
		fields = append(fields, field)
	}
	newSheet.Fields = fields
	return newSheet
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
	switch this.SheetType {
	case SheetTypeEnum:
		return this.kv()
	default:
		return this.hash()
	}
}

// kv 模式
func (this *Sheet) kv() (any, []error) {
	r := map[string]any{}
	var errs []error
	//var emptyCell []int
	rows := this.GetRows()
	if rows == nil {
		return r, errs
	}
	maxRow := this.MaxRow() - 1
	//indexes := this.sheetIndexes
	//if p, ok := this.Parser.(ParserStructType); ok {
	//	indexes = p.StructType(this.ProtoName)
	//}
	for i := this.Skip; i <= maxRow; i++ {
		row := this.GetRow(i)
		if row == nil {
			break
		}
		if field := this.SearchByTag(i); field != nil {
			var data any
			var err error
			if data, err = field.Value(this, row); err == nil {
				r[field.Name] = data
			} else {
				errs = append(errs, fmt.Errorf("解析错误:%v第%v行,%v", this.ProtoName, i+1, err))
			}
		}
	}
	//if len(emptyCell) > 10 {
	//	logger.Trace("%v共%v行ID为空已经忽略", this.ProtoName, len(emptyCell))
	//}
	return r, errs
}

func (this *Sheet) hash() (any, []error) {
	r := map[string]any{}
	var errs []error
	var emptyCell []int
	rows := this.GetRows()
	if rows == nil {
		return r, errs
	}
	maxRow := this.MaxRow() - 1
	for i := this.Skip; i <= maxRow; i++ {
		row := this.GetRow(i)
		if row == nil {
			break
		}

		id := ""
		if len(row) > 0 {
			id = strings.TrimSpace(row[0])
		}
		if utils.Empty(id) {
			emptyCell = append(emptyCell, i+1)
			continue
		}
		val, err := this.Value(row)
		if err != nil {
			errs = append(errs, fmt.Errorf("解析错误:%v第%v行,%v", this.ProtoName, i+1, err))
			continue
		}
		r[id] = val
	}

	if len(emptyCell) > 10 {
		//logger.Trace("%v共%v行ID为空已经忽略:%v", this.ProtoName, len(emptyCell), emptyCell)
	}
	return r, errs
}

func (this *Sheet) Value(row []string) (map[string]any, error) {
	r := map[string]any{}
	for _, field := range this.Fields {
		v, e := field.GetBranch().Value(this, row)
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
	rows := this.GetRows()
	if rows == nil {
		return
	}
	maxRow := this.MaxRow() - 1
	for i := this.Skip; i <= maxRow; i++ {
		row := this.GetRow(i)
		if row == nil {
			break
		}
		id := ""
		if len(row) > 0 {
			id = strings.TrimSpace(row[0])
		}
		if !utils.Empty(id) {
			for _, f := range fields {
				if f.Index[0] < len(row) {
					k := fmt.Sprintf("%v_%v_%v", this.ProtoName, f.Name, id)
					r[k] = row[f.Index[0]]
				}
			}
		}
	}
}

// GlobalObjectsProtoName 通过ProtoName生成子对象
func (this *Sheet) GlobalObjectsProtoName() {
	for _, field := range this.Fields {
		if len(field.Dummy) > 0 {
			for _, dummy := range field.Dummy {
				globalObjects.Insert(this, dummy)
			}
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
