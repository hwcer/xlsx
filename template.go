package xlsx

import (
	"fmt"
	"github.com/hwcer/logger"
	"strings"
	"text/template"
)

var tpl *template.Template

// TemplateTitle 包信息模版
const TemplateTitle = `syntax = "proto3";
option go_package = "./;<% .Package %>";
`

// TemplateDummy 全局子对象模版
const TemplateDummy = `message <%.Name%>{ <%range .Fields%>
	<%.Type%> <%.Name%> = <%.ProtoIndex%>;<%end%>
}
`

// TemplateMessage 基本表结构
const TemplateMessage = `
<%- range .Sheets%>
message <%.ProtoName%>{
	<%- range .Fields %>
	<%ProtoRequire .%> <%.Name%> = <%.ProtoIndex%>; //<% .ProtoDesc%><%end%>
}
<%- if IsArray .SheetType %>
message <%.ProtoName%>Array{
	repeated <%.ProtoName%> Coll = 1;
}
<%- end%>
<%- end%>
`

// TemplateSummary 总表模版
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
		"IsArray":      TemplateIsArray,
		"SummaryType":  TemplateSummaryType,
		"ProtoRequire": TemplateProtoRequire,
	})
	tpl.Delims("<%", "%>")
}

func TemplateIsArray(t SheetType) bool {
	return t == SheetTypeArr
}

func TemplateProtoRequire(field *Field) string {
	if handle, ok := protoRequireHandles[field.ProtoRequire]; ok {
		return handle.Require(field)
	} else {
		return field.ProtoType
	}
	//switch t {
	//case FieldTypeArray, FieldTypeArrObj:
	//	return "repeated "
	//default:
	//	return ""
	//}
}

func TemplateSummaryType(sheet *Sheet) string {
	switch sheet.SheetType {
	case SheetTypeObj:
		return sheet.ProtoName
<<<<<<< HEAD
	case TableTypeArr:
		return fmt.Sprintf("%vArray", sheet.ProtoName)
=======
	case SheetTypeArr:
		t = fmt.Sprintf("%v%vArray", sheet.ProtoName, Config.Suffix)
>>>>>>> cabfa43f3ff1057a9154cc80e61d02d81319fa71
	default:
		primary := sheet.Fields[0]
		return fmt.Sprintf("map<%v,%v>", primary.ProtoType, sheet.ProtoName)
	}
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
