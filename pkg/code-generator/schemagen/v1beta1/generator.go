package v1beta1

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/encoding/openapi"
	"github.com/rotisserie/eris"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
)

type SchemaOptions struct {
	OriginalCrd      apiextv1beta1.CustomResourceDefinition
	OnSchemaComplete func(crdWithSchema apiextv1beta1.CustomResourceDefinition) error
}

type ValidationSchemaGenerator interface {
	GetValidationSchema(resource *model.Resource, specSchema *openapi.OrderedMap) (*apiextv1beta1.CustomResourceValidation, error)
}

func NewValidationSchemaGenerator() ValidationSchemaGenerator {
	return &validationSchemaGenerator{}
}

type validationSchemaGenerator struct {
}

func (g *validationSchemaGenerator) GetValidationSchema(resource *model.Resource, specSchema *openapi.OrderedMap) (*apiextv1beta1.CustomResourceValidation, error) {
	validationSchema := &apiextv1beta1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
			Type:       "object",
			Properties: map[string]apiextv1beta1.JSONSchemaProps{},
		},
	}

	// Spec validation schema
	specJsonSchema, err := getJsonSchema(resource, specSchema)
	if err != nil {
		return nil, eris.Wrapf(err, "constructing spec validation schema for Kind %s", resource.Name)
	}
	validationSchema.OpenAPIV3Schema.Properties["spec"] = *specJsonSchema

	return validationSchema, nil
}

func getJsonSchema(resource *model.Resource, schema *openapi.OrderedMap) (*apiextv1beta1.JSONSchemaProps, error) {
	if schema == nil {
		return nil, eris.Errorf("no open api schema for %s", resource.Name)
	}

	oApiJson, err := schema.MarshalJSON()
	if err != nil {
		return nil, eris.Errorf("Cannot marshal OpenAPI schema for %v: %v", resource.Name, err)
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

	jsonSchema := &apiextv1beta1.JSONSchemaProps{}
	if err = json.Unmarshal(bytes, jsonSchema); err != nil {
		return nil, eris.Errorf("Cannot unmarshal raw OpenAPI schema to JSONSchemaProps for %v: %v", resource.Name, err)
	}

	// TODO - validateStructural
	// this is failing due to the metadata
	if err = validateStructural(jsonSchema); err != nil {
		return nil, err
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
		if !ok || !isObj || properties["@type"] == nil {
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

// Lifted from https://github.com/istio/tools/blob/477454adf7995dd3070129998495cdc8aaec5aff/cmd/cue-gen/crd.go#L108
func validateStructural(s *apiextv1beta1.JSONSchemaProps) error {
	out := &apiext.JSONSchemaProps{}
	if err := apiextv1beta1.Convert_v1beta1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(s, out, nil); err != nil {
		return fmt.Errorf("cannot convert v1beta1 JSONSchemaProps to JSONSchemaProps: %v", err)
	}

	r, err := structuralschema.NewStructural(out)
	if err != nil {
		return fmt.Errorf("cannot convert to a structural schema: %v", err)
	}

	if errs := structuralschema.ValidateStructural(nil, r); len(errs) != 0 {
		return fmt.Errorf("schema is not structural: %v", errs.ToAggregate().Error())
	}

	return nil
}
