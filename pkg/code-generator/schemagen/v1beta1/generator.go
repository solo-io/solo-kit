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
	GetValidationSchema(resource model.Resource, specSchema *openapi.OrderedMap) (*apiextv1beta1.CustomResourceValidation, error)
}

func NewValidationSchemaGenerator() ValidationSchemaGenerator {
	return &validationSchemaGenerator{}
}

type validationSchemaGenerator struct {
}

func (g *validationSchemaGenerator) GetValidationSchema(resource model.Resource, specSchema *openapi.OrderedMap) (*apiextv1beta1.CustomResourceValidation, error) {
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

func getJsonSchema(resource model.Resource, schema *openapi.OrderedMap) (*apiextv1beta1.JSONSchemaProps, error) {
	if schema == nil {
		return nil, eris.Errorf("no open api schema for %s", resource.Name)
	}

	byt, err := schema.MarshalJSON()
	if err != nil {
		return nil, eris.Errorf("Cannot marshal OpenAPI schema for %v: %v", resource.Name, err)
	}

	jsonSchema := &apiextv1beta1.JSONSchemaProps{}
	if err = json.Unmarshal(byt, jsonSchema); err != nil {
		return nil, eris.Errorf("Cannot unmarshal raw OpenAPI schema to JSONSchemaProps for %v: %v", resource.Name, err)
	}

	if err = validateStructural(jsonSchema); err != nil {
		return nil, err
	}

	return jsonSchema, nil
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
