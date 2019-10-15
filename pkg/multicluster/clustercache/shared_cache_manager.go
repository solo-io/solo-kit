package clustercache

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -destination=./mocks/shared_cache_manager.go -source shared_cache_manager.go -package mocks

type sharedCacheWrapper struct {
	cancel      context.CancelFunc
	sharedCache kube.SharedCache
}

type KubeSharedCacheGetter interface {
	GetCache(cluster string) kube.SharedCache
}

type KubeSharedCacheManager interface {
	handler.ClusterHandler
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

func (m *sharedCacheManager) ClusterAdded(cluster string, restConfig *rest.Config) {
	// noop -- new caches are added lazily via GetCache
}

func (m *sharedCacheManager) addCluster(cluster string) sharedCacheWrapper {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	ctx, cancel := context.WithCancel(m.ctx)
	cw := sharedCacheWrapper{
		cancel:      cancel,
		sharedCache: kube.NewKubeCache(ctx),
	}
	m.caches[cluster] = cw
	return cw
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
	cw, exists := m.caches[cluster]
	m.cacheAccess.RUnlock()
	if exists {
		return cw.sharedCache
	}
	return m.addCluster(cluster).sharedCache
}
