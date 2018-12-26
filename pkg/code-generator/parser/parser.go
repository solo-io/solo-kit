package parser

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/iancoleman/strcase"
	"github.com/pseudomuto/protokit"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

func ProcessDescriptors(projectConfig model.ProjectConfig, descriptors []*descriptor.FileDescriptorProto) (*model.Project, error) {
	req := &plugin_go.CodeGeneratorRequest{}
	for _, file := range descriptors {
		var added bool
		for _, addedFile := range req.GetFileToGenerate() {
			if addedFile == file.GetName() {
				added = true
			}
		}
		if added {
			continue
		}
		req.FileToGenerate = append(req.FileToGenerate, file.GetName())
		req.ProtoFile = append(req.ProtoFile, file)
	}
	return ParseRequest(projectConfig, req)
}

func ParseRequest(projectConfig model.ProjectConfig, req *plugin_go.CodeGeneratorRequest) (*model.Project, error) {
	log.Printf("project config: %v", projectConfig)

	descriptors := protokit.ParseCodeGenRequest(req)
	var messages []ProtoMessageWrapper
	for _, file := range descriptors {
		if file.Options == nil || file.Options.GoPackage == nil {
			log.Warnf("skipppig file %v must provide proto option go_package", file.GetName())
			continue
		}
		for _, msg := range file.GetMessages() {
			messages = append(messages, ProtoMessageWrapper{
				Message:   msg,
				GoPackage: *file.Options.GoPackage,
			})
		}
	}

	var services []*protokit.ServiceDescriptor
	for _, file := range descriptors {
		services = append(services, file.GetServices()...)
	}

	project := &model.Project{
		ProjectConfig: projectConfig,
		GroupName:     projectConfig.Name,
		Request:       req,
	}
	resources, resourceGroups, err := getResources(project, messages)
	if err != nil {
		return nil, err
	}

	xdsResources, err := getXdsResources(project, messages, services)
	if err != nil {
		return nil, err
	}

	project.Resources = resources
	project.ResourceGroups = resourceGroups
	project.XDSResources = xdsResources

	return project, nil
}

func loadProjectConfig(path string) (model.ProjectConfig, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return model.ProjectConfig{}, err
	}
	var pc model.ProjectConfig
	err = json.Unmarshal(b, &pc)
	return pc, err
}

func goName(n string) string {
	return strcase.ToCamel(strings.Split(n, ".")[0])
}

func collectFields(msg *protokit.Descriptor) []*model.Field {
	var fields []*model.Field
	for _, f := range msg.GetField() {
		fields = append(fields, &model.Field{
			Name:     f.GetName(),
			TypeName: f.GetTypeName(),
			IsOneof:  f.OneofIndex != nil,
			Original: f,
		})
	}
	log.Printf("%v", fields)
	return fields
}

func hasField(msg *protokit.Descriptor, fieldName, fieldType string) bool {
	for _, field := range msg.Fields {
		if field.GetName() == fieldName && field.GetTypeName() == fieldType {
			return true
		}
	}
	return false
}

func hasPrimitiveField(msg *protokit.Descriptor, fieldName string, fieldType descriptor.FieldDescriptorProto_Type) bool {
	for _, field := range msg.Fields {
		if field.GetName() == fieldName && field.GetType() == fieldType {
			return true
		}
	}
	return false
}

func getCommentValue(comments []string, key string) (string, bool) {
	for _, c := range comments {
		if strings.HasPrefix(c, key) {
			return strings.TrimPrefix(c, key), true
		}
	}
	return "", false
}
