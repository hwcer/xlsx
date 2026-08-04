package xlsx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type ProtoBuffType string

func (pt ProtoBuffType) IsNumber() bool {
	return pt == ProtoBuffTypeInt32 || pt == ProtoBuffTypeInt64 || pt == ProtoBuffTypeUint32 || pt == ProtoBuffTypeUint64
}

type ProtoBuffParse interface {
	Type() string                   //Field.SheetType to protocol type
	Value(v ...string) (any, error) //protocol value
	Repeated() bool                 //proto是否Repeated
}

var protoBuffTypeParse = map[ProtoBuffType]ProtoBuffParse{}

const (
	ProtoBuffTypeInt32  ProtoBuffType = "int32"
	ProtoBuffTypeInt64  ProtoBuffType = "int64"
	ProtoBuffTypeUint32 ProtoBuffType = "uint32"
	ProtoBuffTypeUint64 ProtoBuffType = "uint64"
	ProtoBuffTypeFloat  ProtoBuffType = "float"
	ProtoBuffTypeDouble ProtoBuffType = "double"
	ProtoBuffTypeBool   ProtoBuffType = "bool"
	ProtoBuffTypeByte   ProtoBuffType = "byte"
	ProtoBuffTypeString ProtoBuffType = "string"
)

func init() {
	Register(ProtoBuffTypeInt32, &ProtoBuffParseDefault{pt: ProtoBuffTypeInt32})
	Register(ProtoBuffTypeInt64, &ProtoBuffParseDefault{pt: ProtoBuffTypeInt64})
	Register(ProtoBuffTypeUint32, &ProtoBuffParseDefault{pt: ProtoBuffTypeUint32})
	Register(ProtoBuffTypeUint64, &ProtoBuffParseDefault{pt: ProtoBuffTypeUint64})
	Register(ProtoBuffTypeFloat, &ProtoBuffParseDefault{pt: ProtoBuffTypeFloat})
	Register(ProtoBuffTypeDouble, &ProtoBuffParseDefault{pt: ProtoBuffTypeDouble})
	Register(ProtoBuffTypeBool, &ProtoBuffParseDefault{pt: ProtoBuffTypeBool})
	Register(ProtoBuffTypeByte, &ProtoBuffParseDefault{pt: ProtoBuffTypeByte})
	Register(ProtoBuffTypeString, &ProtoBuffParseDefault{pt: ProtoBuffTypeString})
}

func Register(t ProtoBuffType, handle ProtoBuffParse) {
	t = ProtoBuffType(strings.ToLower(string(t)))
	protoBuffTypeParse[t] = handle
}

func Require(t ProtoBuffType) ProtoBuffParse {
	t = ProtoBuffType(strings.ToLower(string(t)))
	return protoBuffTypeParse[t]
}

func ProtoBuffTypeFormat(t string) ProtoBuffType {
	t = Convert(t)
	t = strings.TrimSpace(t)
	switch t {
	case "int", "int32", "num", "number":
		return "int32"
	case "int64":
		return "int64"
	case "float", "float32":
		return "float"
	case "float64", "double":
		return "double"
	case "str", "string", "text", "lang", "language":
		return "string"
	}
	for _, k := range Config.LanguageTypes {
		if k == t {
			return "string"
		}
	}
	return ProtoBuffType(t)
}

func NewProtoBuffParse(t ProtoBuffType) *ProtoBuffParseDefault {
	return &ProtoBuffParseDefault{pt: t}
}

type ProtoBuffParseDefault struct {
	pt ProtoBuffType
}

func (this *ProtoBuffParseDefault) Type() string {
	return string(this.pt)
}

