package tests

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	// . "github.com/onsi/gomega"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("hashing", func() {
	var allResources v1.MockResourceList
	BeforeEach(func() {
		allResources = nil
		for i := 0; i < 10000; i++ {
			titleInt := strconv.Itoa(i)
			allResources = append(allResources, &v1.MockResource{
				Metadata: core.Metadata{
					Name:            titleInt,
					Namespace:       titleInt,
					Cluster:         titleInt,
					ResourceVersion: titleInt,
					Generation:      int64(i),
					OwnerReferences: nil,
				},
				Data:          titleInt,
				SomeDumbField: "",
				TestOneofFields: &v1.MockResource_OneofOne{
					OneofOne: "hello",
				},
			})
		}
	})
	Context("benchmark", func() {
		Measure("it should do something hard efficiently", func(b Benchmarker) {
			const times = 1
			generatedHash := b.Time(fmt.Sprintf("runtime of %d generated hash calls", times), func() {
				for i := 0; i < times; i++ {
					for _, us := range allResources {
						us.Hash(nil)
					}
				}
			})
			reflectionHash := b.Time(fmt.Sprintf("runtime of %d reflection based hash calls", times), func() {
				for i := 0; i < times; i++ {
					for _, us := range allResources {
						hashstructure.Hash(us, nil)
					}
				}
			})

			// divide by 1e3 to get time in micro seconds instead of nano seconds
			b.RecordValue("Runtime per generated call in µ seconds", float64(int64(generatedHash)/times)/1e3)
			b.RecordValue("Runtime per reflection call in µ seconds", float64(int64(reflectionHash)/times)/1e3)

		}, 10)
	})
	Context("accuracy", func() {
		It("Exhaustive", func() {
			present := make(map[uint64]struct{}, len(allResources))
			for _, v := range allResources {
				hash, err := v.Hash(nil)
				Expect(err).NotTo(HaveOccurred())
				_, ok := present[hash]
				Expect(ok).To(BeFalse())
				present[hash] = struct{}{}
			}
		})
	})
})
