package schemagen

import (
	"fmt"

	"k8s.io/utils/pointer"

	"github.com/solo-io/solo-kit/pkg/code-generator/collector"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

type ValidationSchemaOptions struct {
	// Path to the directory where CRDs will be read from and written to
	CrdDirectory string

	// Tool used to generate JsonSchemas, defaults to protoc
	JsonSchemaTool string
}

type JsonSchemaGenerator interface {
	GetJsonSchemaForProject(project *model.Project) (map[schema.GroupVersionKind]*apiextv1beta1.JSONSchemaProps, error)
}

func GenerateOpenApiValidationSchemas(project *model.Project, options *ValidationSchemaOptions, importsCollector collector.Collector, absoluteRoot string) error {
	if options == nil || options.CrdDirectory == "" {
		log.Debugf("No CRDDirectory provided, skipping schema-gen")
		return nil
	}

	if !project.ProjectConfig.GenKubeValidationSchemas {
		log.Debugf("Project %s not configured to generate validation schema", project.String())
		return nil
	}

	// Extract the CRDs from the directory
	crds, err := getCRDsFromDirectory(options.CrdDirectory)
	if err != nil {
		return err
	}

	if len(crds) == 0 {
		log.Debugf("Found 0 CRDs in directory: %s, skipping schema-gen", options.CrdDirectory)
		return nil
	}

	// Build the JsonSchemas for the project
	var jsonSchemaGenerator JsonSchemaGenerator
	switch options.JsonSchemaTool {
	case "cue":
		jsonSchemaGenerator = NewCueGenerator(importsCollector, absoluteRoot)
	case "protoc":
		jsonSchemaGenerator = NewProtocGenerator(importsCollector, absoluteRoot)
	default:
		jsonSchemaGenerator = NewProtocGenerator(importsCollector, absoluteRoot)
	}

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

		// prevent k8s from validating metadata field
		removeProtoMetadataValidation(specJsonSchema)

		if err := validateStructural(crdGVK, specJsonSchema); err != nil {
			return err
		}

		validationSchema := &apiextv1beta1.CustomResourceValidation{
			OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
				Type:       "object",
				Properties: map[string]apiextv1beta1.JSONSchemaProps{},
			},
		}

		// Either use the status defined on the spec, or a generic status
		statusSchema := specJsonSchema.Properties["status"]
		if statusSchema.Type == "" {
			statusSchema = apiextv1beta1.JSONSchemaProps{
				Type:                   "object",
				XPreserveUnknownFields: pointer.BoolPtr(true),
			}
		}

		validationSchema.OpenAPIV3Schema.Properties["spec"] = *specJsonSchema
		validationSchema.OpenAPIV3Schema.Properties["status"] = statusSchema

		if err = crdWriter.ApplyValidationSchemaToCRD(crd, validationSchema); err != nil {
			return err
		}
	}

	return nil
}

// Lifted from https://github.com/istio/tools/blob/477454adf7995dd3070129998495cdc8aaec5aff/cmd/cue-gen/crd.go#L108
func validateStructural(gvk schema.GroupVersionKind, s *apiextv1beta1.JSONSchemaProps) error {
	out := &apiext.JSONSchemaProps{}
	if err := apiextv1beta1.Convert_v1beta1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(s, out, nil); err != nil {
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

// prevent k8s from validating metadata field
// https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema
// "if metadata is specified, then only restrictions on metadata.name and metadata.generateName are allowed."
// The kube api server is responsible for managing the metadata field, so users are not allowed to define schemas on it.
// We remove validation altogether.
func removeProtoMetadataValidation(s *apiextv1beta1.JSONSchemaProps) {
	delete(s.Properties, "metadata")
}

// prevent k8s from validating proto.Any fields (since it's unstructured)
func removeProtoAnyValidation(d map[string]interface{}, propertyField string) {
	for _, v := range d {
		values, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		desc, ok := values["properties"]
		properties, isObj := desc.(map[string]interface{})
		// detect proto.Any field from presence of [propertyField] as field under "properties"
		if !ok || !isObj || properties[propertyField] == nil {
			removeProtoAnyValidation(values, propertyField)
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
