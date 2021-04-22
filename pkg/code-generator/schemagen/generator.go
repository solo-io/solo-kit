package schemagen

import (
	"log"
	"time"

	"github.com/solo-io/solo-kit/pkg/code-generator/metrics"

	"github.com/solo-io/solo-kit/pkg/code-generator/collector"

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

func GenerateProjectValidationSchema(
	project *model.Project,
	options *ValidationSchemaOptions,
	absoluteRoot string,
	importsCollector collector.Collector) error {
	defer metrics.MeasureProjectElapsed(project, "schema-gen", time.Now())

	log.Printf("Running schemagen for project: %v", model.GetGVForProject(project))

	schemaOptionsByGVK := getSchemaOptionsByGVKForProject(project, options)
	if len(schemaOptionsByGVK) == 0 {
		log.Printf("Skipping schemagen for project: %v. No CRDs found matching Project.Group", model.GetGVForProject(project))
		return nil
	}

	// Step 1. Generate the open api schemas for the project
	openApiGenerator := NewCueOpenApiSchemaGenerator(importsCollector, anyvendor.DefaultDepDir, absoluteRoot)
	openApiSchemas, err := openApiGenerator.GetOpenApiSchemas(project)
	if err != nil {
		return err
	}

	// Step 2. Convert the open api schemas into validation schemas
	p := &SchemaGenerator{
		SchemaOptionsByGVK:        schemaOptionsByGVK,
		ValidationSchemaGenerator: v1beta1.NewValidationSchemaGenerator(),
	}
	return p.GenerateSchemasForProject(project, openApiSchemas)
}

func getSchemaOptionsByGVKForProject(project *model.Project, options *ValidationSchemaOptions) map[kubeschema.GroupVersionKind]*v1beta1.SchemaOptions {
	// Map the schema options by the GVK of the CRD
	// This is the key we will use to associate resources with CRDs
	schemaOptionsByGVK := map[kubeschema.GroupVersionKind]*v1beta1.SchemaOptions{}

	if options == nil || len(options.SchemaOptions) == 0 {
		// No schemagen was configured
		return schemaOptionsByGVK
	}

	if len(project.ProjectConfig.ProjectProtos) == 0 {
		// project has no protos, these are used to generate the schemas
		return schemaOptionsByGVK
	}

	// Use the project Group to match with the CRD Group,Version
	projectGV := model.GetGVForProject(project)

	for _, crdSchemaOptions := range options.SchemaOptions {
		crdSpec := crdSchemaOptions.OriginalCrd.Spec
		crdGVK := kubeschema.GroupVersionKind{
			Group:   crdSpec.Group,
			Version: crdSpec.Version,
			Kind:    crdSpec.Names.Kind,
		}

		// If the group matches, this project is responsible for building the schema of this CRD
		if crdGVK.Group == projectGV.Group {
			schemaOptionsByGVK[crdGVK] = crdSchemaOptions
		}
	}
	return schemaOptionsByGVK
}

type SchemaGenerator struct {
	SchemaOptionsByGVK        map[kubeschema.GroupVersionKind]*v1beta1.SchemaOptions
	ValidationSchemaGenerator v1beta1.ValidationSchemaGenerator
}

func (p *SchemaGenerator) GenerateSchemasForProject(project *model.Project, schemas OpenApiSchemas) error {
	var postApplyFuncs []func() error
	for _, res := range project.Resources {

		// Try to associate the resource with a CRD
		schemaOptions, ok := p.SchemaOptionsByGVK[model.GetGVKForResource(res)]
		if ok {
			specSchema := schemas[res.Original.GetName()]

			// Build the CustomResourceValidation for the CRD
			validationSchema, err := p.ValidationSchemaGenerator.GetValidationSchema(res, specSchema)
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

	// Apply the CRDs with Schema
	// We do this in separate steps to ensure that all CRDs pass/fail collectively
	var multiErr *multierror.Error
	for _, postApply := range postApplyFuncs {
		err := postApply()
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	return multiErr.ErrorOrNil()
}
