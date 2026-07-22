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

func (this GlobalDummy) Insert(sheet *Sheet, d *Dummy) {
	label := d.Compile()
	if !Config.EnableGlobalDummyName && d.Name == "" {
		logger.Fatal("系统设置不允许自动生成子对象名称,必须显式指定子对象名称: %v.%v", sheet.ProtoName, label)
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
	Name         string
	Skip         int                     //数据表中数据部分需要跳过的行数
	Parser       Parser                  //解析器
	Fields       []*Field                //字段列表
	FileName     string                  //文件名
	side         string                  //表名归属标记(S:服务器,C:客户端),用于区分前后端
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

func (this *Sheet) GetRows() [][]string {
	var err error
	if this.rows == nil {
		if this.rows, err = this.excel.GetRows(this.SheetName); err != nil {
			logger.Trace("获取行数据失败:%v,err:%v", this.SheetName, err)
			return nil
		}
		this.calcFormula() //补算缺少缓存值的公式单元格,见 formula.go
	}
	return this.rows
}

func (this *Sheet) GetRow(index int) []string {
	rows := this.GetRows()
	if index >= len(rows) {
		return nil
	}
	return rows[index]
}

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
		if f.kvRow == i {
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
	_, k = TrimProtoName(k)
	if _, ok := this.sheetAttach[k]; ok {
		return fmt.Errorf("attach已经存在,sheet:%v,k:%v", this.ProtoName, k)
	}

	this.sheetAttach[k] = &SheetAttach{k: k, v: v, t: SheetTypeEnum}
	return nil
}

func (this *Sheet) reParseEnum(attach *SheetAttach) *Sheet {
	rows := this.GetRows()
	if rows == nil {
		return nil
	}
	maxRow := this.MaxRow() - 1
	var index int
	var fields []*Field
	attach.k = Convert(attach.k)
	newSheet := this.Clone()
	newSheet.ProtoName = attach.k
	newSheet.SheetName = attach.k
	newSheet.SheetType = SheetTypeEnum
	newSheet.sheetIndexes = attach.v
	indexes := attach.v

	newSheet.side, newSheet.ProtoName = TrimProtoName(newSheet.ProtoName)
	newSheet.SheetName = newSheet.ProtoName
	if !VerifyTag(newSheet.side) {
		return nil
	}
	for i := this.Skip; i <= maxRow; i++ {
		row := this.GetRow(i)
		if row == nil {
			continue
		}
		key := ""
		if indexes[0] < len(row) {
			key = strings.TrimSpace(row[indexes[0]])
		}
		if utils.Empty(key) {
			continue
		}

		index++
		field := NewField(key, "")
		field.kvRow = i
		field.Index = []int{indexes[1]}
		field.ProtoIndex = index
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
	rows := this.GetRows()
	if rows == nil {
		return r, errs
	}
	maxRow := this.MaxRow() - 1
	for i := this.Skip; i <= maxRow; i++ {
		row := this.GetRow(i)
		if len(row) == 0 {
			continue
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
			continue
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
		logger.Trace("%v共%v行ID为空已经忽略", this.ProtoName, len(emptyCell))
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
			continue
		}
		id := ""
		if len(row) > 0 {
			id = strings.TrimSpace(row[0])
		}
		if !utils.Empty(id) {
			for _, f := range fields {
				if f.Index[0] < len(row) {
					k := fmt.Sprintf("%v_%v_%v", this.ProtoName, id, f.Name)
					r[k] = row[f.Index[0]]
				}
			}
		}
	}
}

// GlobalObjectsProtoName 通过ProtoName生成子对象
func (this *Sheet) GlobalObjectsProtoName() {
	for _, field := range this.Fields {
		if len(field.Dummy) == 0 {
			continue
		}
		var name string
		for _, dummy := range field.Dummy {
			if dummy.Name != "" {
				name = dummy.Name
				break
			}
		}
		for _, dummy := range field.Dummy {
			if dummy.Name == "" {
				dummy.Name = name
			}
		}
		if Config.NamedDummyInHeader && name != "" {
			continue
		}
		globalObjects.Insert(this, field.Dummy[0])
	}
}

