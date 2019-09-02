package wrapper_test

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/util"
)

var _ = Describe("watchAggregator", func() {
	var cluster1, cluster2, cluster3, cluster4 *Client // add / remove later
	var watcher WatchAggregator
	clusterName1 := "clustr1"
	clusterName2 := "clustr2"
	clusterName3 := "clustr3"
	clusterName4 := "clustr4"
	BeforeEach(func() {
		base1 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		base2 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		base3 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		base4 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		cluster1 = NewClusterClient(base1, clusterName1)
		cluster2 = NewClusterClient(base2, clusterName2)
		cluster3 = NewClusterClient(base3, clusterName3)
		cluster4 = NewClusterClient(base4, clusterName4)

		watcher = NewWatchAggregator()
		err := watcher.AddWatch(cluster1)
		Expect(err).NotTo(HaveOccurred())
		err = watcher.AddWatch(cluster2)
		Expect(err).NotTo(HaveOccurred())
	})
	It("aggregates watches from multiple clients", func() {
		watch, errs, err := watcher.Watch("", clients.WatchOpts{RefreshRate: time.Millisecond})
		Expect(err).NotTo(HaveOccurred())

		l := sync.Mutex{}

		go func() {
			defer GinkgoRecover()
			_, err = cluster1.Write(v1.NewMockResource("a", "a"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = cluster1.Write(v1.NewMockResource("a", "b"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = cluster2.Write(v1.NewMockResource("a", "a"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = cluster2.Write(v1.NewMockResource("a", "b"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = watcher.AddWatch(cluster3)
			Expect(err).NotTo(HaveOccurred())

			l.Lock()
			_, err = cluster3.Write(v1.NewMockResource("a", "a"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = cluster3.Write(v1.NewMockResource("a", "b"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			l.Unlock()

		}()

		var list resources.ResourceList

		Eventually(func() resources.ResourceList {
			select {
			default:
			case err := <-errs:
				Expect(err).NotTo(HaveOccurred())
			case list = <-watch:
				return list
			case <-time.After(time.Millisecond * 5):
				Fail("expected a message in channel")
			}
			return nil
		}, time.Second*10).Should(HaveLen(6))

		list.Each(util.ZeroResourceVersion)

		Expect(list).To(Equal(resources.ResourceList{
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "a", Cluster: "clustr1"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "b", Cluster: "clustr1"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "a", Cluster: "clustr2"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "b", Cluster: "clustr2"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "a", Cluster: "clustr3"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "b", Cluster: "clustr3"}},
		}))

		go func() {
			l.Lock()
			watcher.RemoveWatch(cluster3)
			err = watcher.AddWatch(cluster4)
			l.Unlock()
			Expect(err).NotTo(HaveOccurred())

			// update
			_, err = cluster4.Write(v1.NewMockResource("a", "b"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		}()

		Eventually(func() resources.ResourceList {
			select {
			default:
			case err := <-errs:
				Expect(err).NotTo(HaveOccurred())
			case list = <-watch:
				return list
			case <-time.After(time.Millisecond * 5):
				Fail("expected a message in channel")
			}
			return nil
		}, time.Second*10).Should(HaveLen(5))

		list.Each(util.ZeroResourceVersion)

		Expect(list).To(Equal(resources.ResourceList{
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "a", Cluster: "clustr1"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "b", Cluster: "clustr1"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "a", Cluster: "clustr2"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "b", Cluster: "clustr2"}},
			&v1.MockResource{Metadata: core.Metadata{Namespace: "a", Name: "b", Cluster: "clustr4"}},
		}))

	})
})
