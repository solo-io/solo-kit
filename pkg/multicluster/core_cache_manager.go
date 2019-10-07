package multicluster

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type coreCacheWrapper struct {
	cancel    context.CancelFunc
	coreCache cache.KubeCoreCache
}

type KubeCoreCacheGetter interface {
	GetCache(cluster string) cache.KubeCoreCache
}

type KubeCoreCacheManager interface {
	ClusterHandler
	KubeCoreCacheGetter
}

type manager struct {
	ctx         context.Context
	caches      map[string]coreCacheWrapper
	cacheAccess sync.RWMutex
}

var _ KubeCoreCacheManager = &manager{}

func NewKubeCoreCacheManager(ctx context.Context) *manager {
	return &manager{
		ctx:         ctx,
		caches:      make(map[string]coreCacheWrapper),
		cacheAccess: sync.RWMutex{},
	}
}

// TODO should this just be a noop since GetCache can handle provisioning new caches?
func (m *manager) ClusterAdded(cluster string, restConfig *rest.Config) {
	ctx, cancel := context.WithCancel(m.ctx)
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return
	}
	coreCache, err := cache.NewKubeCoreCache(ctx, kubeClient)
	if err != nil {
		return
	}

	cw := coreCacheWrapper{
		cancel:    cancel,
		coreCache: coreCache,
	}
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	m.caches[cluster] = cw
}

func (m *manager) ClusterRemoved(cluster string, restConfig *rest.Config) {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	if cacheWrapper, exists := m.caches[cluster]; exists {
		cacheWrapper.cancel()
		delete(m.caches, cluster)
	}
}

func (m *manager) GetCache(cluster string) cache.KubeCoreCache {
	m.cacheAccess.RLock()
	defer m.cacheAccess.RUnlock()
	cw, exists := m.caches[cluster]
	if exists {
		return cw.coreCache
	}
	return nil
}
