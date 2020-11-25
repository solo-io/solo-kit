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
	// This is now implemented on the snapshot level, rather than the resource, so in order to test this
	// the whole snapshot must be hashed
	Context("skip_hashing_annotations=true", func() {
		It("ignores the resource meta annotations in the hash", func() {
			snap1 := &v2alpha1.TestingSnapshot{Fcars: []*v2alpha1.FrequentlyChangingAnnotationsResource{
				v2alpha1.NewFrequentlyChangingAnnotationsResource("a", "b"),
			}}
			snap2 := &v2alpha1.TestingSnapshot{Fcars: []*v2alpha1.FrequentlyChangingAnnotationsResource{
				v2alpha1.NewFrequentlyChangingAnnotationsResource("a", "b"),
			}}

			snap1Hash, _ := snap1.Hash(nil)
			snap2Hash, _ := snap2.Hash(nil)
			// sanity check
			Expect(snap1Hash).To(Equal(snap2Hash))

			annotations := map[string]string{"ignore": "me"}
			snap2.Fcars[0].Metadata.Annotations = annotations
			snap2Hash, _ = snap2.Hash(nil)
			Expect(snap1Hash).To(Equal(snap2Hash))
			// check that metadata of original was not changed
			Expect(snap2.Fcars[0].Metadata.Annotations).To(Equal(annotations))
		})
	})
})

func snapWithFields(hashSensitiveField, hashInsensitiveField string) *v1.TestingSnapshot {
	return &v1.TestingSnapshot{
		Mocks: v1.MockResourceList{
			{
				Metadata:      &core.Metadata{Name: hashSensitiveField, Namespace: hashSensitiveField},
				Data:          hashSensitiveField,
				SomeDumbField: hashInsensitiveField,
				TestOneofFields: &v1.MockResource_OneofOne{
					OneofOne: hashSensitiveField,
				},
			},
			{
				Metadata:      &core.Metadata{Name: hashSensitiveField + "2", Namespace: hashSensitiveField},
				Data:          hashSensitiveField,
				SomeDumbField: hashInsensitiveField,
				TestOneofFields: &v1.MockResource_OneofTwo{
					OneofTwo: true,
				},
			},
		},
	}
}
