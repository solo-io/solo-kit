package jsonschema

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/alecthomas/jsonschema"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	plugin "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/spf13/afero"
	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

var (
	logger = contextutils.LoggerFrom(context.TODO())

	allowNullValues              bool = false
	disallowAdditionalProperties bool = false
	disallowBigIntsAsStrings     bool = false
	debugLogging                 bool = false
	nestedMessagesAsReferences   bool = true

	PackageNotFoundError = func(pkg string) error {
		return errors.Errorf("no such package found: %s", pkg)
	}

	MessageNotFoundError = func(msg string) error {
		return errors.Errorf("no such message type named %s", msg)
	}

	defaultKubeType = v1beta1.JSONSchemaProps{
		Definitions: v1beta1.JSONSchemaDefinitions{
			"Resource": {
				Required: []string{"metadata", "status", "apiVersion", "kind", "spec"},
				Type:     "object",
				Properties: map[string]v1beta1.JSONSchemaProps{
					// "metadata": {
					// 	Type: "object",
					// 	Ref: proto.String(schemaRefName("Metadata")),
					// },
					// "status": {
					// 	Type: "object",
					// 	Ref: proto.String(schemaRefName("Status")),
					// },
					"apiVersion": {
						Type: "string",
					},
					"kind": {
						Type: "string",
					},
					"spec": {
						Type: "object",
					},
				},
			},
		},
	}
)

type ResourceStackItem struct {
	desc *descriptor.DescriptorProto
}

type ResourceStack interface {
	Push(item *ResourceStackItem)
	Pop() *ResourceStackItem
	Contains(descriptorProto *descriptor.DescriptorProto) *ResourceStackItem
}

type resourceStack struct {
	stack []*ResourceStackItem
	sync.Mutex
}

func (r *resourceStack) Contains(descriptorProto *descriptor.DescriptorProto) *ResourceStackItem {
	for _, v := range r.stack {
		if v.desc.GetName() == descriptorProto.GetName() {
			return v
		}
	}
	return nil
}

func NewResourceStack() *resourceStack {
	return &resourceStack{
		stack: make([]*ResourceStackItem, 0, 8),
	}
}

func (r *resourceStack) Push(item *ResourceStackItem) {
	r.Lock()
	defer r.Unlock()
	r.stack = append(r.stack, item)
}

func (r *resourceStack) Pop() *ResourceStackItem {
	r.Lock()
	defer r.Unlock()
	var val *ResourceStackItem
	if len(r.stack) > 0 {
		val = r.stack[len(r.stack)-1]
		r.stack = r.stack[0 : len(r.stack)-1]
	}
	return val
}

// ProtoPackage describes a package of Protobuf, which is an container of message types.
type ProtoPackage struct {
	name     string
	parent   *ProtoPackage
	children map[string]*ProtoPackage
	types    map[string]*descriptor.DescriptorProto
}

func NewProtoPackage() *ProtoPackage {
	return &ProtoPackage{
		name:     "",
		parent:   nil,
		children: make(map[string]*ProtoPackage),
		types:    make(map[string]*descriptor.DescriptorProto),
	}
}

func (pkg *ProtoPackage) reset() {
	pkg.children = make(map[string]*ProtoPackage)
	pkg.types = make(map[string]*descriptor.DescriptorProto)
}

func (pkg *ProtoPackage) registerType(pkgName *string, msg *descriptor.DescriptorProto) {
	if pkgName != nil {
		for _, node := range strings.Split(*pkgName, ".") {
			child, ok := pkg.children[node]
			if !ok {
				child = &ProtoPackage{
					name:     pkg.name + "." + node,
					parent:   pkg,
					children: make(map[string]*ProtoPackage),
					types:    make(map[string]*descriptor.DescriptorProto),
				}
				if pkg.name == "" {
					child.name = node
				}
				pkg.children[node] = child
			}
			pkg = child
		}
	}
	pkg.types[msg.GetName()] = msg
}

func (pkg *ProtoPackage) lookupType(name string) (*descriptor.DescriptorProto, bool) {
	if strings.HasPrefix(name, ".") {
		return pkg.relativelyLookupType(name[1:])
	}

	for ; pkg != nil; pkg = pkg.parent {
		if desc, ok := pkg.relativelyLookupType(name); ok {
			return desc, ok
		}
	}
	return nil, false
}

func relativelyLookupNestedType(desc *descriptor.DescriptorProto, name string) (*descriptor.DescriptorProto, bool) {
	components := strings.Split(name, ".")
componentLoop:
	for _, component := range components {
		for _, nested := range desc.GetNestedType() {
			if nested.GetName() == component {
				desc = nested
				continue componentLoop
			}
		}
		logger.Infof("no such nested message %s in %s", component, desc.GetName())
		return nil, false
	}
	return desc, true
}

