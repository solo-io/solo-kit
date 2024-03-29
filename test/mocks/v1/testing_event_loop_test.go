//go:build solokit
// +build solokit

package v1

import (
	"context"
	"sync"
	"time"

	github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
)

var _ = Describe("TestingEventLoop", func() {
	var (
		ctx       context.Context
		namespace string
		emitter   TestingEmitter
		err       error
	)

	BeforeEach(func() {
		ctx = context.Background()

		simpleMockResourceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		simpleMockResourceClient, err := NewSimpleMockResourceClient(ctx, simpleMockResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())

		mockResourceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		mockResourceClient, err := NewMockResourceClient(ctx, mockResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())

		fakeResourceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		fakeResourceClient, err := NewFakeResourceClient(ctx, fakeResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())

		anotherMockResourceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		anotherMockResourceClient, err := NewAnotherMockResourceClient(ctx, anotherMockResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())

		clusterResourceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		clusterResourceClient, err := NewClusterResourceClient(ctx, clusterResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())

		mockCustomTypeClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		mockCustomTypeClient, err := NewMockCustomTypeClient(ctx, mockCustomTypeClientFactory)
		Expect(err).NotTo(HaveOccurred())

		mockCustomSpecHashTypeClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		mockCustomSpecHashTypeClient, err := NewMockCustomSpecHashTypeClient(ctx, mockCustomSpecHashTypeClientFactory)
		Expect(err).NotTo(HaveOccurred())

		podClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		podClient, err := github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPodClient(ctx, podClientFactory)
		Expect(err).NotTo(HaveOccurred())

		emitter = NewTestingEmitter(simpleMockResourceClient, mockResourceClient, fakeResourceClient, anotherMockResourceClient, clusterResourceClient, mockCustomTypeClient, mockCustomSpecHashTypeClient, podClient)
	})
	It("runs sync function on a new snapshot", func() {
		_, err = emitter.SimpleMockResource().Write(NewSimpleMockResource(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.MockResource().Write(NewMockResource(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.FakeResource().Write(NewFakeResource(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.AnotherMockResource().Write(NewAnotherMockResource(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.ClusterResource().Write(NewClusterResource(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.MockCustomType().Write(NewMockCustomType(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.Pod().Write(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPod(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		sync := &mockTestingSyncer{}
		el := NewTestingEventLoop(emitter, sync)
		_, err := el.Run([]string{namespace}, clients.WatchOpts{})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sync.Synced, 5*time.Second).Should(BeTrue())
	})
})

type mockTestingSyncer struct {
	synced bool
	mutex  sync.Mutex
}

func (s *mockTestingSyncer) Synced() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.synced
}

func (s *mockTestingSyncer) Sync(ctx context.Context, snap *TestingSnapshot) error {
	s.mutex.Lock()
	s.synced = true
	s.mutex.Unlock()
	return nil
}
