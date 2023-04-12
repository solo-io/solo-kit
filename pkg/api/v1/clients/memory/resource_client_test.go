package memory_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/tests/generic"
)

var _ = Describe("Base", func() {
	var (
		client *ResourceClient
	)
	BeforeEach(func() {
		client = NewResourceClient(NewInMemoryResourceCache(), &v1.MockResource{})
	})
	AfterEach(func() {
	})
	It("CRUDs resources", func() {
		selector := map[string]string{
			helpers.TestLabel: helpers.RandString(8),
		}
		generic.TestCrudClient("ns1", "ns2", client, clients.WatchOpts{
			Selector:    selector,
			Ctx:         context.TODO(),
			RefreshRate: time.Minute,
		})
	})
	It("should not return pointer to internal object", func() {
		obj := &v1.MockResource{
			Metadata: &core.Metadata{
				Namespace: "ns",
				Name:      "n",
			},
			Data: "test",
		}
		client.Write(obj, clients.WriteOpts{})
		ret, err := client.Read("ns", "n", clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(ret).NotTo(BeIdenticalTo(obj))

		ret2, err := client.Read("ns", "n", clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(ret).NotTo(BeIdenticalTo(ret2))

		listret, err := client.List("ns", clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(listret[0]).NotTo(BeIdenticalTo(obj))

		listret2, err := client.List("ns", clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(listret[0]).NotTo(BeIdenticalTo(listret2[0]))
	})

	Context("Benchmarks", func() {
		Measure("it should perform list efficiently", func(b Benchmarker) {
			const numobjs = 10000

			for i := 0; i < numobjs; i++ {
				obj := &v1.MockResource{
					Metadata: &core.Metadata{
						Namespace: "ns",
						Name:      fmt.Sprintf("n-%v", numobjs-i),
					},
					Data: strings.Repeat("123", 1000) + fmt.Sprintf("test-%v", i),
				}
				client.Write(obj, clients.WriteOpts{})
			}
			l := clients.ListOpts{}
			var output resources.ResourceList
			var err error
			runtime := b.Time("runtime", func() {
				output, err = client.List("ns", l)
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(HaveLen(numobjs))
			Expect(output[0].GetMetadata().Name).To(Equal("n-1"))

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.5), "List() shouldn't take too long.")
		}, 10)

	})

})
