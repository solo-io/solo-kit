package customresourcedefinition

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
)

type crdCacheWrapper struct {
	cancel   context.CancelFunc
	crdCache KubeCustomResourceDefinitionCache
}

type CrdCacheGetter interface {
	GetCache(cluster string, restConfig *rest.Config) KubeCustomResourceDefinitionCache
}

type CrdCacheManager interface {
	handler.ClusterHandler
	CrdCacheGetter
}

type crdCacheManager struct {
	ctx         context.Context
	caches      map[string]crdCacheWrapper
	cacheAccess sync.RWMutex
}

var _ CrdCacheManager = &crdCacheManager{}

func NewCrdCacheManager(ctx context.Context) *crdCacheManager {
	return &crdCacheManager{
		ctx:         ctx,
		caches:      make(map[string]crdCacheWrapper),
		cacheAccess: sync.RWMutex{},
	}
}

func (m *crdCacheManager) ClusterAdded(cluster string, restConfig *rest.Config) {
	// noop -- new caches are added lazily via GetCache
}

func (m *crdCacheManager) addCluster(cluster string, restConfig *rest.Config) crdCacheWrapper {
	ctx, cancel := context.WithCancel(m.ctx)
	kubeClient, err := apiexts.NewForConfig(restConfig)
	if err != nil {
		return crdCacheWrapper{}
	}
	crdCache, err := NewKubeCustomResourceDefinitionCache(ctx, kubeClient)
	if err != nil {
		return crdCacheWrapper{}
	}

	cw := crdCacheWrapper{
		cancel:   cancel,
		crdCache: crdCache,
	}
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	m.caches[cluster] = cw
	return cw
}

func (m *crdCacheManager) ClusterRemoved(cluster string, restConfig *rest.Config) {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	if cacheWrapper, exists := m.caches[cluster]; exists {
		cacheWrapper.cancel()
		delete(m.caches, cluster)
	}
}

func (m *crdCacheManager) GetCache(cluster string, restConfig *rest.Config) KubeCustomResourceDefinitionCache {
	m.cacheAccess.RLock()
	cw, exists := m.caches[cluster]
	m.cacheAccess.RUnlock()
	if exists {
		return cw.crdCache
	}
	return m.addCluster(cluster, restConfig).crdCache
}
