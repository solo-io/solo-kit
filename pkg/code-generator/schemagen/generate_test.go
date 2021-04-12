package schemagen_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen/v1beta1"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen/v1beta1/mocks"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

var _ = Describe("ValidationSchemaGenerator", func() {

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
			Options:                   options,
			ValidationSchemaGenerator: mockVersionedSchemaGenerator,
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
				SchemaOptionsByName: make(map[string]*v1beta1.SchemaOptions),
			}
		})

		It("does not run", func() {
			err := schemaGenerator.GenerateSchemasForResources([]*model.Resource{{
				Name: "test-resource",
			}})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	FWhen("valid CRDs are provided", func() {

		var (
			customConfigSchemaCompleted = false
		)

		BeforeEach(func() {
			customConfigCRD, err := v1beta1.GetCRDFromFile("v1beta1/fixtures/source/cc.yaml")
			Expect(err).NotTo(HaveOccurred())

			options = &schemagen.ValidationSchemaOptions{
				SchemaOptionsByName: map[string]*v1beta1.SchemaOptions{
					"customconfigs.test.gloo.solo.io": {
						OriginalCrd: customConfigCRD,
						OnSchemaComplete: func(crdWithSchema apiextv1beta1.CustomResourceDefinition) error {
							customConfigSchemaCompleted = true
							return nil
						},
					},
				},
			}
		})

		It("does not call GenerateValidationSchema for resources that do not match any CRDs", func() {
			err := schemaGenerator.GenerateSchemasForResources([]*model.Resource{{
				Name: "test-resource",
			}})
			Expect(err).NotTo(HaveOccurred())
			Expect(customConfigSchemaCompleted).To(BeFalse())
		})

		It("does call GenerateValidationSchema for resources that do match a CRD", func() {
			mockVersionedSchemaGenerator.EXPECT().ApplyValidationSchema(gomock.Any(), options.SchemaOptionsByName["customconfigs.test.gloo.solo.io"])

			err := schemaGenerator.GenerateSchemasForResources([]*model.Resource{{
				Name: "customconfigs.test.gloo.solo.io",
			}})
			Expect(err).NotTo(HaveOccurred())
			Expect(customConfigSchemaCompleted).To(BeTrue())
		})

	})

})
