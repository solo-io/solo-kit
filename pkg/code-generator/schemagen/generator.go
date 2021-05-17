package schemagen

import (
	"fmt"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

type ValidationSchemaOptions struct {
	// Path to the directory where CRDs will be read from and written to
	CrdDirectory string
}

type JsonSchemaGenerator interface {
	GetJsonSchemaForProject(project *model.Project) (map[schema.GroupVersionKind]*v1beta1.JSONSchemaProps, error)
}

func GenerateOpenApiValidationSchemas(project *model.Project, options *ValidationSchemaOptions, jsonSchemaGenerator JsonSchemaGenerator) error {
	if options == nil || options.CrdDirectory == "" {
		log.Debugf("No CRDDirectory provided, skipping schema-gen")
		return nil
	}

	if !project.ProjectConfig.GenKubeValidationSchemas {
		log.Debugf("Project %s not configured to generate validation schema", project.String())
		return nil
	}

	// Extract the CRDs from the directory
	crds, err := GetCRDsFromDirectory(options.CrdDirectory)
	if err != nil {
		return err
	}

	if len(crds) == 0 {
		log.Debugf("Found 0 CRDs in directory: %s, skipping schema-gen", options.CrdDirectory)
		return nil
	}

	// Build the JsonSchemas for the project
	jsonSchemasByGVK, err := jsonSchemaGenerator.GetJsonSchemaForProject(project)
	if err != nil {
		return err
	}

	// For each matching CRD, apply the JSON schema to that CRD
	// Use Group.Version.Kind to match CRDs and Schemas
	crdWriter := NewCrdWriter(options.CrdDirectory)
	for _, crd := range crds {
		crdGVK := schema.GroupVersionKind{
			Group:   crd.Spec.Group,
			Version: crd.Spec.Version,
			Kind:    crd.Spec.Names.Kind,
		}

		specJsonSchema, ok := jsonSchemasByGVK[crdGVK]
		if !ok {
			continue
		}

		if err := validateStructural(crdGVK, specJsonSchema); err != nil {
			return err
		}

		validationSchema := &v1beta1.CustomResourceValidation{
			OpenAPIV3Schema: &v1beta1.JSONSchemaProps{
				Type:       "object",
				Properties: map[string]v1beta1.JSONSchemaProps{},
			},
		}
		validationSchema.OpenAPIV3Schema.Properties["spec"] = *specJsonSchema

		if err = crdWriter.ApplyValidationSchemaToCRD(crd, validationSchema); err != nil {
			return err
		}
	}

	return nil
}

// Lifted from https://github.com/istio/tools/blob/477454adf7995dd3070129998495cdc8aaec5aff/cmd/cue-gen/crd.go#L108
func validateStructural(gvk schema.GroupVersionKind, s *v1beta1.JSONSchemaProps) error {
	out := &apiext.JSONSchemaProps{}
	if err := v1beta1.Convert_v1beta1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(s, out, nil); err != nil {
		return fmt.Errorf("%v cannot convert v1beta1 JSONSchemaProps to JSONSchemaProps: %v", gvk, err)
	}

	r, err := structuralschema.NewStructural(out)
	if err != nil {
		return fmt.Errorf("%v cannot convert to a structural schema: %v", gvk, err)
	}

	if errs := structuralschema.ValidateStructural(nil, r); len(errs) != 0 {
		return fmt.Errorf("%v schema is not structural: %v", gvk, errs.ToAggregate().Error())
	}

	return nil
}
