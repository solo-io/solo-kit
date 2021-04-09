package schemagen

import (
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen/v1beta1"
)

type ValidationSchemaOptions struct {
	SchemaOptionsByName map[string]v1beta1.SchemaOptions
}

func GenerateProjectValidationSchema(project *model.Project, options *ValidationSchemaOptions) error {
	schemaGenerator := v1beta1.NewCueGenerator(project)

	p := &SchemaGenerator{
		Options:                  options,
		VersionedSchemaGenerator: schemaGenerator,
	}
	return p.GenerateSchemasForResources(project.Resources)
}

type SchemaGenerator struct {
	Options                  *ValidationSchemaOptions
	VersionedSchemaGenerator v1beta1.SchemaGenerator
}

func (p *SchemaGenerator) GenerateSchemasForResources(resources []*model.Resource) error {
	if !p.shouldRun() {
		return nil
	}

	var postApplyFuncs []func() error

	// Step 1. Generate the CRDs with Schema
	for _, res := range resources {
		schemaOptions, ok := p.Options.SchemaOptionsByName[res.Name]
		if ok {
			// Generate the validation schema
			crdWithSchema, err := p.VersionedSchemaGenerator.ApplyValidationSchema(res, schemaOptions)
			if err != nil {
				return err
			}

			// Append the OnSchemaComplete to be run at the end
			postApplyFuncs = append(postApplyFuncs, func() error {
				return schemaOptions.OnSchemaComplete(crdWithSchema)
			})

		}
	}

	// Step 2. Apply the CRDs with Schema
	// We do this in separate steps to ensure that all CRDs pass/fail collectively
	var multiErr *multierror.Error
	for _, postApply := range postApplyFuncs {
		err := postApply()
		multiErr = multierror.Append(multiErr, err)
	}

	return multiErr
}

func (p *SchemaGenerator) shouldRun() bool {
	if p.Options == nil {
		return false
	}

	if len(p.Options.SchemaOptionsByName) == 0 {
		return false
	}

	return true
}
