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
`

const TemplateDummy = `message <%.Name%>{ <%range .Fields%>
	<%DummyRequire .%> <%.Name%> = <%.ProtoIndex%>;<%end%>
}
`

const TemplateMessage = `
<%- range .Sheets%>
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
	<%SummaryType .%> <%.ProtoName%>=<%.ProtoIndex%>;
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
//	return t == TableTypeArr
//}

func TemplateProtoRequire(field *Field) string {
	handle := Require(field.ProtoType)
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

func TemplateSummaryType(sheet *Sheet) string {
	primary := sheet.Fields[0]
	//var t string
	switch sheet.SheetType {
	case TableTypeObj:
		return sheet.ProtoName
	//case TableTypeArr:
	//	t = fmt.Sprintf("%v%vArray", sheet.ProtoName, Config.Suffix)
	default:
		//t = fmt.Sprintf("%v%v", sheet.ProtoName, Config.Suffix)
		return fmt.Sprintf("map<%v,%v>", primary.ProtoType, sheet.ProtoName)
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

	return
}
