package schemagen_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen/v1beta1"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen/v1beta1/mocks"
)

var _ = Describe("SchemaGenerator", func() {

	var (
		controller *gomock.Controller

		schemaGenerator              schemagen.SchemaGenerator
		options                      *schemagen.ValidationSchemaOptions
		mockVersionedSchemaGenerator *mocks.MockSchemaGenerator
	)

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())

		options = nil
		mockVersionedSchemaGenerator = mocks.NewMockSchemaGenerator(controller)
	})

	JustBeforeEach(func() {
		schemaGenerator = schemagen.SchemaGenerator{
			Options:                  options,
			VersionedSchemaGenerator: mockVersionedSchemaGenerator,
		}
	})

	When("options are nil", func() {
		BeforeEach(func() {
			options = nil
		})

		It("does not run", func() {
			err := schemaGenerator.GenerateSchemasForResources([]*model.Resource{{
				Name: "test-resource",
			}})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("no CRDs are provided", func() {
		BeforeEach(func() {
			options = &schemagen.ValidationSchemaOptions{
				SchemaOptionsByName: make(map[string]v1beta1.SchemaOptions),
			}
		})

		It("does not run", func() {
			err := schemaGenerator.GenerateSchemasForResources([]*model.Resource{{
				Name: "test-resource",
			}})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("resource does not match CRDs", func() {

		It("does not generate schema for resource", func() {
			// TODO
		})

	})

})