func (this *ProtoBuffParseDefault) Value(vs ...string) (r any, err error) {
	var v string
	if len(vs) > 0 {
		v = vs[0]
	}
	raw := v //原始单元格内容，仅用于报错，trimInt 后的值会丢失现场
	switch this.pt {
	case ProtoBuffTypeInt32, ProtoBuffTypeInt64:
		if v == "" {
			r = int64(0)
		} else {
			v = this.trimInt(v)
			r, err = strconv.Atoi(v)
			if err != nil {
				//Atoi 失败退回 ParseFloat：容忍 Excel 把整数存成 "300.0" / "3e2" 的浮点写法
				var f float64
				if f, err = strconv.ParseFloat(v, 64); err == nil {
					r = int(f)
				} else {
					err = this.parseError(raw, v)
				}
			}
		}
	case ProtoBuffTypeUint32, ProtoBuffTypeUint64:
		if v == "" {
			r = uint64(0)
		} else {
			v = this.trimInt(v)
			r, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				var f float64
				if f, err = strconv.ParseFloat(v, 64); err == nil {
					r = uint64(f)
				} else {
					err = this.parseError(raw, v)
				}
			}
		}
	case ProtoBuffTypeFloat, ProtoBuffTypeDouble:
		if v == "" {
			r = float64(0)
		} else {
			v = this.trimInt(v)
			if r, err = strconv.ParseFloat(v, 64); err != nil {
				err = this.parseError(raw, v)
			}
		}
	case ProtoBuffTypeBool:
		if s := strings.ToLower(strings.TrimSpace(v)); s == "" || s == "0" || s == "false" {
			r = false
		} else {
			r = true
		}
	case ProtoBuffTypeByte:
		r = []byte(v)
	case ProtoBuffTypeString:
		r = v
	}
	return
}

func (*ProtoBuffParseDefault) Repeated() bool {
	return false
}

// CellError 带列坐标的解析错误。
//
// 出错的列只有 Dummy/Field 这层知道（它们持有 SheetIndex），而行号只有 Sheet 那层知道，
// 单靠 fmt.Errorf 串起来的文本没法在最外层再补坐标。用这个类型把列一路带上去，
// Sheet.rowError 再配上行号，报成 Excel 里能直接定位的「第6行第AA列」。
type CellError struct {
	Column int //列索引,0 base
	Err    error
}

func (this *CellError) Error() string {
	return this.Err.Error()
}

func (this *CellError) Unwrap() error {
	return this.Err
}

// NewCellError 给错误补上列坐标；已带坐标的保留最内层(离出错单元格最近的那个)
func NewCellError(column int, err error) error {
	if err == nil {
		return nil
	}
	var ce *CellError
	if errors.As(err, &ce) {
		return err
	}
	return &CellError{Column: column, Err: err}
}

// ColumnName 列索引(0 base)转 Excel 列名:0->A,26->AA
func ColumnName(i int) string {
	if i < 0 {
		return "?"
	}
	var name string
	for n := i + 1; n > 0; n /= 26 {
		n--
		name = string(rune('A'+n%26)) + name
	}
	return name
}

// parseError 数值解析失败时的报错，必须带上原始单元格内容。
//
// trimInt 会滤掉所有非数字字符，单元格里若全是文字/全角空格之类，过滤后就成了空串，
// 底层只会报 `parsing ""`——看上去像"空值解析失败"，可空值在上面早已按 0 处理，
// 于是排查时人人都以为矛盾。把原始值一并带出来，才能指到具体是哪个单元格填错了。
func (this *ProtoBuffParseDefault) parseError(raw, trimmed string) error {
	if trimmed == "" {
		return fmt.Errorf("类型(%v)解析失败,单元格内容%q里没有任何数字字符", this.pt, raw)
	}
	return fmt.Errorf("类型(%v)解析失败,单元格内容%q(过滤后%q)", this.pt, raw, trimmed)
}

func (*ProtoBuffParseDefault) trimInt(s string) string {
	s = Convert(s)
	result := strings.Builder{}
	hasDigit := false
	for _, ch := range s {
		if unicode.IsDigit(ch) || (ch == '-' && !hasDigit) || ch == '.' || ch == 'e' || ch == 'E' || ch == '+' {
			result.WriteRune(ch)
			if unicode.IsDigit(ch) {
				hasDigit = true
			}
		}
	}
	return result.String()
}
