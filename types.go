package xlsx

import (
	"strconv"
	"strings"
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
	//判断语言文件中有没有
	for _, k := range Config.Language {
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
	switch this.pt {
	case ProtoBuffTypeInt32, ProtoBuffTypeInt64:
		if v == "" {
			r = int64(0)
		} else {
			v = this.trimInt(v)
			r, err = strconv.Atoi(v)
		}
	case ProtoBuffTypeUint32, ProtoBuffTypeUint64:
		if v == "" {
			r = uint64(0)
		} else {
			v = this.trimInt(v)
			r, err = strconv.ParseUint(v, 10, 64)
		}
	case ProtoBuffTypeFloat, ProtoBuffTypeDouble:
		if v == "" {
			r = float64(0)
		} else {
			r, err = strconv.ParseFloat(v, 64)
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

func (*ProtoBuffParseDefault) trimInt(s string) string {
	s = Convert(s)
	if i := strings.Index(s, "."); i > 0 {
		s = s[0:i]
	}
	return s
}
