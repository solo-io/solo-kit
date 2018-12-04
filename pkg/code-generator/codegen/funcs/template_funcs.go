package funcs

import (
	"fmt"
	"github.com/solo-io/solo-kit/pkg/errors"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"

	htmltemplate "html/template"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/iancoleman/strcase"
	"github.com/ilackarms/protoc-gen-doc"
	"github.com/pseudomuto/protokit"
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

func TemplateFuncs(project *model.Project) template.FuncMap {
	return template.FuncMap{
		"join":        strings.Join,
		"lowercase":   strings.ToLower,
		"lower_camel": strcase.ToLowerCamel,
		"upper_camel": strcase.ToCamel,
		"snake":       strcase.ToSnake,
		"p":           gendoc.PFilter,
		"para":        gendoc.ParaFilter,
		"nobr":        gendoc.NoBrFilter,
		"fieldType":   fieldType(project),
		"yamlType":    yamlType,
		"noescape":    noEscape,
		"linkForType": linkForType(project),
		"printfptr":   printPointer,
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
	}
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

func fieldType(project *model.Project) func(field *protokit.FieldDescriptor) (string, error) {
	return func(field *protokit.FieldDescriptor) (string, error) {
		fieldTypeStr := func() string {
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
			fieldTypeStr = "[]" + strings.TrimPrefix(fieldTypeStr, ".")
		}
		if strings.HasSuffix(fieldTypeStr, "Entry") {
			msg, err := getMessageForField(project, field)
			if err != nil {
				return "", err
			}
			if len(msg.Field) != 2 {
				return "", errors.Errorf("message %v was Entry type, expected map", msg.GetName())
			}
			key, err := fieldType(project)(&protokit.FieldDescriptor{FieldDescriptorProto: msg.Field[0]})
			if err != nil {
				return "", err
			}
			val, err := fieldType(project)(&protokit.FieldDescriptor{FieldDescriptorProto: msg.Field[1]})
			if err != nil {
				return "", err
			}
			fieldTypeStr = "map<" + key + ", " + val + ">"
		}
		return fieldTypeStr, nil
	}
}

func linkForType(project *model.Project) func(field *protokit.FieldDescriptor) (string, error) {
	return func(field *protokit.FieldDescriptor) (string, error) {
		typeName, err := fieldType(project)(field)
		if err != nil {
			return "", err
		}
		if _, ok := primitiveTypes[field.GetType()]; ok {
			return typeName, nil
		}
		file, err := getFileForField(project, field)
		if err != nil {
			return "", err
		}
		msg, err := getMessageForField(project, field)
		if err != nil {
			return "", err
		}
		link := filepath.Base(strcase.ToSnake(file.GetName())) + ".sk.md#" + msg.GetName()
		if strings.Contains(typeName, "google.protobuf.Duration") {
			link = "https://developers.google.com/protocol-buffers/docs/reference/csharp/class/google/protobuf/well-known-types/duration"
		}
		if strings.Contains(typeName, "google.protobuf.Struct") {
			link = "https://developers.google.com/protocol-buffers/docs/reference/csharp/class/google/protobuf/well-known-types/struct"
		}
		if strings.Contains(typeName, "google.protobuf.StringValue") {
			link = "https://developers.google.com/protocol-buffers/docs/reference/csharp/class/google/protobuf/well-known-types/string-value"
		}
		linkText := "[" + typeName + "](" + link + ")"
		return linkText, nil
	}
}

func getFileForField(project *model.Project, field *protokit.FieldDescriptor) (*descriptor.FileDescriptorProto, error) {
	parts := strings.Split(strings.TrimPrefix(field.GetTypeName(), "."), ".")
	if strings.HasSuffix(parts[len(parts)-1], "Entry") {
		parts = parts[:len(parts)-1]
	}
	messageName := parts[len(parts)-1]
	packageName := strings.Join(parts[:len(parts)-1], ".")
	for _, protoFile := range project.Request.GetProtoFile() {
		if protoFile.GetPackage() == packageName {
			for _, msg := range protoFile.GetMessageType() {
				if messageName == msg.GetName() {
					return protoFile, nil
				}
			}
		}
	}
	return nil, errors.Errorf("message %v.%v not found", packageName, messageName)
}

func getMessageForField(project *model.Project, field *protokit.FieldDescriptor) (*descriptor.DescriptorProto, error) {
	parts := strings.Split(strings.TrimPrefix(field.GetTypeName(), "."), ".")
	messageName := parts[len(parts)-1]
	var nestedMessageParent string
	if strings.HasSuffix(parts[len(parts)-1], "Entry") {
		parts = parts[:len(parts)-1]
		nestedMessageParent = parts[len(parts)-1]
	}
	packageName := strings.Join(parts[:len(parts)-1], ".")
	for _, protoFile := range project.Request.ProtoFile {
		if protoFile.GetPackage() == packageName {
			for _, msg := range protoFile.GetMessageType() {
				if messageName == msg.GetName() {
					return msg, nil
				}
				if nestedMessageParent == msg.GetName() {
					for _, nestedMsg := range msg.GetNestedType() {
						if messageName == nestedMsg.GetName() {
							return nestedMsg, nil
						}
					}
				}
			}
		}
	}
	return nil, errors.Errorf("message %v.%v not found", packageName, messageName)
}

//func linkForMap(project *model.Project, field *protokit.FieldDescriptor) string {
//
//	parts := strings.Split(strings.TrimPrefix(field.GetTypeName(), "."), ".")
//	parts = parts[:len(parts)-1]
//	messageName := parts[len(parts)-1]
//	packageName := strings.Join(parts[:len(parts)-1], ".")
//	for _, protoFile := range project.Request.GetProtoFile() {
//		if protoFile.GetPackage() == packageName {
//			for _, msg := range protoFile.GetMessageType() {
//				if messageName == msg.GetName() {
//					return protoFile, nil
//				}
//			}
//		}
//	}
//	return nil, errors.Errorf("message %v.%v not found", packageName, messageName)
//
//	for _, protoFile := range project.Request.ProtoFile {
//		messages := protoFile.MessageType
//		// remove "entry" types, we are converting these back to map<string, string>
//		for _, message := range messages {
//			if strings.HasSuffix(message.Name, "Entry") {
//				if len(message.Fields) != 2 {
//					log.Fatalf("bad assumption: %#v is not a map entry, or doesn't have 2 fields", message)
//				}
//				mapEntriesToFix[message.Name] = mapEntry{key: message.Fields[0], value: message.Fields[1]}
//			} else {
//				protoFile.Messages = append(protoFile.Messages, message)
//			}
//		}
//	}
//	for _, protoFile := range protoDescriptor.Files {
//		for _, message := range protoFile.Messages {
//			for _, field := range message.Fields {
//				if entry, ok := mapEntriesToFix[field.Type]; ok {
//					field.Type = "map<" + entry.key.Type + "," + entry.value.Type + ">"
//					field.FullType = "map<" + entry.key.FullType + "," + entry.value.FullType + ">"
//					field.LongType = "map<" + entry.key.LongType + "," + entry.value.LongType + ">"
//					field.Label = ""
//					log.Printf("changed field %v", field.Name)
//				}
//			}
//		}
//	}
//}
