package tests_test

import (
	"path/filepath"

	"k8s.io/utils/pointer"

	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/solo-kit/pkg/utils/modutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

			project *model.Project
		)

		BeforeEach(func() {
			soloKitGoMod, err := modutils.GetCurrentModPackageFile()
			Expect(err).NotTo(HaveOccurred())
			soloKitRoot := filepath.Dir(soloKitGoMod)

			commonImports := []string{
				filepath.Join(soloKitRoot, anyvendor.DefaultDepDir),
			}
			importsCollector := collector.NewCollector([]string{}, commonImports)

			cueGenerator = schemagen.NewCueGenerator(importsCollector, soloKitRoot)
			protocGenerator = schemagen.NewProtocGenerator(importsCollector, soloKitRoot)

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
		})

		ExpectJsonSchemasToMatch := func(cue, protoc *v1beta1.JSONSchemaProps) {
			ExpectWithOffset(1, cue.Properties["data"]).To(Equal(protoc.Properties["data"]))
			ExpectWithOffset(1, cue.Properties["mappedData"]).To(Equal(protoc.Properties["mapped_data"]))
			ExpectWithOffset(1, cue.Properties["list"]).To(Equal(protoc.Properties["list"]))

			cueNestedMessage := cue.Properties["nestedMessage"]
			protocNestedMessage := protoc.Properties["nested_message"]
			ExpectWithOffset(1, cueNestedMessage.Properties["optionBool"]).To(Equal(protocNestedMessage.Properties["option_bool"]))
			ExpectWithOffset(1, cueNestedMessage.Properties["optionString"]).To(Equal(protocNestedMessage.Properties["option_string"]))

			cueNestedMessageList := cue.Properties["nestedMessageList"]
			protocNestedMessageList := protoc.Properties["nested_message_list"]
			ExpectWithOffset(1, cueNestedMessageList.Items.Schema.Properties["optionBool"]).To(Equal(protocNestedMessageList.Items.Schema.Properties["option_bool"]))
			ExpectWithOffset(1, cueNestedMessageList.Items.Schema.Properties["optionString"]).To(Equal(protocNestedMessageList.Items.Schema.Properties["option_string"]))

			// cue doesn't preserve unknown fields for structs by default
			cueStruct := cue.Properties["struct"]
			cueStruct.XPreserveUnknownFields = pointer.BoolPtr(true)
			ExpectWithOffset(1, cueStruct).To(Equal(protoc.Properties["struct"]))

			ExpectWithOffset(1, cue.Properties["any"]).To(Equal(protoc.Properties["any"]))
		}

		It("Schema for SimpleMockResource created by cue and protoc match", func() {
			cueSchemas, err := cueGenerator.GetJsonSchemaForProject(project)
			Expect(err).NotTo(HaveOccurred())

			protocSchemas, err := protocGenerator.GetJsonSchemaForProject(project)
			Expect(err).NotTo(HaveOccurred())

			simpleMockResourceGVK := schema.GroupVersionKind{
				Group:   "testing.solo.io.v1",
				Version: "v1",
				Kind:    "SimpleMockResource",
			}

			cueSchema := cueSchemas[simpleMockResourceGVK]
			protocSchema := protocSchemas[simpleMockResourceGVK]

			ExpectJsonSchemasToMatch(cueSchema, protocSchema)
		})

	})

})
