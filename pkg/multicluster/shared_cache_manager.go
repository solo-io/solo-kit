package multicluster

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"k8s.io/client-go/rest"
)

type sharedCacheWrapper struct {
	cancel      context.CancelFunc
	sharedCache kube.SharedCache
}

type KubeSharedCacheGetter interface {
	GetCache(cluster string) kube.SharedCache
}

type KubeSharedCacheManager interface {
	ClusterHandler
	KubeSharedCacheGetter
}

type sharedCacheManager struct {
	ctx         context.Context
	caches      map[string]sharedCacheWrapper
	cacheAccess sync.RWMutex
}

var _ KubeSharedCacheManager = &sharedCacheManager{}

func NewKubeSharedCacheManager(ctx context.Context) *sharedCacheManager {
	return &sharedCacheManager{
		ctx:         ctx,
		caches:      make(map[string]sharedCacheWrapper),
		cacheAccess: sync.RWMutex{},
	}
}

// TODO should this just be a noop since GetCache can handle provisioning new caches?
func (m *sharedCacheManager) ClusterAdded(cluster string, restConfig *rest.Config) {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	ctx, cancel := context.WithCancel(m.ctx)
	cw := sharedCacheWrapper{
		cancel:      cancel,
		sharedCache: kube.NewKubeCache(ctx),
	}
	m.caches[cluster] = cw
}

func (m *sharedCacheManager) ClusterRemoved(cluster string, restConfig *rest.Config) {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	if cacheWrapper, exists := m.caches[cluster]; exists {
		cacheWrapper.cancel()
		delete(m.caches, cluster)
	}
}

func (m *sharedCacheManager) GetCache(cluster string) kube.SharedCache {
	m.cacheAccess.RLock()
	defer m.cacheAccess.RUnlock()
	cw, exists := m.caches[cluster]
	if !exists {
		return nil
	}

}
