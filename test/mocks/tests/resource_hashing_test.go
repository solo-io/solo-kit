package tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("Resource Hashing", func() {
	It("ignores fields marked with skip_hashing=true", func() {
		hashSensitiveField := "hash sensitive"
		hashInsensitiveField := "hash insensitive"
		originalSnap := snapWithFields(hashSensitiveField, hashInsensitiveField)
		snapWithInsensitiveChanged := snapWithFields(hashSensitiveField, hashInsensitiveField+" changed")
		snapWithSensitiveChanged := snapWithFields(hashSensitiveField+" changed", hashInsensitiveField)
		Expect(originalSnap.Hash()).To(Equal(snapWithInsensitiveChanged.Hash()))
		Expect(originalSnap.Hash()).NotTo(Equal(snapWithSensitiveChanged.Hash()))
	})
})

func snapWithFields(hashSensitiveField, hashInsensitiveField string) *v1.TestingSnapshot {
	return &v1.TestingSnapshot{
		Mocks: map[string]v1.MockResourceList{
			"": {
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
		},
	}
}
