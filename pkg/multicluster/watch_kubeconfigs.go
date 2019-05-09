package multicluster

import (
	"context"

	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	"k8s.io/client-go/kubernetes"
)

// empty ClusterId refers to local
const LocalCluster = ""

type KubeConfigWatcher interface {
	WatchKubeConfigs(ctx context.Context, kube kubernetes.Interface, cache cache.KubeCoreCache) (<-chan v1.KubeConfigList, <-chan error, error)
}

type defaultKubeConfigWatcher struct{}

func NewKubeConfigWatcher() KubeConfigWatcher {
	return &defaultKubeConfigWatcher{}
}

func (kcw *defaultKubeConfigWatcher) WatchKubeConfigs(ctx context.Context, kube kubernetes.Interface, cache cache.KubeCoreCache) (<-chan v1.KubeConfigList, <-chan error, error) {
	return WatchKubeConfigs(ctx, kube, cache)
}

func WatchKubeConfigs(ctx context.Context, kube kubernetes.Interface, cache cache.KubeCoreCache) (<-chan v1.KubeConfigList, <-chan error, error) {
	kubeConfigClient, err := v1.NewKubeConfigClient(&factory.KubeSecretClientFactory{
		Clientset:       kube,
		Cache:           cache,
		SecretConverter: &secretconverter.KubeConfigSecretConverter{},
	})
	if err != nil {
		return nil, nil, err
	}
	return kubeConfigClient.Watch("", clients.WatchOpts{Ctx: ctx})
}
