package multicluster_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	v12 "github.com/solo-io/solo-kit/api/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	. "github.com/solo-io/solo-kit/pkg/multicluster"
)

var _ = Describe("ConfigHandler", func() {
	It("splits restconfigs by cluster and adds/removes them from the handler", func() {
		h1, h2 := newMockHandler(), newMockHandler()
		watcher := &fakeKCWatcher{}

		clusters := NewRestConfigHandler(watcher, h1, h2)

		errs, err := clusters.Run(context.TODO(), nil, nil, nil)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			Expect(<-errs).NotTo(HaveOccurred())
		}()

		<-watcher.done
		h1.l.Lock()
		defer h1.l.Unlock()
		Expect(h1.added).To(HaveKey("cluster-0"))
		Expect(h1.added).To(HaveKey("cluster-1"))
		Expect(h1.removed).To(HaveKey("cluster-1"))

		h2.l.Lock()
		defer h2.l.Unlock()
		Expect(h2.added).To(HaveKey("cluster-0"))
		Expect(h2.added).To(HaveKey("cluster-1"))
		Expect(h2.removed).To(HaveKey("cluster-1"))
	})
})

type mockHandler struct {
	l       sync.Mutex
	added   map[string]string
	removed map[string]string
}

func newMockHandler() *mockHandler {
	return &mockHandler{
		added:   map[string]string{},
		removed: map[string]string{},
	}
}

func (h *mockHandler) ClusterAdded(cluster string, restConfig *rest.Config) {
	h.l.Lock()
	h.added[cluster] = restConfig.Host
	h.l.Unlock()
}

func (h *mockHandler) ClusterRemoved(cluster string, restConfig *rest.Config) {
	h.l.Lock()
	h.removed[cluster] = restConfig.Host
	h.l.Unlock()
}

type fakeKCWatcher struct {
	done chan struct{}
}

func (w *fakeKCWatcher) WatchKubeConfigs(ctx context.Context, kube kubernetes.Interface, cache cache.KubeCoreCache) (<-chan v1.KubeConfigList, <-chan error, error) {
	w.done = make(chan struct{})
	out, errs := make(chan v1.KubeConfigList), make(chan error)
	go func() {
		defer close(w.done)
		for i := 0; i < 5; i++ {
			out <- makeKubeConfigs("cluster", "", "", i%2+1) // create jitter
			time.Sleep(time.Millisecond * 5)
		}
	}()

	return out, errs, nil
}

func makeKubeConfigs(cluster, namespace, name string, length int) v1.KubeConfigList {
	kubeConfig, err := kubeutils.GetKubeConfig("", "")
	Expect(err).NotTo(HaveOccurred())

	var kcs v1.KubeConfigList
	for i := 0; i < length; i++ {
		kubeCfgSecret, err := secretconverter.KubeConfigToSecret(&v1.KubeConfig{
			KubeConfig: v12.KubeConfig{
				Metadata: core.Metadata{Namespace: namespace, Name: name},
				Config:   *kubeConfig,
				Cluster:  fmt.Sprintf("%v-%v", cluster, i),
			},
		})
		Expect(err).NotTo(HaveOccurred())
		kc, err := secretconverter.KubeCfgFromSecret(kubeCfgSecret)
		Expect(err).NotTo(HaveOccurred())

		kcs = append(kcs, kc)
	}

	return kcs
}
