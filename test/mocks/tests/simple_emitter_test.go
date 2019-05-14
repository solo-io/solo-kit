package tests_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/util"
)

var _ = Describe("SimpleEmitter", func() {
	var clusterMocks1, clusterMocks2, clusterFakes3, clusterFakes4 *wrapper.Client // add / remove later
	var watcher wrapper.WatchAggregator
	var emitter v1.TestingSimpleEmitter
	clusterName1 := "clustr1"
	clusterName2 := "clustr2"
	clusterName3 := "clustr3"
	clusterName4 := "clustr4"
	BeforeEach(func() {
		baseMocks1 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		baseMocks2 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		baseFakes3 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.FakeResource{})
		baseFakes4 := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.FakeResource{})
		clusterMocks1 = wrapper.NewClusterClient(baseMocks1, clusterName1)
		clusterMocks2 = wrapper.NewClusterClient(baseMocks2, clusterName2)
		clusterFakes3 = wrapper.NewClusterClient(baseFakes3, clusterName3)
		clusterFakes4 = wrapper.NewClusterClient(baseFakes4, clusterName4)

		watcher = wrapper.NewWatchAggregator()
		err := watcher.AddWatch(clusterMocks1)
		Expect(err).NotTo(HaveOccurred())
		err = watcher.AddWatch(clusterMocks2)
		Expect(err).NotTo(HaveOccurred())

		watch := wrapper.AggregatedWatchFromClients(wrapper.ClientWatchOpts{
			BaseClient: watcher,
			Namespace:  "a",
		}, wrapper.ClientWatchOpts{
			BaseClient: watcher,
			Namespace:  "b",
		}, wrapper.ClientWatchOpts{
			BaseClient:   watcher,
			Namespace:    "c",
			ResourceName: "just-me",
		})

		emitter = v1.NewTestingSimpleEmitter(watch)
	})
	It("sends snapshots from a single aggregated watch", func() {
		snapshots, errs, err := emitter.Snapshots(context.TODO())
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			_, err := clusterMocks1.Write(v1.NewMockResource("a", "a"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = clusterMocks1.Write(v1.NewMockResource("a", "b"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = clusterMocks2.Write(v1.NewMockResource("a", "a"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = clusterMocks2.Write(v1.NewMockResource("a", "b"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = clusterMocks2.Write(v1.NewMockResource("c", "not-me"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = clusterMocks2.Write(v1.NewMockResource("c", "just-me"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = watcher.AddWatch(clusterFakes3)
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterFakes3.Write(v1.NewFakeResource("a", "a"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = clusterFakes3.Write(v1.NewFakeResource("a", "b"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

		}()

		var snap *v1.TestingSnapshot
		Eventually(func() *v1.TestingSnapshot {
			select {
			default:
			case err := <-errs:
				Expect(err).NotTo(HaveOccurred())
			case snap = <-snapshots:
				snap.Mocks.EachResource(util.ZeroResourceVersion)
				snap.Fakes.EachResource(util.ZeroResourceVersion)
			case <-time.After(time.Millisecond * 5):
				Fail("expected a snpashot in channel")
			}
			return snap
		}, time.Second*2).Should(Equal(&v1.TestingSnapshot{
			Mocks: v1.MockResourceList{
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "a",
						Namespace: "a",
						Cluster:   "clustr1",
					},
				},
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "b",
						Namespace: "a",
						Cluster:   "clustr1",
					},
				},
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "a",
						Namespace: "a",
						Cluster:   "clustr2",
					},
				},
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "b",
						Namespace: "a",
						Cluster:   "clustr2",
					},
				},
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "just-me",
						Namespace: "c",
						Cluster:   "clustr2",
					},
				},
			},
			Fakes: v1.FakeResourceList{
				&v1.FakeResource{
					Count: 0x00000000,
					Metadata: core.Metadata{
						Name:      "a",
						Namespace: "a",
						Cluster:   "clustr3",
					},
				},
				&v1.FakeResource{
					Count: 0x00000000,
					Metadata: core.Metadata{
						Name:      "b",
						Namespace: "a",
						Cluster:   "clustr3",
					},
				},
			},
		}))

		go func() {
			watcher.RemoveWatch(clusterFakes3)
			err = watcher.AddWatch(clusterFakes4)
			Expect(err).NotTo(HaveOccurred())

			// update
			_, err = clusterFakes4.Write(v1.NewFakeResource("a", "b"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		}()

		snap = nil
		Eventually(func() *v1.TestingSnapshot {
			select {
			default:
			case err := <-errs:
				Expect(err).NotTo(HaveOccurred())
			case snap = <-snapshots:
				snap.Mocks.EachResource(util.ZeroResourceVersion)
				snap.Fakes.EachResource(util.ZeroResourceVersion)
			case <-time.After(time.Millisecond * 5):
				Fail("expected a snpashot in channel")
			}
			return snap
		}, time.Second*4).Should(Equal(&v1.TestingSnapshot{
			Mocks: v1.MockResourceList{
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "a",
						Namespace: "a",
						Cluster:   "clustr1",
					},
				},
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "b",
						Namespace: "a",
						Cluster:   "clustr1",
					},
				},
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "a",
						Namespace: "a",
						Cluster:   "clustr2",
					},
				},
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "b",
						Namespace: "a",
						Cluster:   "clustr2",
					},
				},
				&v1.MockResource{
					Metadata: core.Metadata{
						Name:      "just-me",
						Namespace: "c",
						Cluster:   "clustr2",
					},
				},
			},
			Fakes: v1.FakeResourceList{
				&v1.FakeResource{
					Count: 0x00000000,
					Metadata: core.Metadata{
						Name:      "b",
						Namespace: "a",
						Cluster:   "clustr4",
					},
				},
			},
		}))

	})
})
