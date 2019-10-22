package parser

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/iancoleman/strcase"
	"github.com/ilackarms/protokit"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

type ProjectMap map[*model.ProjectConfig]*model.Project

func ProcessDescriptorsFromConfigs(projectConfigs []*model.ProjectConfig, protoDescriptors []*descriptor.FileDescriptorProto) (ProjectMap, error) {
	projectMap := make(ProjectMap)
	for _, projectConfig := range projectConfigs {
		project, err := ProcessDescriptorsFromConfig(projectConfig, projectConfigs, protoDescriptors)
		if err != nil {
			return nil, err
		}
		projectMap[projectConfig] = project
	}
	return projectMap, nil
}

// Build a 'Project' object that contains a resource for each message that:
// - is contained in the FileDescriptor and
// - is a solo kit resource (i.e. it has a field named 'metadata')
func ProcessDescriptorsFromConfig(projectConfig *model.ProjectConfig, allProjectConfigs []*model.ProjectConfig, descriptors []*descriptor.FileDescriptorProto) (*model.Project, error) {
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
	return parseRequest(projectConfig, allProjectConfigs, req)
}

func parseRequest(projectConfig *model.ProjectConfig, allProjectConfigs []*model.ProjectConfig, req *plugin_go.CodeGeneratorRequest) (*model.Project, error) {
	log.Printf("project config: %v", projectConfig)

	descriptors := protokit.ParseCodeGenRequest(req)
	var messages []ProtoMessageWrapper
	for _, file := range descriptors {
		if file.Options == nil || file.Options.GoPackage == nil {
			log.Warnf("skipping file %v must provide proto option go_package", file.GetName())
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
		ProjectConfig: *projectConfig,
		ProtoPackage:  projectConfig.Name,
		Request:       req,
		Descriptors:   descriptors,
	}
	resources, resourceGroups, err := getResources(project, allProjectConfigs, messages)
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

func goName(n string) string {
	return strcase.ToCamel(strings.Split(n, ".")[0])
}

func collectFields(msg *protokit.Descriptor) []*model.Field {
	var fields []*model.Field
	for _, f := range msg.GetField() {
		skipHashing := proto.GetBoolExtension(f.Options, core.E_SkipHashing, false)
		fields = append(fields, &model.Field{
			Name:        f.GetName(),
			TypeName:    f.GetTypeName(),
			IsOneof:     f.OneofIndex != nil,
			SkipHashing: skipHashing,
			Original:    f,
		})
	}
	log.Printf("%v", fields)
	return fields
}

func collectOneofs(msg *protokit.Descriptor) []*model.Oneof {
	var oneofs []*model.Oneof
	for _, f := range msg.GetOneofDecl() {
		oneofs = append(oneofs, &model.Oneof{
			Name: f.GetName(),
		})
	}
	return oneofs
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

func FilterDuplicateDescriptors(descriptors []*model.DescriptorWithPath) []*model.DescriptorWithPath {
	var uniqueDescriptors []*model.DescriptorWithPath
	for _, desc := range descriptors {
		unique, matchingDesc := isUnique(desc, uniqueDescriptors)
		// if this proto file first came in as an import, but later as a solo-kit project proto,
		// ensure the original proto gets updated with the correct proto file path
		if !unique && matchingDesc.ProtoFilePath == "" {
			matchingDesc.ProtoFilePath = desc.ProtoFilePath
		}
		if unique {
			uniqueDescriptors = append(uniqueDescriptors, desc)
		}
	}
	return uniqueDescriptors
}

// If it finds a matching proto, also returns the matching proto's file descriptor
func isUnique(desc *model.DescriptorWithPath, descriptors []*model.DescriptorWithPath) (bool, *model.DescriptorWithPath) {
	for _, existing := range descriptors {
		existingCopy := proto.Clone(existing.FileDescriptorProto).(*descriptor.FileDescriptorProto)
		existingCopy.Name = desc.Name
		if proto.Equal(existingCopy, desc.FileDescriptorProto) {
			return false, existing
		}
	}
	return true, nil
}
