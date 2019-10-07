package multicluster

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"k8s.io/client-go/rest"
)

type cacheWrapper struct {
	cancel    context.CancelFunc
	kubeCache kube.SharedCache
}

type CacheGetter interface {
	GetCache(cluster string) kube.SharedCache
}

type Manager interface {
	ClusterHandler
	CacheGetter
}

type manager struct {
	ctx         context.Context
	caches      map[string]cacheWrapper
	cacheAccess sync.RWMutex
}

var _ Manager = &manager{}

func NewKubeSharedCacheManager(ctx context.Context) *manager {
	return &manager{
		ctx:         ctx,
		caches:      make(map[string]cacheWrapper),
		cacheAccess: sync.RWMutex{},
	}
}

// TODO should this just be a noop since GetCache can handle provisioning new caches?
func (m *manager) ClusterAdded(cluster string, restConfig *rest.Config) {
	m.clusterAdded(cluster)
}

func (m *manager) clusterAdded(cluster string) cacheWrapper {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	ctx, cancel := context.WithCancel(m.ctx)
	cw := cacheWrapper{
		cancel:    cancel,
		kubeCache: kube.NewKubeCache(ctx),
	}
	m.caches[cluster] = cw
	return cw
}

func (m *manager) ClusterRemoved(cluster string, restConfig *rest.Config) {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	cacheWrapper, exists := m.caches[cluster]
	if !exists {
		return
	}
	cacheWrapper.cancel()
	delete(m.caches, cluster)
}

func (m *manager) GetCache(cluster string) kube.SharedCache {
	m.cacheAccess.RLock()
	cw, exists := m.caches[cluster]
	m.cacheAccess.RUnlock()
	if !exists {
		cw = m.clusterAdded(cluster)
		return cw.kubeCache
	}

	return cw.kubeCache
}
