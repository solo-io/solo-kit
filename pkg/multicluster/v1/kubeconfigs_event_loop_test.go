// Code generated by solo-kit. DO NOT EDIT.

// +build solokit

package v1

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
)

var _ = Describe("KubeconfigsEventLoop", func() {
	var (
		namespace string
		emitter   KubeconfigsEmitter
		err       error
	)

	BeforeEach(func() {

		kubeConfigClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		kubeConfigClient, err := NewKubeConfigClient(kubeConfigClientFactory)
		Expect(err).NotTo(HaveOccurred())

		emitter = NewKubeconfigsEmitter(kubeConfigClient)
	})
	It("runs sync function on a new snapshot", func() {
		_, err = emitter.KubeConfig().Write(NewKubeConfig(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		sync := &mockKubeconfigsSyncer{}
		el := NewKubeconfigsEventLoop(emitter, sync)
		_, err := el.Run([]string{namespace}, clients.WatchOpts{})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sync.Synced, 5*time.Second).Should(BeTrue())
	})
})

type mockKubeconfigsSyncer struct {
	synced bool
	mutex  sync.Mutex
}

func (s *mockKubeconfigsSyncer) Synced() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.synced
}

func (s *mockKubeconfigsSyncer) Sync(ctx context.Context, snap *KubeconfigsSnapshot) error {
	s.mutex.Lock()
	s.synced = true
	s.mutex.Unlock()
	return nil
}
