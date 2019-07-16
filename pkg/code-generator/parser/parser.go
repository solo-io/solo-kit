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

func ProcessDescriptors(versionConfig *model.VersionConfig, apiGroup *model.ApiGroup, descriptors []*descriptor.FileDescriptorProto) (*model.Version, error) {
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
	return parseRequest(versionConfig, apiGroup, req)
}

func parseRequest(versionConfig *model.VersionConfig, apiGroup *model.ApiGroup, req *plugin_go.CodeGeneratorRequest) (*model.Version, error) {
	log.Printf("version config: %v", versionConfig)

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

	version := &model.Version{
		VersionConfig: *versionConfig,
		ProtoPackage:  versionConfig.ApiGroup.Name,
		Request:       req,
	}
	resources, err := getResources(version, apiGroup, messages)
	if err != nil {
		return nil, err
	}

	xdsResources, err := getXdsResources(version, messages, services)
	if err != nil {
		return nil, err
	}

	version.Resources = resources
	version.XDSResources = xdsResources

	return version, nil
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
