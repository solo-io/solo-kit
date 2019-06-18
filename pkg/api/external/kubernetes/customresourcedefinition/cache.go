package customresourcedefinition

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextsinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	apiextslisters "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1beta1"
)

type KubeCustomResourceDefinitionCache interface {
	CustomResourceDefinitionLister() apiextslisters.CustomResourceDefinitionLister
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type kubeCustomResourceDefinitionCache struct {
	customResourceDefinitionLister apiextslisters.CustomResourceDefinitionLister

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewKubeCustomResourceDefinitionCache(ctx context.Context, apiExtsClient apiexts.Interface) (*kubeCustomResourceDefinitionCache, error) {
	resyncDuration := 12 * time.Hour

	apiExtsInformerFactory := apiextsinformers.NewSharedInformerFactory(apiExtsClient, resyncDuration)

	customResourceDefinitions := apiExtsInformerFactory.Apiextensions().V1beta1().CustomResourceDefinitions()

	k := &kubeCustomResourceDefinitionCache{
		customResourceDefinitionLister: customResourceDefinitions.Lister(),
	}

	kubeController := controller.NewController("kube-plugin-controller",
		controller.NewLockingSyncHandler(k.updatedOccured),
		customResourceDefinitions.Informer(),
	)

	stop := ctx.Done()
	if err := kubeController.Run(2, stop); err != nil {
		return nil, err
	}

	return k, nil
}

func (k *kubeCustomResourceDefinitionCache) CustomResourceDefinitionLister() apiextslisters.CustomResourceDefinitionLister {
	return k.customResourceDefinitionLister
}

func (k *kubeCustomResourceDefinitionCache) Subscribe() <-chan struct{} {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 10)
	k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers, c)
	return c
}

func (k *kubeCustomResourceDefinitionCache) Unsubscribe(c <-chan struct{}) {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range k.cacheUpdatedWatchers {
		if cacheUpdated == c {
			k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers[:i], k.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (k *kubeCustomResourceDefinitionCache) updatedOccured() {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range k.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}
