package schemagen

import (
	"log"

	"github.com/solo-io/anyvendor/anyvendor"
	"k8s.io/utils/pointer"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen/v1beta1"

	kubeschema "k8s.io/apimachinery/pkg/runtime/schema"
)

type ValidationSchemaOptions struct {
	SchemaOptions []*v1beta1.SchemaOptions
}

func GenerateProjectValidationSchema(project *model.Project, options *ValidationSchemaOptions) error {
	// Attempt to short circuit if we do not need to generate schemas
	if shouldSkipSchemaGetForProject(project, options) {
		log.Printf("Skipping schemagen for project: %v", model.GetGVForProject(project))
		return nil
	}

	log.Printf("Running schemagen for project: %v", model.GetGVForProject(project))

	// Map the schema options by the GVK of the CRD
	// This is the key we will use to associate resources with CRDs
	schemaOptionsByGVK := make(map[kubeschema.GroupVersionKind]*v1beta1.SchemaOptions, len(options.SchemaOptions))
	for _, crdSchemaOptions := range options.SchemaOptions {
		schemaOptionsByGVK[crdSchemaOptions.OriginalCrd.GroupVersionKind()] = crdSchemaOptions
	}

	p := &SchemaGenerator{
		SchemaOptionsByGVK:        schemaOptionsByGVK,
		OpenApiSchemaGenerator:    NewCueOpenApiSchemaGenerator(),
		ValidationSchemaGenerator: v1beta1.NewValidationSchemaGenerator(),
	}
	return p.GenerateSchemasForProject(project)
}

func shouldSkipSchemaGetForProject(project *model.Project, options *ValidationSchemaOptions) bool {
	if options == nil {
		return true
	}

	if len(options.SchemaOptions) == 0 {
		return true
	}

	// TODO - more checks and set to false by default
	// for now just never run
	return false
}

type SchemaGenerator struct {
	SchemaOptionsByGVK        map[kubeschema.GroupVersionKind]*v1beta1.SchemaOptions
	OpenApiSchemaGenerator    OpenApiSchemaGenerator
	ValidationSchemaGenerator v1beta1.ValidationSchemaGenerator
}

func (p *SchemaGenerator) GenerateSchemasForProject(project *model.Project) error {
	// Step 1. Generate the open api schemas for the project
	openApiSchemas, err := p.OpenApiSchemaGenerator.GetOpenApiSchemas(*project, anyvendor.DefaultDepDir)
	if err != nil {
		return err
	}
	if len(openApiSchemas) == 0 {
		// There were no open api schemas generated for this project, skip it
		return nil
	}

	// Step 2. Generate the schemas for the CRDs
	var postApplyFuncs []func() error
	for _, res := range project.Resources {

		// Try to associate the resource with a CRD
		schemaOptions, ok := p.SchemaOptionsByGVK[model.GetGVKForResource(*res)]
		if ok {
			// TODO - get proper schema
			specSchema := openApiSchemas[res.Original.GetName()]

			// Build the CustomResourceValidation for the CRD
			validationSchema, err := p.ValidationSchemaGenerator.GetValidationSchema(*res, specSchema)
			if err != nil {
				return err
			}

			// Take the original CRD and apply the validation schema
			crdWithSchema := schemaOptions.OriginalCrd
			crdWithSchema.Spec.Validation = validationSchema

			// Setting PreserveUnknownFields to false ensures that objects with unknown fields are rejected.
			// This is deprecated and will default to false in future versions.
			crdWithSchema.Spec.PreserveUnknownFields = pointer.BoolPtr(false)

			// Append the OnSchemaComplete to be run at the end
			postApplyFuncs = append(postApplyFuncs, func() error {
				return schemaOptions.OnSchemaComplete(crdWithSchema)
			})

		}
	}

	// Step 3. Apply the CRDs with Schema
	// We do this in separate steps to ensure that all CRDs pass/fail collectively
	var multiErr *multierror.Error
	for _, postApply := range postApplyFuncs {
		err := postApply()
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	return multiErr
}
