package config

import (
	"context"
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/errutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type RestConfigs map[string]*rest.Config

type RestConfigHandler struct {
	kcWatcher     multicluster.KubeConfigWatcher
	handlers      []handler.ClusterHandler
	cache         RestConfigs
	cacheAccess   sync.Mutex
	handlerAccess sync.Mutex
}

func NewRestConfigHandler(kcWatcher multicluster.KubeConfigWatcher, handlers ...handler.ClusterHandler) *RestConfigHandler {
	return &RestConfigHandler{kcWatcher: kcWatcher, handlers: handlers}
}

func (h *RestConfigHandler) Run(ctx context.Context, local *rest.Config, kubeClient kubernetes.Interface, kubeCache cache.KubeCoreCache) (<-chan error, error) {
	kubeConfigs, errs, err := multicluster.WatchKubeConfigs(ctx, kubeClient, kubeCache)
	if err != nil {
		return nil, err
	}

	ourErrs := make(chan error)
	go errutils.AggregateErrs(ctx, ourErrs, errs, "watching kubernetes *rest.Configs")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case kcs := <-kubeConfigs:
				restConfigs, err := parseRestConfigs(local, kcs)
				if err != nil {
					ourErrs <- err
					continue
				}

				h.handleNewRestConfigs(restConfigs)
			}
		}
	}()

	return errs, nil
}

func (h *RestConfigHandler) handleNewRestConfigs(cfgs RestConfigs) {
	h.cacheAccess.Lock()
	defer h.cacheAccess.Unlock()
	for cluster, oldCfg := range h.cache {
		if _, persisted := cfgs[cluster]; persisted {
			continue
		}
		h.clusterRemoved(cluster, oldCfg)
	}
	for cluster, newCfg := range cfgs {
		if _, exists := h.cache[cluster]; exists {
			continue
		}
		h.clusterAdded(cluster, newCfg)
	}
	// update cache
	h.cache = cfgs
}

func (h *RestConfigHandler) clusterAdded(cluster string, cfg *rest.Config) {
	h.handlerAccess.Lock()
	defer h.handlerAccess.Unlock()
	for _, handler := range h.handlers {
		handler.ClusterAdded(cluster, cfg)
	}
}

func (h *RestConfigHandler) clusterRemoved(cluster string, cfg *rest.Config) {
	h.handlerAccess.Lock()
	defer h.handlerAccess.Unlock()
	for _, handler := range h.handlers {
		handler.ClusterRemoved(cluster, cfg)
	}
}

func parseRestConfigs(local *rest.Config, kcs v1.KubeConfigList) (RestConfigs, error) {
	cfgs := RestConfigs{}
	if local != nil {
		cfgs[multicluster.LocalCluster] = local
	}

	for _, kc := range kcs {
		raw, err := clientcmd.Write(kc.Config)
		if err != nil {
			return nil, err
		}
		restCfg, err := clientcmd.RESTConfigFromKubeConfig(raw)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to construct *rest.Config from kubeconfig %v", kc.Metadata.Ref())
		}
		cfgs[kc.Cluster] = restCfg
	}
	return cfgs, nil
}
