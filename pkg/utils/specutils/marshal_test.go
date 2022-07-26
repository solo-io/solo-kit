package specutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	"github.com/solo-io/solo-kit/pkg/utils/specutils"
	mocksv1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("Marshal", func() {

	Context("UnmarshalSpecMapToResource", func() {

		It("can unmarshal map of proto resource", func() {
			originalResource := &mocksv1.MockResource{
				Data:          "data",
				SomeDumbField: "some dumb field",
			}

			originalResourceMap, err := protoutils.MarshalMap(originalResource)
			Expect(err).NotTo(HaveOccurred())

			var newResource mocksv1.MockResource
			err = specutils.UnmarshalSpecMapToProto(originalResourceMap, &newResource)
			Expect(err).NotTo(HaveOccurred())
			testutils.ExpectEqualProtoMessages(originalResource, &newResource)
		})

		It("can unmarshal map of raw fields", func() {
			originalResourceMap := map[string]interface{}{
				"data":          "data",
				"someDumbField": "some dumb field",
			}

			var newResource mocksv1.MockResource
			err := specutils.UnmarshalSpecMapToProto(originalResourceMap, &newResource)
			Expect(err).NotTo(HaveOccurred())
			Expect(newResource.Data).To(Equal("data"))
			Expect(newResource.SomeDumbField).To(Equal("some dumb field"))
		})

		It("can unmarshal map of raw fields with unknown field", func() {
			originalResourceMap := map[string]interface{}{
				"data":          "data",
				"someDumbField": "some dumb field",
				"invalidField":  "intentionally included field that is not on MockResource",
			}

			By("do not error on unknown fields when using UnmarshalSpecMapToProto")
			var newResource mocksv1.MockResource
			err := specutils.UnmarshalSpecMapToProto(originalResourceMap, &newResource)
			Expect(err).NotTo(HaveOccurred())
			Expect(newResource.Data).To(Equal("data"))
			Expect(newResource.SomeDumbField).To(Equal("some dumb field"))

			By("error on unknown fields when using UnmarshalMapToProto")
			err = protoutils.UnmarshalMapToProto(originalResourceMap, &newResource)
			Expect(err).To(HaveOccurred())
		})

	})

})
