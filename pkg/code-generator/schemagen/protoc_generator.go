package schemagen

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rotisserie/eris"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/ghodss/yaml"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/errors"
)

// TODO (sam-heilbron)
// This is failing on validateStructural. Use this once it is fixed
func GenerateJsonSchemasUsingProtoc(project *model.Project, importsCollector collector.Collector) (map[schema.GroupVersionKind]*v1beta1.JSONSchemaProps, error) {
	// Use a tmp directory as the output of schemas
	// The schemas will then be matched with the appropriate CRD
	tmpOutputDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpOutputDir)

	// Use a directory that is specific to this project
	// This ensures that when we traverse the outputDir, we only traverse project specific schemas
	projectOutputDir := filepath.Join(tmpOutputDir, project.String())
	_ = os.MkdirAll(projectOutputDir, os.ModePerm)

	// Generate the validation schemas
	generator := &protocGenerator{
		tmpSchemaOutputDir: tmpOutputDir,
		project:            project,
		importsCollector:   importsCollector,
		protocExecutor: &collector.OpenApiProtocExecutor{
			OutputDir: tmpOutputDir,
			Project:   project,
		},
	}
	return generator.Generate()
}

type protocGenerator struct {
	// The path to the tmp directory used to write schemas to
	tmpSchemaOutputDir string
	// The project to create schemas for
	project *model.Project
	// The Collector used to extract imports for proto files
	importsCollector collector.Collector
	// The Executor used to compile protos
	protocExecutor collector.ProtocExecutor
}

func (p *protocGenerator) Generate() (map[schema.GroupVersionKind]*v1beta1.JSONSchemaProps, error) {
	absoluteRoot, err := filepath.Abs(anyvendor.DefaultDepDir)
	if err != nil {
		return nil, err
	}

	// 1. Generate the openApiSchemas for the project, writing them to a temp directory (tmpSchemaOutputDir)
	for _, projectProto := range p.project.ProjectConfig.ProjectProtos {
		if err := p.generateSchemasForProjectProto(absoluteRoot, projectProto); err != nil {
			return nil, err
		}
	}

	// 2. Walk the tmpSchemaOutputDir and convert the open api schemas into JSONSchemaProps
	return p.processGeneratedSchemas()
}

func (p *protocGenerator) generateSchemasForProjectProto(root, projectProtoFile string) error {
	imports, err := p.importsCollector.CollectImportsForFile(root, filepath.Join(root, projectProtoFile))
	if err != nil {
		return errors.Wrapf(err, "collecting imports for proto file")
	}

	// we don't use the output of protoc so use a temp file
	tmpFile, err := ioutil.TempFile("", "schema-gen-")
	if err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	return p.protocExecutor.Execute(projectProtoFile, tmpFile.Name(), imports)
}

func (p *protocGenerator) processGeneratedSchemas() (map[schema.GroupVersionKind]*v1beta1.JSONSchemaProps, error) {
	jsonSchemasByGVK := make(map[schema.GroupVersionKind]*v1beta1.JSONSchemaProps)
	err := filepath.Walk(p.tmpSchemaOutputDir, func(schemaFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(schemaFile, ".yaml") {
			return nil
		}

		log.Debugf("Generated Schema File: %s", schemaFile)
		doc, err := p.readOpenApiDocumentFromFile(schemaFile)
		if err != nil {
			// Stop traversing the output directory
			return err
		}

		schemas := doc.Components.Schemas
		if schemas == nil {
			// Continue traversing the output directory
			return nil
		}

		for schemaKey, schemaValue := range schemas {
			schemaGVK := p.getGVKForSchemaKey(schemaKey)

			// Spec validation schema
			specJsonSchema, err := getJsonSchema(schemaKey, schemaValue)
			if err != nil {
				return err
			}

			jsonSchemasByGVK[schemaGVK] = specJsonSchema
		}
		// Continue traversing the output directory
		return nil
	})

	return jsonSchemasByGVK, err
}

func (p *protocGenerator) readOpenApiDocumentFromFile(file string) (*openapi3.Swagger, error) {
	var openApiDocument *openapi3.Swagger
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "reading file")
	}
	if err := yaml.Unmarshal(bytes, &openApiDocument); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling tmp file as schemas")
	}
	return openApiDocument, nil
}

func (p *protocGenerator) getGVKForSchemaKey(schemaKey string) schema.GroupVersionKind {
	// The generated keys look like testing.solo.io.MockResource
	// The kind is the `MockResource` portion
	ss := strings.Split(schemaKey, ".")
	kind := ss[len(ss)-1]

	projectGV := model.GetGVForProject(p.project)

	return schema.GroupVersionKind{
		Group:   projectGV.Group,
		Version: projectGV.Version,
		Kind:    kind,
	}
}

func getJsonSchema(schemaKey string, schema *openapi3.SchemaRef) (*v1beta1.JSONSchemaProps, error) {
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

	// remove 'properties' and 'required' fields to prevent validating proto.Any fields
	removeProtoAnyValidation(obj)

	// TODO (sam-heilbron) - Determine the proper way to do this
	removeProtoMetadataValidation(obj)

	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	jsonSchema := &v1beta1.JSONSchemaProps{}
	if err = json.Unmarshal(bytes, jsonSchema); err != nil {
		return nil, eris.Errorf("Cannot unmarshal raw OpenAPI schema to JSONSchemaProps for %v: %v", schemaKey, err)
	}

	return jsonSchema, nil
}

// prevent k8s from validating metadata field
// TODO - add details for why
func removeProtoMetadataValidation(d map[string]interface{}) {
	for _, v := range d {
		values, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		_, hasProperties := values["properties"]
		_, hasMetadata := values["metadata"]

		if hasMetadata && !hasProperties {
			delete(values, "metadata")
		}
	}
}

// prevent k8s from validating proto.Any fields (since it's unstructured)
func removeProtoAnyValidation(d map[string]interface{}) {
	for _, v := range d {
		values, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		desc, ok := values["properties"]
		properties, isObj := desc.(map[string]interface{})
		// detect proto.Any field from presence of "@type" as field under "properties"
		if !ok || !isObj || properties["type_url"] == nil {
			removeProtoAnyValidation(values)
			continue
		}
		// remove "properties" value
		delete(values, "properties")
		// remove "required" value
		delete(values, "required")
		// x-kubernetes-preserve-unknown-fields allows for unknown fields from a particular node
		// see https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema
		values["x-kubernetes-preserve-unknown-fields"] = true
	}
}