func (pkg *ProtoPackage) relativelyLookupType(name string) (*descriptor.DescriptorProto, bool) {
	components := strings.SplitN(name, ".", 2)
	switch len(components) {
	case 0:
		logger.Debugf("empty message name")
		return nil, false
	case 1:
		found, ok := pkg.types[components[0]]
		return found, ok
	case 2:
		logger.Debugf("looking for %s in %s at %s (%v)", components[1], components[0], pkg.name, pkg)
		if child, ok := pkg.children[components[0]]; ok {
			found, ok := child.relativelyLookupType(components[1])
			return found, ok
		}
		if msg, ok := pkg.types[components[0]]; ok {
			found, ok := relativelyLookupNestedType(msg, components[1])
			return found, ok
		}
		logger.Infof("no such package nor message %s in %s", components[0], pkg.name)
		return nil, false
	default:
		logger.Fatalf("not reached")
		return nil, false
	}
}

func (pkg *ProtoPackage) relativelyLookupPackage(name string) (*ProtoPackage, bool) {
	components := strings.Split(name, ".")
	for _, c := range components {
		var ok bool
		pkg, ok = pkg.children[c]
		if !ok {
			return nil, false
		}
	}
	return pkg, true
}

// Convert a proto "field" (essentially a type-switch with some recursion):
func (g *generator) convertField(curPkg *ProtoPackage, desc *descriptor.FieldDescriptorProto, msg *descriptor.DescriptorProto) (*jsonschema.Type, error) {

	// Prepare a new jsonschema.Type for our eventual return value:
	jsonSchemaType := &jsonschema.Type{
		Properties: make(map[string]*jsonschema.Type),
	}

	if desc.TypeName != nil && desc.GetType() != descriptor.FieldDescriptorProto_TYPE_ENUM {
		_, ok := g.protoPackage.lookupType(desc.GetTypeName())
		if !ok {
			return nil, errors.Errorf("could not find proper type name")
		}
		pkgName := strings.TrimPrefix(desc.GetTypeName(), ".")
		jsonSchemaType.Title = pkgName
		jsonSchemaType.Ref = fmt.Sprintf("#/definitions/%s", pkgName)
	}

	// Switch the types, and pick a JSONSchema equivalent:
	switch desc.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_NUMBER},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_NUMBER
		}

	case descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SINT32:
		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_INTEGER},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
		}

	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_SINT64:
		jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_INTEGER})
		if !disallowBigIntsAsStrings {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})
		}
		if allowNullValues {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_NULL})
		}

	case descriptor.FieldDescriptorProto_TYPE_STRING,
		descriptor.FieldDescriptorProto_TYPE_BYTES:
		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_STRING},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
		}

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})
		jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_INTEGER})
		if allowNullValues {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_NULL})
		}

		// Go through all the enums we have, see if we can match any to this field by name:
		for _, enumDescriptor := range msg.GetEnumType() {

			// Each one has several values:
			for _, enumValue := range enumDescriptor.Value {

				// Figure out the entire name of this field:
				fullFieldName := fmt.Sprintf(".%v.%v", *msg.Name, *enumDescriptor.Name)

				// If we find ENUM values for this field then put them into the JSONSchema list of allowed ENUM values:
				if strings.HasSuffix(desc.GetTypeName(), fullFieldName) {
					jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Name)
					jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Number)
				}
			}
		}

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_BOOLEAN},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_BOOLEAN
		}

	case descriptor.FieldDescriptorProto_TYPE_GROUP,
		descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
		if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL {
			jsonSchemaType.AdditionalProperties = []byte("true")
		}
		if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
			jsonSchemaType.AdditionalProperties = []byte("false")
		}

	default:
		return nil, fmt.Errorf("unrecognized field type: %s", desc.GetType().String())
	}

	// Recurse array of primitive types:
	if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED && jsonSchemaType.Type != gojsonschema.TYPE_OBJECT {
		jsonSchemaType.Items = &jsonschema.Type{}

		if len(jsonSchemaType.Enum) > 0 {
			jsonSchemaType.Items.Enum = jsonSchemaType.Enum
			jsonSchemaType.Enum = nil
			jsonSchemaType.Items.OneOf = nil
		} else {
			jsonSchemaType.Items.Type = jsonSchemaType.Type
			jsonSchemaType.Items.OneOf = jsonSchemaType.OneOf
		}

		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_ARRAY},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_ARRAY
			jsonSchemaType.OneOf = []*jsonschema.Type{}
		}

		return jsonSchemaType, nil
	}

	// Recurse nested objects / arrays of objects (if necessary):
	if jsonSchemaType.Type == gojsonschema.TYPE_OBJECT {

		recordType, ok := g.protoPackage.lookupType(desc.GetTypeName())
		if !ok {
			return nil, MessageNotFoundError(desc.GetTypeName())
		}

		var recursedJSONSchemaType *jsonschema.Type
		var err error
		if nestedMessagesAsReferences {
			recursedJSONSchemaType = convertMessageTypeReference(desc.GetTypeName(), recordType)
		} else {
			recursedJSONSchemaType, err = g.convertMessageType(curPkg, recordType)
		}
		if err != nil {
			return nil, err
		}

		// The result is stored differently for arrays of objects (they become "items"):
		if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			jsonSchemaType.Items = recursedJSONSchemaType
			jsonSchemaType.Type = gojsonschema.TYPE_ARRAY
		} else {
			// Nested objects are more straight-forward:
			jsonSchemaType.Properties = recursedJSONSchemaType.Properties
		}

		// Optionally allow NULL values:
		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: jsonSchemaType.Type},
			}
			jsonSchemaType.Type = ""
		}
	}

	return jsonSchemaType, nil
}

