package sample

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hwcer/logger"
	cosxls "github.com/hwcer/xlsx"
)

func New(sheet *cosxls.Sheet) cosxls.Parser {
	return &Parser{sheet: sheet}
}

type Parser struct {
	sheet *cosxls.Sheet
}

func s2a(s string) [4]int {
	r := [4]int{-1, -1, -1}
	arr := strings.Split(s, ",")
	for i, v := range arr {
		if i < 4 {
			n, _ := strconv.Atoi(v)
			r[i] = n
		}
	}
	return r
}

// Verify 验证工作表格式是否合法，并获取工作表名称
// 返回：
//
//	skip: 跳过的行数
//	name: 工作表名称
//	ok: 是否验证通过
func (this *Parser) Verify() (skip int, name string, ok bool) {
	skip = 4 // 默认跳过4行

	// 获取第一行数据
	row := this.sheet.GetRow(0)
	if row == nil {
		fmt.Printf("获取sheet行错误 name:%v", this.sheet.Name)
		return
	}
	if len(row) == 0 {
		return
	}

	// 验证第一行第一个单元格是否有内容
	ok = row[0] != ""
	var err error

	// 获取工作表名称
	if ok {
		name = row[0]
	}

	// 处理附加信息（如kv模式配置）
	m := len(row) - 1
	for i := 1; i <= m; i++ {
		if i < len(row) && row[i] != "" {
			arr := strings.Split(row[i], ":")
			// 处理kv模式配置
			if strings.ToLower(arr[0]) == "kv" {
				if len(arr) == 1 {
					// 默认kv模式
					err = this.sheet.AddEnum(name, [4]int{0, 1, 2, 3})
				} else if len(arr) == 3 {
					// 自定义kv模式参数
					err = this.sheet.AddEnum(arr[1], s2a(arr[2]))
				} else {
					// 配置格式错误
					err = fmt.Errorf("attach error,sheet:%v,value:%v", this.sheet.Name, row[i])
				}
			}
			if err != nil {
				logger.Fatal(err)
			}
		}
	}
	return
}

// Fields 解析工作表中的所有字段定义
// 返回：解析成功的字段列表
func (this *Parser) Fields() (r []*cosxls.Field) {
	// 检查工作表是否至少有4行数据
	if this.sheet.MaxRow() < 4 {
		return
	}

	// 解析第二行的proto buff类型定义
	row1 := this.sheet.GetRow(1)
	if row1 == nil {
		return
	}
	protoType := map[int]cosxls.ProtoBuffType{} // 存储每个列的proto类型
	fieldType := map[int]string{}               // 存储每个列的字段类型
	maxCol := len(row1) - 1
	for j := 0; j <= maxCol; j++ {
		if j < len(row1) && row1[j] != "" {
			protoType[j] = cosxls.ProtoBuffTypeFormat(row1[j])
			fieldType[j] = strings.ToLower(strings.TrimSpace(row1[j]))
		}
	}

	// 解析第四行的字段描述
	row3 := this.sheet.GetRow(3)
	if row3 == nil {
		row3 = []string{}
	}
	sheetDesc := map[int]string{} // 存储每个列的描述信息
	if len(row3) > maxCol {
		maxCol = len(row3) - 1
	}
	for j := 0; j <= maxCol; j++ {
		if j < len(row3) {
			sheetDesc[j] = row3[j]
		}
	}

	// 解析第三行的字段名和嵌套结构
	var end bool
	var field = &Field{}
	row2 := this.sheet.GetRow(2)
	if row2 == nil {
		return
	}
	if len(row2) > maxCol {
		maxCol = len(row2) - 1
	}

	// 遍历所有列，解析字段定义
	for j := 0; j <= maxCol; j++ {
		// 设置字段描述
		if field.ProtoDesc == "" {
			field.ProtoDesc = sheetDesc[j]
		}
		// 设置字段类型
		if field.FieldType == "" {
			field.FieldType = fieldType[j]
		}
		// 获取当前列的proto类型
		pt, _ := protoType[j]
		if pt == "" {
			pt = field.ProtoType
		}

		// 获取当前列的字段值
		var value string
		if j < len(row2) {
			value = row2[j]
		}

		// 解析字段值，判断是否解析结束
		if end = field.parse(pt, value, j); end {
			// 编译字段，验证是否合法
			if field.compile() {
				r = append(r, &field.Field)
				field = &Field{} // 创建新的字段对象
			}
		}
	}
	return
}

// SheetType 获取工作表类型和索引配置
// 返回：
//
//	r: 工作表类型
//	index: 索引配置数组
func (this *Parser) SheetType() (r cosxls.SheetType, index [4]int) {
	row := this.sheet.GetRow(0)
	if row == nil {
		return
	}

	// 从第一行第二个单元格获取工作表类型
	if len(row) > 1 && row[1] != "" {
		r = cosxls.Config.GetType(row[1])
	}

	// 默认索引配置
	index = [4]int{0, 1, 2, 3}
	return
}

func (this *Parser) NewStruct() (r map[string][4]int) {
	r = map[string][4]int{}
	if strings.ToUpper(this.sheet.ProtoName) == "EMITTER" {
		r["events"] = [4]int{1, 0, 2, 7}
	}
	return
}

// StructType Struct表(kv模式)下解析方式
// key index
// val index
// type index type默认为int32
//func (this *Parser) StructType(protoName string) [4]int {
//	if name := strings.ToUpper(protoName); name == "EVENTS" {
//		return [4]int{1, 0, 2, 7}
//	}
//	return [4]int{0, 1, 2, 3} //默认值,仅仅演示
//}
