package wrapper_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/util"

	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
)

var _ = Describe("ResourceWatch", func() {
	It("aggregates watch funcs", func() {
		base1 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		base2 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})

		watch1 := ResourceWatch(base1, "a", nil)
		watch2 := ResourceWatch(base1, "b", nil)
		watch3 := ResourceWatch(base2, "d", nil)

		multiWatch := AggregatedWatch(watch1, watch2, watch3)
		lists, errs, err := multiWatch(context.TODO())
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			_, err := base1.Write(v1.NewMockResource("a", "a"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = base1.Write(v1.NewMockResource("b", "b"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = base2.Write(v1.NewMockResource("c", "c"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = base2.Write(v1.NewMockResource("d", "d"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		}()

		var list resources.ResourceList

		Eventually(func() resources.ResourceList {
			select {
			case err := <-errs:
				Expect(err).NotTo(HaveOccurred())
			case list = <-lists:
			default:
			}
			return list
		}, time.Second*1000).Should(HaveLen(3))

		list.Each(util.ZeroResourceVersion)

		Expect(list).To(Equal(resources.ResourceList{
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "a"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "b", Name: "b"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "d", Name: "d"}},
		}))
	})
})