func schemaRefName(packageName, name string) string {
	return fmt.Sprintf("#/definitions/%s", schemaTitleName(packageName, name))
}

func schemaTitleName(packageName, name string) string {
	if packageName == "" {
		return name
	}
	return fmt.Sprintf("%s.%s", packageName, name)
}

func convertMessageTypeReference(pkgName string, msg *descriptor.DescriptorProto) *jsonschema.Type {
	pkgName = strings.TrimPrefix(pkgName, ".")
	// Prepare a new jsonschema:
	jsonSchemaType := &jsonschema.Type{
		Properties: make(map[string]*jsonschema.Type),
		Version:    jsonschema.Version,
		Ref:        fmt.Sprintf("#/definitions/%s", pkgName),
		Title:      pkgName,
	}
	return jsonSchemaType
}

// Converts a proto "MESSAGE" into a JSON-Schema:
func (g *generator) convertMessageType(curPkg *ProtoPackage, msg *descriptor.DescriptorProto) (*jsonschema.Type, error) {

	// Prepare a new jsonschema:
	jsonSchemaType := &jsonschema.Type{
		Properties: make(map[string]*jsonschema.Type),
		Version:    jsonschema.Version,
		Ref:        schemaRefName(curPkg.name, msg.GetName()),
		Title:      schemaTitleName(curPkg.name, msg.GetName()),
	}

	// Optionally allow NULL values:
	if allowNullValues {
		jsonSchemaType.OneOf = []*jsonschema.Type{
			{Type: gojsonschema.TYPE_NULL},
			{Type: gojsonschema.TYPE_OBJECT},
		}
	} else {
		jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
	}

	// disallowAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
	if disallowAdditionalProperties {
		jsonSchemaType.AdditionalProperties = []byte("false")
	} else {
		jsonSchemaType.AdditionalProperties = []byte("true")
	}

	logger.Debugf("Converting message: %s", proto.MarshalTextString(msg))
	for _, fieldDesc := range msg.GetField() {
		recursedJSONSchemaType, err := g.convertField(curPkg, fieldDesc, msg)
		if err != nil {
			logger.Errorf("Failed to convert field %s in %s: %v", fieldDesc.GetName(), msg.GetName(), err)
			return jsonSchemaType, err
		}
		jsonSchemaType.Properties[fieldDesc.GetName()] = recursedJSONSchemaType
	}
	return jsonSchemaType, nil
}

// Converts a proto "ENUM" into a JSON-Schema:
func convertEnumType(enum *descriptor.EnumDescriptorProto) (jsonschema.Type, error) {

	// Prepare a new jsonschema.Type for our eventual return value:
	jsonSchemaType := jsonschema.Type{
		Version: jsonschema.Version,
	}

	// Allow both strings and integers:
	jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: "string"})
	jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: "integer"})

	// Add the allowed values:
	for _, enumValue := range enum.Value {
		jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Name)
		jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Number)
	}

	return jsonSchemaType, nil
}

