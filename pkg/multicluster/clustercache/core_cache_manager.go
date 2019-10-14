package clustercache

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type coreCacheWrapper struct {
	cancel    context.CancelFunc
	coreCache cache.KubeCoreCache
}

type KubeCoreCacheGetter interface {
	GetCache(cluster string, restConfig *rest.Config) cache.KubeCoreCache
}

type KubeCoreCacheManager interface {
	handler.ClusterHandler
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

func (m *manager) ClusterAdded(cluster string, restConfig *rest.Config) {
	// noop -- new caches are added lazily via GetCache
}

func (m *manager) addCluster(cluster string, restConfig *rest.Config) coreCacheWrapper {
	ctx, cancel := context.WithCancel(m.ctx)
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return coreCacheWrapper{}
	}
	coreCache, err := cache.NewKubeCoreCache(ctx, kubeClient)
	if err != nil {
		return coreCacheWrapper{}
	}

	cw := coreCacheWrapper{
		cancel:    cancel,
		coreCache: coreCache,
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

func (m *manager) GetCache(cluster string, restConfig *rest.Config) cache.KubeCoreCache {
	m.cacheAccess.RLock()
	cw, exists := m.caches[cluster]
	m.cacheAccess.RUnlock()
	if exists {
		return cw.coreCache
	}
	return m.addCluster(cluster, restConfig).coreCache
}
