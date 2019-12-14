package tests

import (
	"fmt"

	"github.com/mitchellh/hashstructure"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	// . "github.com/onsi/gomega"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = FDescribe("SnapshotBenchmark", func() {
	var allResources v1.MockResourceList
	BeforeEach(func() {
		for i := 0; i < 10000; i++ {
			allResources = append(allResources, &v1.MockResource{
				Status: core.Status{
					State:      0,
					Reason:     "",
					ReportedBy: "",
				},
				Metadata: core.Metadata{
					Name:                 "",
					Namespace:            "",
					Cluster:              "",
					ResourceVersion:      "",
					Labels:               nil,
					Annotations:          nil,
					Generation:           0,
					OwnerReferences:      nil,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
				Data:          "",
				SomeDumbField: "",
				TestOneofFields: &v1.MockResource_OneofOne{
					OneofOne: "hello",
				},
				XXX_NoUnkeyedLiteral: struct{}{},
				XXX_unrecognized:     nil,
				XXX_sizecache:        0,
			})
		}
	})
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
		b.RecordValue("Runtime per reflection call in µ seconds", float64(int64(generatedHash)/times)/1e3)
		b.RecordValue("Runtime per generated call in µ seconds", float64(int64(reflectionHash)/times)/1e3)

	}, 10)
})
