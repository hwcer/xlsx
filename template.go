package xlsx

import (
	"fmt"
	"github.com/hwcer/logger"
	"strings"
	"text/template"
)

var tpl *template.Template

const TemplateTitle = `syntax = "proto3";
option go_package = "./;<% .Package %>";
package <% .Package %>;
`

const TemplateDummy = `//sign:<% .Label%>
message <%.Name%>{ <%range .Fields%>
	<%DummyRequire .%> <%.Name%> = <%.ProtoIndex%>; <%end%>
}
`

const TemplateMessage = `
<%- range .Sheets%>
// Sheet:<%.Name%>
// File:<%.FileName%>
message <%.ProtoName%>{
	<%- range .Fields %>
	<%ProtoRequire .%> <%.Name%> = <%.ProtoIndex%>; //<% .ProtoDesc%><%end%>
}
<%- end%>
`

// TemplateSummary 输出一个总表
const TemplateSummary = `
message <%.Name%>{
<%- range .Sheets%>
	<%SummaryType .%>=<%.ProtoIndex%>;
<%- end %>
}
`

func init() {
	tpl = template.New("")
	tpl.Funcs(template.FuncMap{
		//"IsArray":      TemplateIsArray,
		"SummaryType":  TemplateSummaryType,
		"ProtoRequire": TemplateProtoRequire,
		"DummyRequire": TemplateDummyRequire,
	})
	tpl.Delims("<%", "%>")
}

//func TemplateIsArray(t SheetType) bool {
//	return t == SheetTypeArray
//}

func TemplateProtoRequire(field *Field) string {
	handle := Require(field.ProtoType)
	if handle == nil {
		return string(field.ProtoType)
	}
	protoType := field.Type()
	if handle.Repeated() {
		return fmt.Sprintf("%v %v", "repeated", protoType)
	} else {
		return protoType
	}
}

func TemplateDummyRequire(field *DummyField) string {
	handle := Require(field.ProtoType)
	protoType := string(field.ProtoType)
	if handle.Repeated() {
		return fmt.Sprintf("%v %v", "repeated", protoType)
	} else {
		return protoType
	}
}

func TemplateSummaryType(sheet *Sheet) (r string) {
	primary := sheet.Fields[0]
	//var t string
	switch sheet.SheetType {
	case SheetTypeEnum:
		return fmt.Sprintf("%v %v", sheet.ProtoName, sheet.ProtoName)
	//case SheetTypeArray:
	//	return fmt.Sprintf("map<int32,%v> %v", sheet.ProtoName, sheet.ProtoName)
	//return fmt.Sprintf("repeated %v", sheet.DummyName)
	default:
		//t = fmt.Sprintf("%v%v", sheet.ProtoName, Config.Suffix)
		return fmt.Sprintf("map<%v,%v> %v", primary.ProtoType, sheet.ProtoName, sheet.ProtoName)
	}
	//return fmt.Sprintf("map<%v,%v>", primary.ProtoType, sheet.ProtoName)
}

func ProtoTitle(builder *strings.Builder) {
	t, err := tpl.Parse(TemplateTitle)
	if err != nil {
		logger.Fatal(err)
	}
	data := &struct {
		Package string
	}{
		Package: Config.Package,
	}
	err = t.Execute(builder, data)
	if err != nil {
		logger.Fatal(err)
	}
	return
}
func ProtoDummy(dummy *Dummy, builder *strings.Builder) {
	t, err := tpl.Parse(TemplateDummy)
	if err != nil {
		logger.Fatal(err)
	}
	err = t.Execute(builder, dummy)
	if err != nil {
		logger.Fatal(err)
	}
	return
}

func ProtoMessage(sheets []*Sheet, builder *strings.Builder) {
	t, err := tpl.Parse(TemplateMessage)
	if err != nil {
		logger.Fatal(err)
	}

	data := &struct {
		//Suffix string
		Sheets []*Sheet
	}{
		//Suffix: Config.Suffix,
		Sheets: sheets,
	}

	err = t.Execute(builder, data)
	if err != nil {
		logger.Fatal(err)
	}

	t, err = tpl.Parse(TemplateSummary)
	if err != nil {
		logger.Fatal(err)
	}
	//输出总表
	if Config.Summary != "" {
		data2 := &struct {
			Name   string
			Sheets []*Sheet
		}{
			Name:   Config.Summary,
			Sheets: sheets,
		}
		err = t.Execute(builder, data2)
		if err != nil {
			logger.Fatal(err)
		}
	}
	return
}
