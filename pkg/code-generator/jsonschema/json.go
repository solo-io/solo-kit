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
	"github.com/solo-io/solo-kit/pkg/code-generator/jsonschema/internal/prototree"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/spf13/afero"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/sync/errgroup"
)

var (
	logger = contextutils.LoggerFrom(context.TODO())

	PackageNotFoundError = func(pkg string) error {
		return errors.Errorf("no such package found: %s", pkg)
	}

	MessageNotFoundError = func(msg string) error {
		return errors.Errorf("no such message type named %s", msg)
	}

	UnrecognizedFieldTypeError = func(name string) error {
		return errors.Errorf("unrecognized field type: %s", name)
	}

	TypeNotFoundError = func(typeName string) error {
		return errors.Errorf("could not find proper type name")
	}

	NoPreviouslyRegisteredTypeError = func(typeName string) error {
		return errors.Errorf("could not find ref to previously existing type %s", typeName)
	}

	NotCustomResourceJsonError = func(title string) error {
		return errors.Errorf("%s is not a valid custom resource json schema")
	}

	defaultOptions = &Options{
		AllowNullValues:              false,
		DisallowAdditionalProperties: false,
		DisallowBigIntsAsStrings:     false,
	}
)

// Convert a proto "field" (essentially a type-switch with some recursion):
func (g *generator) convertField(curPkg *prototree.ProtoPackage, desc *descriptor.FieldDescriptorProto, msg *descriptor.DescriptorProto) (*jsonschema.Type, error) {

	// Prepare a new jsonschema.Type for our eventual return value:
	jsonSchemaType := &jsonschema.Type{
		Properties: make(map[string]*jsonschema.Type),
	}

	if desc.TypeName != nil && desc.GetType() != descriptor.FieldDescriptorProto_TYPE_ENUM {
		_, ok := g.protoPackage.LookupType(desc.GetTypeName())
		if !ok {
			return nil, TypeNotFoundError(desc.GetTypeName())
		}
		pkgName := strings.TrimPrefix(desc.GetTypeName(), ".")
		jsonSchemaType.Title = pkgName
		jsonSchemaType.Ref = fmt.Sprintf("#/definitions/%s", pkgName)
	}

	// Switch the types, and pick a JSONSchema equivalent:
	switch desc.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if g.opts.AllowNullValues {
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
		if g.opts.AllowNullValues {
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
		if !g.opts.DisallowBigIntsAsStrings {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})
		}
		if g.opts.AllowNullValues {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_NULL})
		}

	case descriptor.FieldDescriptorProto_TYPE_STRING,
		descriptor.FieldDescriptorProto_TYPE_BYTES:
		if g.opts.AllowNullValues {
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
		if g.opts.AllowNullValues {
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
		if g.opts.AllowNullValues {
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
		if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL ||
			(desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED && IsMap(msg)) {
			jsonSchemaType.AdditionalProperties = []byte("true")
		}
		if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
			jsonSchemaType.AdditionalProperties = []byte("false")
		}

	default:
		return nil, UnrecognizedFieldTypeError(desc.GetType().String())
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

		if g.opts.AllowNullValues {
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

	// If the field is an object that create a reference go that object
	if jsonSchemaType.Type == gojsonschema.TYPE_OBJECT {
		var err error
		jsonSchemaType, err = g.handleNestedObject(desc, jsonSchemaType)
		if err != nil {
			return nil, err
		}
	}

	return jsonSchemaType, nil
}

func (g *generator) handleNestedObject(desc *descriptor.FieldDescriptorProto, jsonSchemaType *jsonschema.Type) (*jsonschema.Type, error) {
	recordType, ok := g.protoPackage.LookupType(desc.GetTypeName())
	if !ok {
		return nil, MessageNotFoundError(desc.GetTypeName())
	}
	recursedJSONSchemaType := &jsonschema.Type{
		Version: jsonschema.Version,
	}

	if IsMap(recordType) {
		// TODO(EItanya) add validation for map secondary type
		// Currently this code sets additional to true rather than limiting it to the secondary type
		recursedJSONSchemaType.Title = desc.GetJsonName()
		jsonSchemaType.Properties = nil
		jsonSchemaType.Ref = ""
		jsonSchemaType.AdditionalProperties = []byte("true")
	} else {
		pkgName := strings.TrimPrefix(desc.GetTypeName(), ".")
		// recursedJSONSchemaType.Ref = fmt.Sprintf("#/definitions/%s", pkgName)
		recursedJSONSchemaType.Title = pkgName
		// The result is stored differently for arrays of objects (they become "items"):
		if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			jsonSchemaType.Items = recursedJSONSchemaType
			jsonSchemaType.Type = gojsonschema.TYPE_ARRAY
		}
		jsonSchemaType.Properties = recursedJSONSchemaType.Properties
	}

	// Optionally allow NULL values:
	if g.opts.AllowNullValues {
		jsonSchemaType.OneOf = []*jsonschema.Type{
			{Type: gojsonschema.TYPE_NULL},
			{Type: jsonSchemaType.Type},
		}
		jsonSchemaType.Type = ""
	}
	return jsonSchemaType, nil
}

func IsMap(msg *descriptor.DescriptorProto) bool {
	if msg.GetNestedType() != nil {
		return false
	}
	if len(msg.GetField()) != 2 {
		return false
	}
	key, value := false, false
	for _, field := range msg.GetField() {
		// Best guess that this is a map
		if field.GetName() == "key" && field.GetType() == descriptor.FieldDescriptorProto_TYPE_STRING {
			key = true
		}
		if field.GetName() == "value" {
			value = true
		}
	}
	return key && value
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

// Converts a proto "MESSAGE" into a JSON-Schema:
func (g *generator) convertMessageType(curPkg *prototree.ProtoPackage, msg *descriptor.DescriptorProto) (*jsonschema.Type, error) {

	// Prepare a new jsonschema:
	jsonSchemaType := &jsonschema.Type{
		Properties: make(map[string]*jsonschema.Type),
		Version:    jsonschema.Version,
		Title:      schemaTitleName(curPkg.Name(), msg.GetName()),
	}

	// Optionally allow NULL values:
	if g.opts.DisallowBigIntsAsStrings {
		jsonSchemaType.OneOf = []*jsonschema.Type{
			{Type: gojsonschema.TYPE_NULL},
			{Type: gojsonschema.TYPE_OBJECT},
		}
	} else {
		jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
	}

	// DisallowAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
	// if g.opts.DisallowAdditionalProperties {
	// 	jsonSchemaType.AdditionalProperties = []byte("false")
	// } else {
	// 	jsonSchemaType.AdditionalProperties = []byte("true")
	// }

	logger.Debugf("Converting message: %s", proto.MarshalTextString(msg))

	var oneOfTypes []*jsonschema.Type
	for idx := range msg.GetOneofDecl() {
		var oneOfFields []*jsonschema.Type
		for _, fieldDesc := range msg.GetField() {
			if fieldDesc.OneofIndex == nil || fieldDesc.GetOneofIndex() != int32(idx) {
				continue
			}
			childJsonType, err := g.convertField(curPkg, fieldDesc, msg)
			if err != nil {
				logger.Errorf("Failed to convert field %s in %s: %v", fieldDesc.GetName(), msg.GetName(), err)
				return jsonSchemaType, err
			}
			props := make(map[string]*jsonschema.Type, 1)
			props[fieldDesc.GetJsonName()] = childJsonType
			oneOfFields = append(oneOfFields, &jsonschema.Type{
				Properties: props,
			})
		}
		oneOfType := &jsonschema.Type{
			Type:  gojsonschema.TYPE_OBJECT,
			OneOf: oneOfFields,
		}
		oneOfTypes = append(oneOfTypes, oneOfType)
	}

	if len(oneOfTypes) > 1 {
		anyOf := &jsonschema.Type{
			Version: jsonschema.Version,
			AnyOf:   oneOfTypes,
		}
		byt, err := Marshal(anyOf)
		if err != nil {
			return nil, err
		}
		jsonSchemaType.AdditionalProperties = byt
	} else if len(oneOfTypes) == 1 {
		byt, err := Marshal(oneOfTypes[0])
		if err != nil {
			return nil, err
		}
		jsonSchemaType.AdditionalProperties = byt
	}

	for _, fieldDesc := range msg.GetField() {
		if fieldDesc.OneofIndex != nil {
			continue
		}
		childJsonTypes, err := g.convertField(curPkg, fieldDesc, msg)
		if err != nil {
			logger.Errorf("Failed to convert field %s in %s: %v", fieldDesc.GetName(), msg.GetName(), err)
			return jsonSchemaType, err
		}
		jsonSchemaType.Properties[fieldDesc.GetJsonName()] = childJsonTypes
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
		pkg, ok := g.protoPackage.LookupPackage(file.GetPackage())
		if !ok {
			return nil, fmt.Errorf("no such package found: %s", file.GetPackage())
		}
		eg := errgroup.Group{}
		for _, msg := range file.GetMessageType() {
			msg := msg
			eg.Go(func() error {
				jsonSchemaFileName := fmt.Sprintf("%s.jsonschema", msg.GetName())
				logger.Infof("Generating JSON-schema for MESSAGE (%v) in file [%v] => %v", msg.GetName(), protoFileName, jsonSchemaFileName)
				messageJSONSchemas, err := g.recursivelyConvertFields(pkg, msg)
				if err != nil {
					logger.Errorf("Failed to convert %s: %v", protoFileName, err)
					return err
				} else {
					response = append(response, messageJSONSchemas...)
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return nil, err
		}
	}

	return response, nil
}

func (g *generator) recursivelyConvertFields(curPkg *prototree.ProtoPackage, msg *descriptor.DescriptorProto) ([]*jsonschema.Type, error) {
	result := make([]*jsonschema.Type, 0, len(msg.GetNestedType())+1)

	jsonSchema, err := g.convertMessageType(curPkg, msg)
	if err != nil {
		return nil, err
	}
	result = append(result, jsonSchema)
	for _, nested := range msg.GetNestedType() {
		jsonSchemas, err := g.recursivelyConvertFields(curPkg, nested)
		if err != nil {
			return nil, err
		}
		result = append(result, jsonSchemas...)
	}
	return result, nil
}

type Generator interface {
	Convert(resource *model.Resource) (*jsonschema.Type, error)
}

type schemaMap struct {
	data map[string]jsonschema.Type
	sync.RWMutex
}

func newSchemaMap() *schemaMap {
	return &schemaMap{data: make(map[string]jsonschema.Type)}
}

func (s *schemaMap) Get(key string) (*jsonschema.Type, bool) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.data[key]
	return &val, ok
}

func (s *schemaMap) Set(key string, val *jsonschema.Type) {
	s.Lock()
	defer s.Unlock()
	// if _, ok := s.data[key]; ok {
	// 	return
	// }
	s.data[key] = *val
}

type Options struct {
	AllowNullValues              bool
	DisallowAdditionalProperties bool
	DisallowBigIntsAsStrings     bool
}

type generator struct {
	// Internal objects used to construct schema types, and build kube schemas
	protoPackage   *prototree.ProtoPackage
	fs             afero.Fs
	generatedTypes *schemaMap
	opts           *Options
}

func (g *generator) Convert(resource *model.Resource) (*jsonschema.Type, error) {
	preGenType, ok := g.generatedTypes.Get(schemaTitleName(resource.ProtoPackage, resource.Name))
	if !ok {
		return nil, NoPreviouslyRegisteredTypeError(schemaTitleName(resource.ProtoPackage, resource.Name))
	}
	allTypesToAdd, err := g.buildKubeSpec(preGenType)
	if err != nil {
		return nil, err
	}

	definitions := make(jsonschema.Definitions, len(allTypesToAdd))
	for _, v := range allTypesToAdd {
		nestedPreGenType, ok := g.generatedTypes.Get(v.Title)
		if !ok {
			return nil, NoPreviouslyRegisteredTypeError(v.Title)
		}
		definitions[v.Title] = nestedPreGenType
	}
	definitions[preGenType.Title] = preGenType
	result := &jsonschema.Type{
		Version:     jsonschema.Version,
		Ref:         schemaRefName(resource.ProtoPackage, resource.Name),
		Definitions: definitions,
	}
	return result, nil
}

func Marshal(schemaType *jsonschema.Type) ([]byte, error) {
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
		preGenType, ok := g.generatedTypes.Get(val.Title)
		if !ok {
			continue
		}
		pregenTypes = append(pregenTypes, preGenType)
	}
	if len(pregenTypes) == 0 {
		result = append(result, recursedType)
	} else {
		result = append(result, pregenTypes...)
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

func NewGenerator(req *plugin.CodeGeneratorRequest, opts *Options) (*generator, error) {
	if opts == nil {
		opts = defaultOptions
	}
	g := &generator{protoPackage: prototree.NewProtoTree(context.TODO()), fs: afero.NewOsFs(), generatedTypes: newSchemaMap(), opts: opts}
	wg := &sync.WaitGroup{}
	for _, file := range req.GetProtoFile() {
		for _, msg := range file.GetMessageType() {
			file, msg := file, msg
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				logger.Debugf("Loading a message type %s from package %s", msg.GetName(), file.GetPackage())
				g.protoPackage.RegisterMessage(file.Package, msg)
			}(wg)
		}
	}
	wg.Wait()

	for _, file := range req.GetProtoFile() {
		logger.Debugf("Converting file (%v)", file.GetName())
		types, err := g.convertFile(file)
		if err != nil {
			return nil, err
		}
		for _, v := range types {
			g.generatedTypes.Set(v.Title, v)
		}
	}
	return g, nil
}

func TranformToKubeSpec(schemaType *jsonschema.Type, res *model.Resource) (*jsonschema.Type, error) {
	var (
		resourceSpec *jsonschema.Type
		key          string
	)
	for k, v := range schemaType.Definitions {
		if v.Title == schemaTitleName(res.ProtoPackage, res.Original.GetName()) {
			resourceSpec = v
			key = k
		}
	}
	if resourceSpec == nil {
		return nil, NotCustomResourceJsonError(schemaType.Title)
	}
	resourceType := &jsonschema.Type{
		Title:      resourceSpec.Title,
		Ref:        resourceSpec.Ref,
		Properties: resourceSpec.Properties,
	}

	resourceSpec.Title = "spec"
	resourceSpec.Ref = ""
	result := make(map[string]*jsonschema.Type)
	for k, v := range resourceSpec.Properties {
		if !(k == "metadata" || k == "status") {
			result[k] = v
		}
	}
	resourceSpec.Properties = result

	result = make(map[string]*jsonschema.Type)
	for k, v := range resourceType.Properties {
		if k == "metadata" || k == "status" {
			result[k] = v
		}
	}
	resourceType.Properties = result

	resourceType.Properties["spec"] = resourceSpec
	schemaType.Definitions[key] = resourceType
	return schemaType, nil
}
