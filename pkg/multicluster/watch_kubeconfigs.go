package multicluster

import (
	"context"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
	"k8s.io/client-go/kubernetes"
)

type ClusterId core.ResourceRef

// empty ClusterId refers to local
var LocalCluster = ClusterId{}

type KubeConfigs map[ClusterId]*v1.KubeConfig

func WatchKubeConfigs(ctx context.Context, kube kubernetes.Interface, cache cache.KubeCoreCache) (<-chan KubeConfigs, <-chan error, error) {
	kubeConfigClient, err := v1.NewKubeConfigClient(&factory.KubeSecretClientFactory{
		Clientset:       kube,
		Cache:           cache,
		SecretConverter: &secretconverter.KubeConfigSecretConverter{},
	})
	if err != nil {
		return nil, nil, err
	}
	emitter := v1.NewKubeconfigsEmitter(kubeConfigClient)
	kubeConfigsChan := make(chan KubeConfigs)
	eventLoop := v1.NewKubeconfigsEventLoop(emitter, &kubeConfigSyncer{kubeConfigsChan: kubeConfigsChan})
	errs, err := eventLoop.Run(nil, clients.WatchOpts{Ctx: ctx})
	if err != nil {
		return nil, nil, err
	}
	return kubeConfigsChan, errs, nil
}

type kubeConfigSyncer struct {
	kubeConfigsChan chan KubeConfigs
}

func (s *kubeConfigSyncer) Sync(_ context.Context, snap *v1.KubeconfigsSnapshot) error {
	cfgs := snap.Kubeconfigs.List()
	kubeConfigs := make(KubeConfigs)
	for _, cfg := range cfgs {
		kubeConfigs[ClusterId(cfg.GetMetadata().Ref())] = cfg
	}
	s.kubeConfigsChan <- kubeConfigs
	return nil
}
