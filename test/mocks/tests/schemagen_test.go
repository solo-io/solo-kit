package tests_test

import (
	"path/filepath"

	"k8s.io/utils/pointer"

	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/solo-kit/pkg/utils/modutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("schemagen", func() {

	Context("JSONSchemaGenerator", func() {
		// At the moment we rely on a plugin for the protocol buffer compiler to generate
		// validation schemas for our CRDs.
		// We'd like to move towards relying on cuelang packages to support this feature.
		// Until we are able to do so, we want to ensure that the generated schemas
		// match the schemas we will generate using cue.
		// These tests verify that the generated schemas will match
		// NOTE: Since oneof's are treated differently between the two implementation's
		//	we rely on a resource that does not contain oneof's. This is intentional.

		var (
			cueGenerator    schemagen.JsonSchemaGenerator
			protocGenerator schemagen.JsonSchemaGenerator

			project                 *model.Project
			simpleMockResourceGVK   schema.GroupVersionKind
			validationSchemaOptions *schemagen.ValidationSchemaOptions
		)

		BeforeEach(func() {
			// This is a modified Project model, to only include the SimpleMockResource type
			project = &model.Project{
				ProjectConfig: model.ProjectConfig{
					Title:   "Solo-Kit Schemagen Testing",
					Name:    "testing.solo.io.v1",
					Version: "v1",
					ResourceGroups: map[string][]model.ResourceConfig{
						"testing.solo.io": {
							{
								ResourceName:    "SimpleMockResource",
								ResourcePackage: "testing.solo.io.v1",
							},
						},
					},
					GoPackage: "github.com/solo-io/solo-kit/test/mocks/v1",
					ProjectProtos: []string{
						"github.com/solo-io/solo-kit/test/mocks/api/v1/simple_mock_resources.proto",
					},
				},
				ProtoPackage: "testing.solo.io.v1",
			}

			// This is the resource we've configured to test various schemagen behaviors
			simpleMockResourceGVK = schema.GroupVersionKind{
				Group:   "testing.solo.io.v1",
				Version: "v1",
				Kind:    "SimpleMockResource",
			}

			validationSchemaOptions = &schemagen.ValidationSchemaOptions{}
		})

		JustBeforeEach(func() {
			soloKitGoMod, err := modutils.GetCurrentModPackageFile()
			Expect(err).NotTo(HaveOccurred())
			soloKitRoot := filepath.Dir(soloKitGoMod)

			commonImports := []string{
				filepath.Join(soloKitRoot, anyvendor.DefaultDepDir),
			}
			importsCollector := collector.NewCollector([]string{}, commonImports)

			cueGenerator = schemagen.NewCueGenerator(importsCollector, soloKitRoot)
			protocGenerator = schemagen.NewProtocGenerator(importsCollector, soloKitRoot, validationSchemaOptions)
		})

		ExpectSchemaPropertiesAreEqual := func(cue, protoc *v1.JSONSchemaProps, property string) {
			cueSchema := cue.Properties[property]
			protocSchema := protoc.Properties[property]

			// Do not compare descriptions
			cueSchema.Description = ""
			protocSchema.Description = ""

			ExpectWithOffset(2, cueSchema).To(Equal(protocSchema))
		}

		ExpectJsonSchemasToMatch := func(cue, protoc *v1.JSONSchemaProps) {
			var (
				fieldName               string
				cueSchema, protocSchema v1.JSONSchemaProps
			)

			// google.protobuf types
			ExpectSchemaPropertiesAreEqual(cue, protoc, "boolValue")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "int32Value")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "uint32Value")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "floatValue")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "duration")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "empty")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "stringValue")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "doubleValue")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "timestamp")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "any")

			// type: google.protobuf.Any
			fieldName = "any"
			cueSchema = cue.Properties[fieldName]
			cueSchema.XPreserveUnknownFields = pointer.BoolPtr(true) // cue doesn't preserve unknown fields for any by default
			protocSchema = protoc.Properties[fieldName]
			ExpectWithOffset(1, cueSchema).To(Equal(protocSchema))

			// type: google.protobuf.Struct
			fieldName = "struct"
			cueSchema = cue.Properties[fieldName]
			cueSchema.XPreserveUnknownFields = pointer.BoolPtr(true) // cue doesn't preserve unknown fields for structs by default
			protocSchema = protoc.Properties[fieldName]
			ExpectWithOffset(1, cueSchema).To(Equal(protocSchema))

			// type: int64
			fieldName = "int64Data"
			cueSchema = cue.Properties[fieldName]
			cueSchema.XIntOrString = true // cue doesn't set x-int-or-string by default
			protocSchema = protoc.Properties[fieldName]
			ExpectWithOffset(1, cueSchema).To(Equal(protocSchema))

			// primitive types
			ExpectSchemaPropertiesAreEqual(cue, protoc, "data")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "mappedData")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "list")
			ExpectSchemaPropertiesAreEqual(cue, protoc, "enumOptions")

			// type: NestedMessage
			fieldName = "nestedMessage"
			cueSchema = cue.Properties[fieldName]
			protocSchema = protoc.Properties[fieldName]
			for _, nestedFieldName := range []string{"optionBool", "optionString"} {
				ExpectWithOffset(1, cueSchema.Properties[nestedFieldName]).To(Equal(protocSchema.Properties[nestedFieldName]))
			}

			// type: repeated NestedMessage
			fieldName = "nestedMessageList"
			cueSchema = cue.Properties[fieldName]
			protocSchema = protoc.Properties[fieldName]
			for _, nestedFieldName := range []string{"optionBool", "optionString"} {
				ExpectWithOffset(1, cueSchema.Items.Schema.Properties[nestedFieldName]).To(Equal(protocSchema.Items.Schema.Properties[nestedFieldName]))
			}
		}

		It("Schema for SimpleMockResource created by cue and protoc match", func() {
			cueSchemas, err := cueGenerator.GetJsonSchemaForProject(project)
			Expect(err).NotTo(HaveOccurred())

			protocSchemas, err := protocGenerator.GetJsonSchemaForProject(project)
			Expect(err).NotTo(HaveOccurred())

			cueSchema := cueSchemas[simpleMockResourceGVK]
			protocSchema := protocSchemas[simpleMockResourceGVK]

			ExpectJsonSchemasToMatch(cueSchema, protocSchema)
		})

		Context("Descriptions for SimpleMockResource can be removed", func() {

			BeforeEach(func() {
				validationSchemaOptions = &schemagen.ValidationSchemaOptions{
					RemoveDescriptionsFromSchema: true,
				}
			})

			It("using protoc", func() {
				protocSchemas, err := protocGenerator.GetJsonSchemaForProject(project)
				Expect(err).NotTo(HaveOccurred())

				protocSchema := protocSchemas[simpleMockResourceGVK]

				fieldNameWithLongComment := "dataWithLongComment"
				fieldWithLongComment := protocSchema.Properties[fieldNameWithLongComment]

				Expect(fieldWithLongComment.Description).To(HaveLen(0))
			})
		})

		Context("Enums for SimpleMockResource can contain only strings", func() {

			BeforeEach(func() {
				validationSchemaOptions = &schemagen.ValidationSchemaOptions{
					EnumAsIntOrString: false,
				}
			})

			It("using protoc", func() {
				protocSchemas, err := protocGenerator.GetJsonSchemaForProject(project)
				Expect(err).NotTo(HaveOccurred())

				protocSchema := protocSchemas[simpleMockResourceGVK]

				enumFieldName := "enumOptions"
				enumField := protocSchema.Properties[enumFieldName]

				Expect(enumField.XIntOrString).To(BeFalse())
			})
		})

		Context("Enums for SimpleMockResource can contain strings or integers", func() {

			BeforeEach(func() {
				validationSchemaOptions = &schemagen.ValidationSchemaOptions{
					EnumAsIntOrString: true,
				}
			})

			It("using protoc", func() {
				protocSchemas, err := protocGenerator.GetJsonSchemaForProject(project)
				Expect(err).NotTo(HaveOccurred())

				protocSchema := protocSchemas[simpleMockResourceGVK]

				enumFieldName := "enumOptions"
				enumField := protocSchema.Properties[enumFieldName]

				Expect(enumField.XIntOrString).To(BeTrue())
			})
		})

		Context("Fields for SimpleMockResource can be configured with empty schemas", func() {

			BeforeEach(func() {
				validationSchemaOptions = &schemagen.ValidationSchemaOptions{
					MessagesWithEmptySchema: []string{
						"testing.solo.io.SimpleMockResource.NestedMessage",
						"core.solo.io.Metadata",
					},
				}
			})

			It("using protoc", func() {
				protocSchemas, err := protocGenerator.GetJsonSchemaForProject(project)
				Expect(err).NotTo(HaveOccurred())

				protocSchema := protocSchemas[simpleMockResourceGVK]

				fieldsWithEmptySchema := []v1.JSONSchemaProps{
					protocSchema.Properties["metadata"],
					protocSchema.Properties["nestedMessage"],
				}
				for _, field := range fieldsWithEmptySchema {
					Expect(field.XPreserveUnknownFields).To(Equal(pointer.BoolPtr(true)))
					Expect(field.Type).To(Equal("object"))
					Expect(field.Properties).To(BeEmpty())
				}

			})
		})

	})

})
