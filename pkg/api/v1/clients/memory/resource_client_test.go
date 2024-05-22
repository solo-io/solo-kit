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

	Context("listing resources", func() {
		var (
			obj1 *v1.MockResource
			obj2 *v1.MockResource
			obj3 *v1.MockResource
			obj4 *v1.MockResource
			obj5 *v1.MockResource
		)

		BeforeEach(func() {
			obj1 = &v1.MockResource{
				Metadata: &core.Metadata{
					Name:      "name1",
					Namespace: "ns1",
					Labels: map[string]string{
						"key": "val1",
					},
				},
			}
			obj2 = &v1.MockResource{
				Metadata: &core.Metadata{
					Name:      "name2",
					Namespace: "ns2",
					Labels: map[string]string{
						"key": "val2",
					},
				},
			}
			obj3 = &v1.MockResource{
				Metadata: &core.Metadata{
					Name:      "name3",
					Namespace: "ns1",
					Labels: map[string]string{
						"key": "val2",
					},
				},
			}
			obj4 = &v1.MockResource{
				Metadata: &core.Metadata{
					Name:      "name4",
					Namespace: "ns2",
					Labels: map[string]string{
						"key": "val3",
					},
				},
			}
			obj5 = &v1.MockResource{
				Metadata: &core.Metadata{
					Name:      "name5",
					Namespace: "ns1",
					Labels: map[string]string{
						"key": "val3",
					},
				},
			}

			_, err := client.Write(obj1, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = client.Write(obj2, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = client.Write(obj3, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = client.Write(obj4, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = client.Write(obj5, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("lists all resources when empty namespace is provided", func() {
			resources, err := client.List("", clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())

			// resources are sorted by namespace, then name
			expectedResourceNames := []string{
				"name1", "name3", "name5", // ns1
				"name2", "name4", // ns2
			}
			Expect(resources).To(HaveLen(len(expectedResourceNames)))
			for i, r := range resources {
				Expect(r.GetMetadata().GetName()).To(Equal(expectedResourceNames[i]))
			}
		})

		It("lists resources in a given namespace", func() {
			resources, err := client.List("ns2", clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())

			expectedResourceNames := []string{
				"name2", "name4",
			}
			Expect(resources).To(HaveLen(len(expectedResourceNames)))
			for i, r := range resources {
				Expect(r.GetMetadata().GetName()).To(Equal(expectedResourceNames[i]))
			}
		})

		It("returns empty list if namespace is invalid", func() {
			resources, err := client.List("invalid-namespace", clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(resources).To(HaveLen(0))
		})

		It("returns resources matching the given selector, across all namespaces", func() {
			resources, err := client.List("", clients.ListOpts{
				Selector: map[string]string{
					"key": "val2",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			// resources are sorted by namespace, then name
			expectedResourceNames := []string{
				"name3", "name2",
			}
			Expect(resources).To(HaveLen(len(expectedResourceNames)))
			for i, r := range resources {
				Expect(r.GetMetadata().GetName()).To(Equal(expectedResourceNames[i]))
			}
		})

		It("returns resources matching the given selector, in given namespace", func() {
			resources, err := client.List("ns2", clients.ListOpts{
				Selector: map[string]string{
					"key": "val2",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(resources).To(HaveLen(1))
			Expect(resources[0].GetMetadata().GetName()).To(Equal("name2"))
		})

		It("returns resources matching the given expression selector, across all namespaces", func() {
			resources, err := client.List("", clients.ListOpts{
				ExpressionSelector: "key in (val1,val3)",
			})
			Expect(err).NotTo(HaveOccurred())

			// resources are sorted by namespace, then name
			expectedResourceNames := []string{
				"name1", "name5", "name4",
			}
			Expect(resources).To(HaveLen(len(expectedResourceNames)))
			for i, r := range resources {
				Expect(r.GetMetadata().GetName()).To(Equal(expectedResourceNames[i]))
			}
		})

		It("returns resources matching the given expression selector, in given namespace", func() {
			resources, err := client.List("ns2", clients.ListOpts{
				ExpressionSelector: "key in (val1,val3)",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(resources).To(HaveLen(1))
			Expect(resources[0].GetMetadata().GetName()).To(Equal("name4"))
		})

		It("when both selector and expression selector are provided, uses expression selector", func() {
			resources, err := client.List("ns2", clients.ListOpts{
				Selector: map[string]string{
					"key": "val2",
				},
				ExpressionSelector: "key in (val1,val3)",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(resources).To(HaveLen(1))
			Expect(resources[0].GetMetadata().GetName()).To(Equal("name4"))
		})
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