func (g *generator) convertFile(file *descriptor.FileDescriptorProto) ([]*jsonschema.Type, error) {
	// Input filename:
	protoFileName := path.Base(file.GetName())

	// Prepare a list of responses:
	response := make([]*jsonschema.Type, 0, len(file.GetMessageType()))

	// Warn about multiple messages / enums in files:
	if len(file.GetMessageType()) > 1 {
		logger.Warnf("protoc-gen-jsonschema will create multiple MESSAGE schemas (%d) from one proto file (%v)", len(file.GetMessageType()), protoFileName)
	}
	if len(file.GetEnumType()) > 1 {
		logger.Warnf("protoc-gen-jsonschema will create multiple ENUM schemas (%d) from one proto file (%v)", len(file.GetEnumType()), protoFileName)
	}

	// Generate standalone ENUMs:
	if len(file.GetMessageType()) == 0 {
		for _, enum := range file.GetEnumType() {
			jsonSchemaFileName := fmt.Sprintf("%s.jsonschema", enum.GetName())
			logger.Infof("Generating JSON-schema for stand-alone ENUM (%v) in file [%v] => %v", enum.GetName(), protoFileName, jsonSchemaFileName)
			enumJsonSchema, err := convertEnumType(enum)
			if err != nil {
				logger.Errorf("Failed to convert %s: %v", protoFileName, err)
				return nil, err
			} else {
				response = append(response, &enumJsonSchema)
			}
		}
	} else {
		// Otherwise process MESSAGES (packages):
		pkg, ok := g.protoPackage.relativelyLookupPackage(file.GetPackage())
		if !ok {
			return nil, fmt.Errorf("no such package found: %s", file.GetPackage())
		}
		for _, msg := range file.GetMessageType() {
			jsonSchemaFileName := fmt.Sprintf("%s.jsonschema", msg.GetName())
			logger.Infof("Generating JSON-schema for MESSAGE (%v) in file [%v] => %v", msg.GetName(), protoFileName, jsonSchemaFileName)
			messageJSONSchema, err := g.convertMessageType(pkg, msg)
			if err != nil {
				logger.Errorf("Failed to convert %s: %v", protoFileName, err)
				return nil, err
			} else {
				// Marshal the JSON-Schema into JSON:
				response = append(response, messageJSONSchema)
			}
		}
	}

	return response, nil
}

type Generator interface {
	Convert(resource *model.Resource) (*jsonschema.Type, error)
	KubeConvert(resource *model.Resource) (*v1beta1.JSONSchemaProps, error)
}

type generator struct {
	// Internal objects used to construct schema types, and build kube schemas
	protoPackage   *ProtoPackage
	fs             afero.Fs
	generatedTypes map[string]*jsonschema.Type
}

func (g *generator) KubeConvert(resource *model.Resource) (*v1beta1.JSONSchemaProps, error) {
	panic("implement me")
}

func (g *generator) Convert(resource *model.Resource) (*jsonschema.Type, error) {
	preGenType, ok := g.generatedTypes[schemaTitleName(resource.ProtoPackage, resource.Name)]
	if !ok {
		return nil, errors.Errorf("could not find ref to previously existing type")
	}
	allTypesToAdd, err := g.buildKubeSpec(preGenType)
	if err != nil {
		return nil, err
	}

	definitions := make(jsonschema.Definitions, len(allTypesToAdd))
	for _, v := range allTypesToAdd {
		preGenType, ok := g.generatedTypes[v.Title]
		if !ok {
			return nil, errors.Errorf("could not find ref to previously existing type")
		}
		definitions[v.Title] = preGenType
	}
	definitions[preGenType.Title] = preGenType
	result := &jsonschema.Type{
		Version:     jsonschema.Version,
		Ref:         preGenType.Ref,
		Definitions: definitions,
	}
	return result, nil
}

func (g *generator) Marshal(schemaType *jsonschema.Type) ([]byte, error) {
	jsonSchemaJSON, err := json.MarshalIndent(schemaType, "", "    ")
	if err != nil {
		return nil, err
	}
	return jsonSchemaJSON, nil
}

func (g *generator) buildKubeSpec(recursedType *jsonschema.Type) ([]*jsonschema.Type, error) {
	var result []*jsonschema.Type
	pregenTypes := make([]*jsonschema.Type, 0, len(recursedType.Properties))
	for _, val := range recursedType.Properties {
		preGenType, ok := g.generatedTypes[val.Title]
		if !ok {
			continue
		}
		pregenTypes = append(pregenTypes, preGenType)
	}
	if len(pregenTypes) == 0 {
		result = append(result, recursedType)
	}
	for _, preGenType := range pregenTypes {
		recursiveResult, err := g.buildKubeSpec(preGenType)
		if err != nil {
			return nil, err
		}

		result = append(result, recursiveResult...)

	}
	return result, nil
}

func NewGenerator(req *plugin.CodeGeneratorRequest) (*generator, error) {

	g := &generator{protoPackage: NewProtoPackage(), fs: afero.NewOsFs(), generatedTypes: make(map[string]*jsonschema.Type)}

	for _, file := range req.GetProtoFile() {
		for _, msg := range file.GetMessageType() {
			logger.Debugf("Loading a message type %s from package %s", msg.GetName(), file.GetPackage())
			g.protoPackage.registerType(file.Package, msg)
		}
	}

	for _, file := range req.GetProtoFile() {
		logger.Debugf("Converting file (%v)", file.GetName())
		types, err := g.convertFile(file)
		if err != nil {
			return nil, err
		}
		for _, v := range types {
			g.generatedTypes[v.Title] = v
		}
	}
	return g, nil
}
