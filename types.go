package xlsx

import (
	"strconv"
	"strings"
)

type ProtoBuffType string

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
	Register(ProtoBuffTypeInt32, &protoBuffTypeParseDefault{pt: ProtoBuffTypeInt32})
	Register(ProtoBuffTypeInt64, &protoBuffTypeParseDefault{pt: ProtoBuffTypeInt64})
	Register(ProtoBuffTypeUint32, &protoBuffTypeParseDefault{pt: ProtoBuffTypeUint32})
	Register(ProtoBuffTypeUint64, &protoBuffTypeParseDefault{pt: ProtoBuffTypeUint64})
	Register(ProtoBuffTypeFloat, &protoBuffTypeParseDefault{pt: ProtoBuffTypeFloat})
	Register(ProtoBuffTypeDouble, &protoBuffTypeParseDefault{pt: ProtoBuffTypeDouble})
	Register(ProtoBuffTypeBool, &protoBuffTypeParseDefault{pt: ProtoBuffTypeBool})
	Register(ProtoBuffTypeByte, &protoBuffTypeParseDefault{pt: ProtoBuffTypeByte})
	Register(ProtoBuffTypeString, &protoBuffTypeParseDefault{pt: ProtoBuffTypeString})
}

func Register(t ProtoBuffType, handle ProtoBuffParse) {
	t = ProtoBuffType(strings.ToLower(string(t)))
	protoBuffTypeParse[t] = handle
}

func Require(t ProtoBuffType) ProtoBuffParse {
	t = ProtoBuffType(strings.ToLower(string(t)))
	return protoBuffTypeParse[t]
}

type protoBuffTypeParseDefault struct {
	pt ProtoBuffType
}

func ProtoBuffTypeFormat(t string) ProtoBuffType {
	t = strings.TrimSpace(t)
	switch t {
	case "int", "int32":
		return "int32"
	case "int64":
		return "int64"
	case "float", "float32":
		return "float"
	case "float64", "double":
		return "double"
	case "str", "string", "text":
		return "string"
	}
	return ProtoBuffType(strings.ToLower(t))
}

func (this *protoBuffTypeParseDefault) Type() string {
	return string(this.pt)
}

func (this *protoBuffTypeParseDefault) Value(vs ...string) (r any, err error) {
	v := vs[0]
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

func (*protoBuffTypeParseDefault) Repeated() bool {
	return false
}

func (*protoBuffTypeParseDefault) trimInt(s string) string {
	if i := strings.Index(s, "."); i > 0 {
		s = s[0:i]
	}
	return s
}
