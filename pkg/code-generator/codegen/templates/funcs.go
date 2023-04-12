package templates

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/utils"

	htmltemplate "html/template"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/iancoleman/strcase"
	gendoc "github.com/pseudomuto/protoc-gen-doc"
	"github.com/pseudomuto/protokit"
	"github.com/solo-io/go-utils/stringutils"
)

var primitiveTypes = map[descriptor.FieldDescriptorProto_Type]string{
	descriptor.FieldDescriptorProto_TYPE_FLOAT:  "float",
	descriptor.FieldDescriptorProto_TYPE_DOUBLE: "float",
	descriptor.FieldDescriptorProto_TYPE_STRING: "string",
	descriptor.FieldDescriptorProto_TYPE_BOOL:   "bool",
	descriptor.FieldDescriptorProto_TYPE_UINT32: "int",
	descriptor.FieldDescriptorProto_TYPE_UINT64: "int",
	descriptor.FieldDescriptorProto_TYPE_INT32:  "int",
	descriptor.FieldDescriptorProto_TYPE_INT64:  "int",
	descriptor.FieldDescriptorProto_TYPE_ENUM:   "***TODO ENUMS***!",
	descriptor.FieldDescriptorProto_TYPE_BYTES:  "***TODO BYTES***!",
}

var magicCommentRegex = regexp.MustCompile("@solo-kit:.*")

var Funcs = template.FuncMap{
	"resourceBelongsToProject": utils.IsProjectResource,
	"join":                     strings.Join,
	"lowercase":                strings.ToLower,
	"lower_camel":              strcase.ToLowerCamel,
	"upper_camel":              strcase.ToCamel,
	"snake":                    strcase.ToSnake,
	"p":                        gendoc.PFilter,
	"para":                     gendoc.ParaFilter,
	"nobr":                     gendoc.NoBrFilter,
	"fieldType":                fieldType,
	"yamlType":                 yamlType,
	"noescape":                 noEscape,
	"printfptr":                printPointer,
	"remove_magic_comments": func(in string) string {
		lines := strings.Split(in, "\n")
		var linesWithoutMagicComments []string
		for _, line := range lines {
			if magicCommentRegex.MatchString(line) {
				continue
			}
			linesWithoutMagicComments = append(linesWithoutMagicComments, line)
		}
		return strings.Join(linesWithoutMagicComments, "\n")
	},
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
	"unique": func(vals []string) []string {
		result := make([]string, 0, len(vals))
		for _, v := range vals {
			if !stringutils.ContainsString(v, result) {
				result = append(result, v)
			}
		}
		return result
	},
	"backtick": func() string {
		return "`"
	},
	"inc": func(a int) int {
		return a + 1
	},
	"ge": func(a, b int) bool {
		return a >= b
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
			fallthrough
		case "NamespacedStatuses":
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

func fieldType(field *protokit.FieldDescriptor) string {
	fieldType := func() string {
		switch field.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			return field.GetTypeName()
		}
		if typeName, ok := primitiveTypes[field.GetType()]; ok {
			return typeName
		}
		return "UNSUPPORTED: " + field.GetType().String() + ": " + field.GetName()
	}()
	if field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		fieldType = "[" + fieldType + "]"
	}
	return fieldType
}
