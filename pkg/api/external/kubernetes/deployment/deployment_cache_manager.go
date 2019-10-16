package deployment

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type deploymentCacheWrapper struct {
	cancel          context.CancelFunc
	deploymentCache cache.KubeDeploymentCache
}

type KubeDeploymentCacheGetter interface {
	GetCache(cluster string, restConfig *rest.Config) cache.KubeDeploymentCache
}

type KubeDeploymentCacheManager interface {
	handler.ClusterHandler
	KubeDeploymentCacheGetter
}

type deploymentCacheManager struct {
	ctx         context.Context
	caches      map[string]deploymentCacheWrapper
	cacheAccess sync.RWMutex
}

var _ KubeDeploymentCacheManager = &deploymentCacheManager{}

func NewKubeDeploymentCacheManager(ctx context.Context) *deploymentCacheManager {
	return &deploymentCacheManager{
		ctx:         ctx,
		caches:      make(map[string]deploymentCacheWrapper),
		cacheAccess: sync.RWMutex{},
	}
}

func (m *deploymentCacheManager) ClusterAdded(cluster string, restConfig *rest.Config) {
	// noop -- new caches are added lazily via GetCache
}

func (m *deploymentCacheManager) addCluster(cluster string, restConfig *rest.Config) deploymentCacheWrapper {
	ctx, cancel := context.WithCancel(m.ctx)
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return deploymentCacheWrapper{}
	}
	deploymentCache, err := cache.NewKubeDeploymentCache(ctx, kubeClient)
	if err != nil {
		return deploymentCacheWrapper{}
	}

	cw := deploymentCacheWrapper{
		cancel:          cancel,
		deploymentCache: deploymentCache,
	}
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	m.caches[cluster] = cw
	return cw
}

func (m *deploymentCacheManager) ClusterRemoved(cluster string, restConfig *rest.Config) {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	if cacheWrapper, exists := m.caches[cluster]; exists {
		cacheWrapper.cancel()
		delete(m.caches, cluster)
	}
}

func (m *deploymentCacheManager) GetCache(cluster string, restConfig *rest.Config) cache.KubeDeploymentCache {
	m.cacheAccess.RLock()
	cw, exists := m.caches[cluster]
	m.cacheAccess.RUnlock()
	if exists {
		return cw.deploymentCache
	}
	return m.addCluster(cluster, restConfig).deploymentCache
}
