package tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
)

var _ = Describe("Resource Hashing", func() {
	It("ignores fields marked with skip_hashing=true", func() {
		hashSensitiveField := "hash sensitive"
		hashInsensitiveField := "hash insensitive"
		originalSnap := snapWithFields(hashSensitiveField, hashInsensitiveField)
		snapWithInsensitiveChanged := snapWithFields(hashSensitiveField, hashInsensitiveField+" changed")
		snapWithSensitiveChanged := snapWithFields(hashSensitiveField+" changed", hashInsensitiveField)
		originalSnapHash, err := originalSnap.Hash(nil)
		Expect(err).NotTo(HaveOccurred())
		snapWithInsensitiveChangedHash, err := snapWithInsensitiveChanged.Hash(nil)
		Expect(err).NotTo(HaveOccurred())
		snapWithSensitiveChangedHash, err := snapWithSensitiveChanged.Hash(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(originalSnapHash).To(Equal(snapWithInsensitiveChangedHash))
		Expect(originalSnapHash).NotTo(Equal(snapWithSensitiveChangedHash))
	})
	Context("skip_hashing_annotations=true", func() {
		It("ignores the resource meta annotations in the hash", func() {
			res1 := v2alpha1.NewFrequentlyChangingAnnotationsResource("a", "b")
			res2 := v2alpha1.NewFrequentlyChangingAnnotationsResource("a", "b")

			res1Hash, _ := res1.Hash(nil)
			res2Hash, _ := res2.Hash(nil)
			// sanity check
			Expect(res1Hash).To(Equal(res2Hash))

			res1.Metadata.Annotations = map[string]string{"ignore": "me"}
			res2Hash, _ = res2.Hash(nil)
			Expect(res1Hash).To(Equal(res2Hash))
		})
	})
})

func snapWithFields(hashSensitiveField, hashInsensitiveField string) *v1.TestingSnapshot {
	return &v1.TestingSnapshot{
		Mocks: v1.MockResourceList{
			{
				Metadata:      core.Metadata{Name: hashSensitiveField, Namespace: hashSensitiveField},
				Data:          hashSensitiveField,
				SomeDumbField: hashInsensitiveField,
				TestOneofFields: &v1.MockResource_OneofOne{
					OneofOne: hashSensitiveField,
				},
			},
			{
				Metadata:      core.Metadata{Name: hashSensitiveField + "2", Namespace: hashSensitiveField},
				Data:          hashSensitiveField,
				SomeDumbField: hashInsensitiveField,
				TestOneofFields: &v1.MockResource_OneofTwo{
					OneofTwo: true,
				},
			},
		},
	}
}
