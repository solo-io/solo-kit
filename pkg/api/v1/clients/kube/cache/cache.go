package cache

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/go-utils/errors"
	"k8s.io/client-go/tools/cache"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"

	kubeinformers "k8s.io/client-go/informers"
	kubelisters "k8s.io/client-go/listers/core/v1"

	"k8s.io/client-go/kubernetes"
)

type KubeCoreCache interface {
	PodLister() kubelisters.PodLister
	ConfigMapLister() kubelisters.ConfigMapLister
	SecretLister() kubelisters.SecretLister
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type KubeCoreCaches struct {
	initError error

	podLister       kubelisters.PodLister
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

	pods := kubeInformerFactory.Core().V1().Pods()
	configMaps := kubeInformerFactory.Core().V1().ConfigMaps()
	secrets := kubeInformerFactory.Core().V1().Secrets()

	k := &KubeCoreCaches{
		podLister:       pods.Lister(),
		configMapLister: configMaps.Lister(),
		secretLister:    secrets.Lister(),
	}

	kubeController := controller.NewController("kube-plugin-controller",
		controller.NewLockingSyncHandler(k.updatedOccured),
		pods.Informer(), configMaps.Informer(), secrets.Informer())

	stop := ctx.Done()
	go kubeInformerFactory.Start(stop)
	go func() {
		err := kubeController.Run(2, stop)
		if err != nil {
			k.initError = err
		}
	}()

	ok := cache.WaitForCacheSync(stop,
		pods.Informer().HasSynced,
		configMaps.Informer().HasSynced,
		secrets.Informer().HasSynced)
	if !ok {
		// if initError is non-nil, the kube resource client will panic
		k.initError = errors.Errorf("waiting for kube cache sync failed")
	}

	return k
}

func (k *KubeCoreCaches) PodLister() kubelisters.PodLister {
	return k.podLister
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
	c := make(chan struct{}, 10)
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
