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

type ProjectValidationSchemaOptions struct {
	SkipGeneration          bool
	GenerationTimeout       time.Duration
	SchemaGenerationOptions []*v1beta1.SchemaOptions
}

type ValidationSchemaOptions struct {
	// SchemaOptions per Project, indexed by the Project Group (ie. gloo.solo.io)
	ProjectSchemaOptions map[string]*ProjectValidationSchemaOptions
}

func GenerateProjectValidationSchema(
	project *model.Project,
	options *ValidationSchemaOptions,
	absoluteRoot string,
	importsCollector collector.Collector) error {
	projectGV := model.GetGVForProject(project)

	if options == nil || len(options.ProjectSchemaOptions) == 0 {
		// No schemagen was configured
		return nil
	}

	projectSchemaOptions, ok := options.ProjectSchemaOptions[projectGV.Group]
	if !ok {
		// No schemagen was configured for this project
		return nil
	}

	if projectSchemaOptions.SkipGeneration {
		// Project configured to skip schemagen
		return nil
	}

	schemaOptionsByGVK := map[kubeschema.GroupVersionKind]*v1beta1.SchemaOptions{}
	for _, crdSchemaOptions := range projectSchemaOptions.SchemaGenerationOptions {
		crdSpec := crdSchemaOptions.OriginalCrd.Spec
		crdGVK := kubeschema.GroupVersionKind{
			Group:   crdSpec.Group,
			Version: crdSpec.Version,
			Kind:    crdSpec.Names.Kind,
		}
		schemaOptionsByGVK[crdGVK] = crdSchemaOptions
	}
	if len(schemaOptionsByGVK) == 0 {
		// No options defined for project
		return nil
	}

	log.Printf("Running schemagen for project: %v", projectGV)
	defer metrics.MeasureProjectElapsed(project, "schema-gen", time.Now())

	// Step 1. Generate the open api schemas for the project
	openApiGenerator := NewCueOpenApiSchemaGenerator(importsCollector, anyvendor.DefaultDepDir, absoluteRoot)
	timeout := projectSchemaOptions.GenerationTimeout
	if timeout == 0 {
		timeout = time.Second * 60
	}
	openApiSchemas, err := openApiGenerator.GetOpenApiSchemas(project, projectSchemaOptions.GenerationTimeout)
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
