package multicluster

import (
	"context"
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
	"k8s.io/client-go/kubernetes"
)

type ClusterId string

type KubeConfigs map[ClusterId]*v1.KubeConfig
type OnNewKubeConfigs func(updated KubeConfigs) error

type KubeConfigHandler struct {
	access   sync.Mutex
	start    func(ctx context.Context) (<-chan error, error)
	callback OnNewKubeConfigs
}

func NewKubeConfigHandler(kube kubernetes.Interface, cache cache.KubeCoreCache) (*KubeConfigHandler, error) {
	kubeConfigClient, err := v1.NewKubeConfigClient(&factory.KubeSecretClientFactory{
		Clientset:       kube,
		Cache:           cache,
		SecretConverter: &secretconverter.KubeConfigSecretConverter{},
	})
	if err != nil {
		return nil, err
	}
	handler := &KubeConfigHandler{}
	emitter := v1.NewKubeconfigsEmitter(kubeConfigClient)
	eventLoop := v1.NewKubeconfigsEventLoop(emitter, &kubeConfigSyncer{handler: handler})
	handler.start = func(ctx context.Context) (errors <-chan error, e error) {
		return eventLoop.Run(nil, clients.WatchOpts{Ctx: ctx})
	}
	return handler, nil
}

func (h *KubeConfigHandler) SetCallback(callback OnNewKubeConfigs) {
	h.access.Lock()
	defer h.access.Unlock()
	h.callback = callback
}

func (h *KubeConfigHandler) Start(ctx context.Context) (<-chan error, error) {
	return h.start(ctx)
}

type kubeConfigSyncer struct {
	handler *KubeConfigHandler
}

func (s *kubeConfigSyncer) Sync(_ context.Context, snap *v1.KubeconfigsSnapshot) error {
	cfgs := snap.Kubeconfigs.List()
	kubeConfigs := make(KubeConfigs)
	for _, cfg := range cfgs {
		kubeConfigs[ClusterId(cfg.GetMetadata().Name)] = cfg
	}
	if s.handler.callback == nil {
		return errors.Errorf("kube config callback has not been defined")
	}
	return s.handler.callback(kubeConfigs)
}
