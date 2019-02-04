package cache

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"

	kubeinformers "k8s.io/client-go/informers"
	kubelisters "k8s.io/client-go/listers/core/v1"

	"k8s.io/client-go/kubernetes"
)

type KubeCoreCache interface {
	ConfigMapLister() kubelisters.ConfigMapLister
	SecretLister() kubelisters.SecretLister
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type KubeCoreCaches struct {
	initError error

	configMapLister kubelisters.ConfigMapLister
	secretLister    kubelisters.SecretLister

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewKubeCoreCache(ctx context.Context, client kubernetes.Interface) *KubeCoreCaches {
	resyncDuration := 12 * time.Hour
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(client, resyncDuration)

	configMaps := kubeInformerFactory.Core().V1().ConfigMaps()
	secrets := kubeInformerFactory.Core().V1().Secrets()
	k := &KubeCoreCaches{
		configMapLister: configMaps.Lister(),
		secretLister:    secrets.Lister(),
	}

	kubeController := controller.NewController("kube-plugin-controller",
		controller.NewLockingSyncHandler(k.updatedOccured),
		configMaps.Informer(), secrets.Informer())

	stop := ctx.Done()
	go kubeInformerFactory.Start(stop)
	go func() {
		err := kubeController.Run(2, stop)
		if err != nil {
			k.initError = err
		}
	}()

	ok := cache.WaitForCacheSync(stop,
		configMaps.Informer().HasSynced,
		secrets.Informer().HasSynced)
	if !ok {
		// if initError is non-nil, the kube resource client will panic
		k.initError = errors.Errorf("waiting for kube pod, endpoints, services cache sync failed")
	}

	return k
}

func (k *KubeCoreCaches) ConfigMapLister() kubelisters.ConfigMapLister {
	return k.configMapLister
}

func (k *KubeCoreCaches) SecretLister() kubelisters.SecretLister {
	return k.secretLister
}

func (k *KubeCoreCaches) Subscribe() <-chan struct{} {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 1)
	k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers, c)
	return c
}

func (k *KubeCoreCaches) Unsubscribe(c <-chan struct{}) {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range k.cacheUpdatedWatchers {
		if cacheUpdated == c {
			k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers[:i], k.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (k *KubeCoreCaches) updatedOccured() {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range k.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}
