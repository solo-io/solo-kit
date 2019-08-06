package cache

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
	"go.opencensus.io/tag"

	v1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	kubelisters "k8s.io/client-go/listers/core/v1"

	skkube "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Cache interface {
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type KubeCoreCache interface {
	Cache
	PodLister() kubelisters.PodLister
	ServiceLister() kubelisters.ServiceLister
	ConfigMapLister() kubelisters.ConfigMapLister
	SecretLister() kubelisters.SecretLister
	NamespaceLister() kubelisters.NamespaceLister
}

type kubeCoreCaches struct {
	podListers       map[string]kubelisters.PodLister
	serviceListers   map[string]kubelisters.ServiceLister
	configMapListers map[string]kubelisters.ConfigMapLister
	secretListers    map[string]kubelisters.SecretLister
	namespaceLister  kubelisters.NamespaceLister

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewKubeCoreCache(ctx context.Context, client kubernetes.Interface) (*kubeCoreCaches, error) {
	resyncDuration := 12 * time.Hour
	return NewKubeCoreCacheWithOptions(ctx, client, resyncDuration, []string{metav1.NamespaceAll})
}

func NewKubeCoreCacheWithOptions(ctx context.Context, client kubernetes.Interface, resyncDuration time.Duration, namesapcesToWatch []string) (*kubeCoreCaches, error) {
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(client, resyncDuration)

	var informers []cache.SharedIndexInformer

	pods := map[string]kubelisters.PodLister{}
	services := map[string]kubelisters.ServiceLister{}
	configMaps := map[string]kubelisters.ConfigMapLister{}
	secrets := map[string]kubelisters.SecretLister{}

	for _, nsToWatch := range namesapcesToWatch {
		nsCtx := ctx
		if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyNamespaceKind, skkube.NotEmptyValue(nsToWatch))); err == nil {
			nsCtx = ctxWithTags
		}

		{
			// Pods
			watch := client.CoreV1().Pods(nsToWatch).Watch
			list := func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Pods(nsToWatch).List(options)
			}
			informer := skkube.NewSharedInformer(nsCtx, resyncDuration, &v1.Pod{}, list, watch)
			informers = append(informers, informer)
			lister := kubelisters.NewPodLister(informer.GetIndexer())
			pods[nsToWatch] = lister
		}
		{
			// Services
			watch := client.CoreV1().Services(nsToWatch).Watch
			list := func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Services(nsToWatch).List(options)
			}
			informer := skkube.NewSharedInformer(nsCtx, resyncDuration, &v1.Service{}, list, watch)
			informers = append(informers, informer)
			lister := kubelisters.NewServiceLister(informer.GetIndexer())
			services[nsToWatch] = lister
		}
		{
			// ConfigMap
			watch := client.CoreV1().ConfigMaps(nsToWatch).Watch
			list := func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().ConfigMaps(nsToWatch).List(options)
			}
			informer := skkube.NewSharedInformer(nsCtx, resyncDuration, &v1.ConfigMap{}, list, watch)
			informers = append(informers, informer)
			lister := kubelisters.NewConfigMapLister(informer.GetIndexer())
			configMaps[nsToWatch] = lister
		}
		{
			// Secrets
			watch := client.CoreV1().Secrets(nsToWatch).Watch
			list := func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Secrets(nsToWatch).List(options)
			}
			informer := skkube.NewSharedInformer(nsCtx, resyncDuration, &v1.Secret{}, list, watch)
			informers = append(informers, informer)
			lister := kubelisters.NewSecretLister(informer.GetIndexer())
			secrets[nsToWatch] = lister
		}

	}

	var namespaceLister kubelisters.NamespaceLister
	if len(namesapcesToWatch) == 1 && namesapcesToWatch[0] == metav1.NamespaceAll {
		namespaces := kubeInformerFactory.Core().V1().Namespaces()
		namespaceLister = namespaces.Lister()
		informers = append(informers, namespaces.Informer())
	}

	k := &kubeCoreCaches{
		podListers:       pods,
		serviceListers:   services,
		configMapListers: configMaps,
		secretListers:    secrets,
		namespaceLister:  namespaceLister,
	}

	kubeController := controller.NewController("kube-plugin-controller",
		controller.NewLockingSyncHandler(k.updatedOccured), informers...,
	)

	stop := ctx.Done()
	err := kubeController.Run(2, stop)
	if err != nil {
		return nil, err
	}

	return k, nil
}

func (k *kubeCoreCaches) PodLister() kubelisters.PodLister {
	return k.NamespacedPodLister(metav1.NamespaceAll)
}

func (k *kubeCoreCaches) ServiceLister() kubelisters.ServiceLister {
	return k.NamespacedServiceLister(metav1.NamespaceAll)
}

func (k *kubeCoreCaches) ConfigMapLister() kubelisters.ConfigMapLister {
	return k.NamespacedConfigMapLister(metav1.NamespaceAll)
}

func (k *kubeCoreCaches) SecretLister() kubelisters.SecretLister {
	return k.NamespacedSecretLister(metav1.NamespaceAll)
}

func (k *kubeCoreCaches) NamespaceLister() kubelisters.NamespaceLister {
	return k.namespaceLister
}

func (k *kubeCoreCaches) NamespacedPodLister(ns string) kubelisters.PodLister {
	return k.podListers[ns]
}

func (k *kubeCoreCaches) NamespacedServiceLister(ns string) kubelisters.ServiceLister {
	return k.serviceListers[ns]
}

func (k *kubeCoreCaches) NamespacedConfigMapLister(ns string) kubelisters.ConfigMapLister {
	return k.configMapListers[ns]
}

func (k *kubeCoreCaches) NamespacedSecretLister(ns string) kubelisters.SecretLister {
	return k.secretListers[ns]
}

func (k *kubeCoreCaches) Subscribe() <-chan struct{} {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 10)
	k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers, c)
	return c
}

func (k *kubeCoreCaches) Unsubscribe(c <-chan struct{}) {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range k.cacheUpdatedWatchers {
		if cacheUpdated == c {
			k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers[:i], k.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (k *kubeCoreCaches) updatedOccured() {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range k.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}
