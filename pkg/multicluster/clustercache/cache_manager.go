package clustercache

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -destination=./mocks/cache_manager.go -source cache_manager.go -package mocks

type cacheWrapper struct {
	cancel    context.CancelFunc
	coreCache PerClusterCache
}

type PerClusterCache interface {
	IsPerCluster()
}

type FromConfig func(ctx context.Context, cluster string, restConfig *rest.Config) PerClusterCache

type CacheGetter interface {
	GetCache(cluster string, restConfig *rest.Config) PerClusterCache
}

type CacheManager interface {
	handler.ClusterHandler
	CacheGetter
}

type manager struct {
	ctx         context.Context
	caches      map[string]cacheWrapper
	cacheAccess sync.RWMutex
	fromConfig  FromConfig
}

var _ CacheManager = &manager{}

func NewCacheManager(ctx context.Context, fromConfig FromConfig) (*manager, error) {
	if fromConfig == nil {
		return nil, errors.Errorf("cache manager requires a callback for generating per-cluster caches")
	}

	return &manager{
		ctx:         ctx,
		caches:      make(map[string]cacheWrapper),
		cacheAccess: sync.RWMutex{},
		fromConfig:  fromConfig,
	}, nil
}

func (m *manager) ClusterAdded(cluster string, restConfig *rest.Config) {
	// noop -- new caches are added lazily via GetCache
}

func (m *manager) addCluster(cluster string, restConfig *rest.Config) cacheWrapper {
	ctx, cancel := context.WithCancel(m.ctx)
	cw := cacheWrapper{
		cancel:    cancel,
		coreCache: m.fromConfig(ctx, cluster, restConfig),
	}
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	m.caches[cluster] = cw
	return cw
}

func (m *manager) ClusterRemoved(cluster string, restConfig *rest.Config) {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	if cacheWrapper, exists := m.caches[cluster]; exists {
		cacheWrapper.cancel()
		delete(m.caches, cluster)
	}
}

func (m *manager) GetCache(cluster string, restConfig *rest.Config) PerClusterCache {
	m.cacheAccess.RLock()
	cw, exists := m.caches[cluster]
	m.cacheAccess.RUnlock()
	if exists {
		return cw.coreCache
	}
	return m.addCluster(cluster, restConfig).coreCache
}
