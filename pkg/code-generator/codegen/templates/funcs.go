package templates

import (
	"fmt"
	"github.com/pseudomuto/protokit"
	htmltemplate "html/template"
	"strings"
	"text/template"

	"github.com/ilackarms/protoc-gen-doc"

	"github.com/iancoleman/strcase"
)

var Funcs = template.FuncMap{
	"join":        strings.Join,
	"lowercase":   strings.ToLower,
	"lower_camel": strcase.ToLowerCamel,
	"upper_camel": strcase.ToCamel,
	"snake":       strcase.ToSnake,
	"p":           gendoc.PFilter,
	"para":        gendoc.ParaFilter,
	"nobr":        gendoc.NoBrFilter,
	"fieldType": func(field *protokit.FieldDescriptor) string {

	},
	"yamlType":    yamlType,
	"noescape":    noEscape,
	"linkForType": linkForType,
	"printfptr":   printPointer,
	"new_str_slice": func() *[]string {
		var v []string
		return &v
	},
	"append_str_slice": func(to *[]string, str string) *[]string {
		*to = append(*to, str)
		return to
	},
	"join_str_slice": func(slc *[]string, sep string) string {
		return strings.Join(*slc, sep)
	},
	"new_bool": func() *bool {
		var v bool
		return &v
	},
	"set_bool": func(v *bool, val bool) *bool {
		*v = val
		return v
	},
}

func printPointer(format string, p *string) string {
	return fmt.Sprintf(format, *p)
}

func yamlType(longType, label string) string {
	yamlType := func() string {
		if strings.HasPrefix(longType, "map<") {
			return longType
		}
		switch longType {
		case "string":
			fallthrough
		case "uint32":
			fallthrough
		case "bool":
			fallthrough
		case "int32":
			return longType
		case "Status":
			return "(read only)"
		}
		return "{" + longType + "}"
	}()
	if label == "repeated" {
		yamlType = "[" + yamlType + "]"
	}
	return yamlType
}

func noEscape(s string) htmltemplate.HTML {
	return htmltemplate.HTML(s)
}

func linkForType(longType, fullType string) string {
	if !isObjectType(longType) {
		return longType //no linking for primitives
	}
	var link string
	switch {
	case longType == "google.protobuf.Duration":
		link = "https://developers.google.com/protocol-buffers/docs/reference/csharp/class/google/protobuf/well-known-types/duration"
	case longType == "google.protobuf.Struct":
		link = "https://developers.google.com/protocol-buffers/docs/reference/csharp/class/google/protobuf/well-known-types/struct"
	default:
		link = strcase.ToSnake(fullType) + ".sk.md#" + fullType
	}
	return "[" + longType + "](" + link + ")"
}

func isObjectType(longType string) bool {
	if strings.HasPrefix(longType, "map<") {
		return false
	}
	switch longType {
	case "string":
		fallthrough
	case "uint32":
		fallthrough
	case "bool":
		fallthrough
	case "int32":
		return false
	}
	return true
}
