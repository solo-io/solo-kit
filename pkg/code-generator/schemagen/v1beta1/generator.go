package v1beta1

import (
	"errors"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

//go:generate mockgen -destination mocks/mock_schema_generator.go -package mocks -source generator.go SchemaGenerator

type SchemaOptions struct {
	OriginalCrd           apiextv1beta1.CustomResourceDefinition
	PreserveUnknownFields bool
	OnSchemaComplete      func(crdWithSchema apiextv1beta1.CustomResourceDefinition) error
}

type SchemaGenerator interface {
	ApplyValidationSchema(resource *model.Resource, options SchemaOptions) (apiextv1beta1.CustomResourceDefinition, error)
}

func NewCueGenerator(project *model.Project) SchemaGenerator {
	return &cueGenerator{
		project: project,
	}
}

type cueGenerator struct {
	project *model.Project
}

func (c *cueGenerator) ApplyValidationSchema(resource *model.Resource, options SchemaOptions) (apiextv1beta1.CustomResourceDefinition, error) {
	return apiextv1beta1.CustomResourceDefinition{}, errors.New("not implemented")
}
