package schemagen

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/openapi"
	"cuelang.org/go/encoding/protobuf"
	"github.com/rotisserie/eris"
	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Implementation of JsonSchemaGenerator that uses cuelang (https://github.com/cuelang/cue)
// packages to translate protobuf definitions into OpenAPI schemas with structural constraints
// It is modeled after the istio tool: https://github.com/istio/tools/tree/master/cmd/cue-gen
// TODO (sam-heilbron) This tool has a known flaw that makes it unusable with Gloo Edge protos
// 	 There are open issues to track this and a GitHub discussion with more context:
//	 https://github.com/cuelang/cue/discussions/944. We would like to move towards this
//	 implementation and have added it now to compare schemas generated with this implementation
//	 to schemas generated with other implementations to ensure consistency.
type cueGenerator struct {
	// The Collector used to extract imports for proto files
	importsCollector collector.Collector
	absoluteRoot     string
	protoDir         string
}

// TODO (sam-heilbron) CUE deprecates OrderedMap
// 	 We should start using ast.File and other APIs in the ast package to manipulate the schemas
//	 This requires some more support from ast package thus leaving this as to-do

// Mapping from protobuf message name to OpenApi schema
type OpenApiSchemas map[string]*openapi.OrderedMap

func NewCueGenerator(importsCollector collector.Collector, absoluteRoot string) *cueGenerator {
	return &cueGenerator{
		importsCollector: importsCollector,
		absoluteRoot:     absoluteRoot,
		protoDir:         anyvendor.DefaultDepDir,
	}
}

func (c *cueGenerator) GetJsonSchemaForProject(project *model.Project) (map[schema.GroupVersionKind]*apiextv1.JSONSchemaProps, error) {
	protobufExtractor, err := c.getProtobufExtractorForProject(project)
	if err != nil {
		return nil, err
	}

	instances, err := protobufExtractor.Instances()
	if err != nil {
		return nil, err
	}

	// Convert cuelang to openapi
	openApiGenerator := &openapi.Generator{
		// k8s structural schemas do not allow $refs, i.e. all references must be expanded
		ExpandReferences: true,
	}

	openApiSchemas := OpenApiSchemas{}
	built := cue.Build(instances)
	for _, builtInstance := range built {
		// Avoid generating openapi for irrelevant proto imports.
		if !strings.HasSuffix(builtInstance.ImportPath, project.ProjectConfig.GoPackage) {
			continue
		}

		if builtInstance.Err != nil {
			return nil, err
		}
		if err = builtInstance.Value().Validate(); err != nil {
			return nil, eris.Errorf("Cue instance validation failed for %s: %+v", project.ProtoPackage, err)
		}
		schemas, err := openApiGenerator.Schemas(builtInstance)
		if err != nil {
			return nil, eris.Errorf("Cue openapi generation failed for %s: %+v", project.ProtoPackage, err)
		}

		// Iterate openapi objects to construct mapping from proto message name to openapi schema
		for _, kv := range schemas.Pairs() {
			openApiSchemas[kv.Key] = kv.Value.(*openapi.OrderedMap)
		}

		return c.convertOpenApiSchemasToJsonSchemas(project, openApiSchemas)
	}

	return nil, nil
}

func (c *cueGenerator) getProtobufExtractorForProject(project *model.Project) (*protobuf.Extractor, error) {
	// Collect all protobuf definitions including transitive dependencies.
	var imports []string
	protoRoot := filepath.Join(c.absoluteRoot, c.protoDir)

	for _, projectProto := range project.ProjectConfig.ProjectProtos {
		importsForResource, err := c.importsCollector.CollectImportsForFile(protoRoot, filepath.Join(protoRoot, projectProto))
		if err != nil {
			return nil, err
		}
		imports = append(imports, importsForResource...)
	}
	imports = stringutils.Unique(imports)

	// Parse protobuf into cuelang
	protobufExtractor := protobuf.NewExtractor(&protobuf.Config{
		Root:   protoRoot,
		Module: project.ProjectConfig.GoPackage,
		Paths:  imports,
	})

	for _, projectProto := range project.ProjectConfig.ProjectProtos {
		if err := protobufExtractor.AddFile(projectProto, nil); err != nil {
			return nil, err
		}
	}
	return protobufExtractor, nil
}

func (c *cueGenerator) convertOpenApiSchemasToJsonSchemas(project *model.Project, schemas OpenApiSchemas) (map[schema.GroupVersionKind]*apiextv1.JSONSchemaProps, error) {
	jsonSchemasByGVK := make(map[schema.GroupVersionKind]*apiextv1.JSONSchemaProps)

	for schemaKey, schemaValue := range schemas {
		schemaGVK := c.getGVKForSchemaKey(project, schemaKey)

		// Spec validation schema
		specJsonSchema, err := c.convertOpenApiSchemaToJsonSchema(schemaKey, schemaValue)
		if err != nil {
			return nil, err
		}

		jsonSchemasByGVK[schemaGVK] = specJsonSchema
	}

	return jsonSchemasByGVK, nil
}

func (c *cueGenerator) convertOpenApiSchemaToJsonSchema(schemaKey string, schema *openapi.OrderedMap) (*apiextv1.JSONSchemaProps, error) {
	if schema == nil {
		return nil, eris.Errorf("no open api schema for %s", schemaKey)
	}

	oApiJson, err := schema.MarshalJSON()
	if err != nil {
		return nil, eris.Errorf("Cannot marshal OpenAPI schema for %v: %v", schemaKey, err)
	}

	var obj map[string]interface{}
	if err = json.Unmarshal(oApiJson, &obj); err != nil {
		return nil, err
	}

	// detect proto.Any field from presence of "@type" as field under "properties"
	removeProtoAnyValidation(obj, "@type")

	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	jsonSchema := &apiextv1.JSONSchemaProps{}
	if err = json.Unmarshal(bytes, jsonSchema); err != nil {
		return nil, eris.Errorf("Cannot unmarshal raw OpenAPI schema to JSONSchemaProps for %v: %v", schemaKey, err)
	}

	return jsonSchema, nil
}

func (c *cueGenerator) getGVKForSchemaKey(project *model.Project, schemaKey string) schema.GroupVersionKind {
	// The generated keys look like testing.solo.io.MockResource
	// The kind is the `MockResource` portion
	ss := strings.Split(schemaKey, ".")
	kind := ss[len(ss)-1]

	projectGV := model.GetGVForProject(project)

	return schema.GroupVersionKind{
		Group:   projectGV.Group,
		Version: projectGV.Version,
		Kind:    kind,
	}
}
