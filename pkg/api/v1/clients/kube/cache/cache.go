package cache

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"go.opencensus.io/tag"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/core/v1"
	kubelisters "k8s.io/client-go/listers/core/v1"

	skkube "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// onceAndSent is used to capture errs made by once functions
type onceAndSent struct {
	Err  error
	Once *sync.Once
}

type ServiceLister interface {
	// List lists all Services in the indexer.
	List(selector labels.Selector) (ret []*v1.Service, err error)
}

type PodLister interface {
	// List lists all Pods in the indexer.
	List(selector labels.Selector) (ret []*v1.Pod, err error)
}

type ConfigMapLister interface {
	// List lists all ConfigMaps in the indexer.
	List(selector labels.Selector) (ret []*v1.ConfigMap, err error)
}

type SecretLister interface {
	// List lists all Secrets in the indexer.
	List(selector labels.Selector) (ret []*v1.Secret, err error)
}

type Cache interface {
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type KubeCoreCache interface {
	Cache
	clustercache.ClusterCache

	// RegisterNewNamespaceCache will register the namespace so that the resources
	// are available in the cache listers.
	RegisterNewNamespaceCache(ns string) error
	// Deprecated: Use NamespacedPodLister instead
	PodLister() kubelisters.PodLister
	// Deprecated: Use NamespacedServiceLister instead
	ServiceLister() kubelisters.ServiceLister
	// Deprecated: Use NamespacedConfigMapLister instead
	ConfigMapLister() kubelisters.ConfigMapLister
	// Deprecated: Use NamespacedSecretLister instead
	SecretLister() kubelisters.SecretLister

	NamespaceLister() kubelisters.NamespaceLister

	NamespacedPodLister(ns string) PodLister
	NamespacedServiceLister(ns string) ServiceLister
	NamespacedConfigMapLister(ns string) ConfigMapLister
	NamespacedSecretLister(ns string) SecretLister
}

type kubeCoreCaches struct {
	podListers       map[string]kubelisters.PodLister
	serviceListers   map[string]kubelisters.ServiceLister
	configMapListers map[string]kubelisters.ConfigMapLister
	secretListers    map[string]kubelisters.SecretLister
	namespaceLister  kubelisters.NamespaceLister
	// ctx is the context of the cache
	ctx context.Context
	// client kubernetes client
	client kubernetes.Interface
	// kubeController is the controller used to start the informers, and is used to
	// watch events that occur on the informers.  This is used to send information back to the
	// [resource]listers.
	kubeController *controller.Controller
	// resyncDuration is the time
	resyncDuration time.Duration
	// informers are the kube resources that provide events
	informers []cache.SharedIndexInformer

	// registerNamespaceLock is a map string(namespace) -> sync.Once. Is used to register namespaces only once.
	registerNamespaceLock sync.Map

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

var _ KubeCoreCache = &kubeCoreCaches{}

func NewCoreCacheForConfig(ctx context.Context, cluster string, restConfig *rest.Config) clustercache.ClusterCache {
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil
	}
	c, err := NewKubeCoreCache(ctx, kubeClient)
	if err != nil {
		return nil
	}
	return c
}

var _ clustercache.NewClusterCacheForConfig = NewCoreCacheForConfig

// creates the namespace lister.
func NewFromConfigWithOptions(resyncDuration time.Duration, namesapcesToWatch []string) clustercache.NewClusterCacheForConfig {
	return func(ctx context.Context, cluster string, restConfig *rest.Config) clustercache.ClusterCache {
		kubeClient, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return nil
		}
		c, err := NewKubeCoreCacheWithOptions(ctx, kubeClient, resyncDuration, namesapcesToWatch, true)
		if err != nil {
			return nil
		}
		return c
	}
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewKubeCoreCache(ctx context.Context, client kubernetes.Interface) (*kubeCoreCaches, error) {
	resyncDuration := 12 * time.Hour
	return NewKubeCoreCacheWithOptions(ctx, client, resyncDuration, []string{metav1.NamespaceAll}, true)
}

func NewKubeCoreCacheWithOptions(ctx context.Context, client kubernetes.Interface, resyncDuration time.Duration, namesapcesToWatch []string, createNamespaceLister bool) (*kubeCoreCaches, error) {

	if len(namesapcesToWatch) == 0 {
		namesapcesToWatch = []string{metav1.NamespaceAll}
	}

	if len(namesapcesToWatch) > 1 {
		if stringutils.ContainsString(metav1.NamespaceAll, namesapcesToWatch) {
			return nil, errors.Errorf("if metav1.NamespaceAll is provided, it must be the only one. namespaces provided %v", namesapcesToWatch)
		}
	}

	var informers []cache.SharedIndexInformer

	pods := map[string]kubelisters.PodLister{}
	services := map[string]kubelisters.ServiceLister{}
	configMaps := map[string]kubelisters.ConfigMapLister{}
	secrets := map[string]kubelisters.SecretLister{}

	k := &kubeCoreCaches{
		podListers:       pods,
		serviceListers:   services,
		configMapListers: configMaps,
		secretListers:    secrets,
		client:           client,
		ctx:              ctx,
		resyncDuration:   resyncDuration,
		informers:        informers,
	}

	for _, nsToWatch := range namesapcesToWatch {
		k.addNewNamespace(nsToWatch)
	}

	// since we now allow namespaces to be registered dynamically we will always create the namespace lister
	k.addNamespaceLister()

	k.kubeController = controller.NewController("kube-plugin-controller",
		controller.NewLockingSyncHandler(k.updatedOccured), k.informers...,
	)

	stop := ctx.Done()
	err := k.kubeController.Run(2, stop)
	if err != nil {
		return nil, err
	}

	return k, nil
}

func (c *kubeCoreCaches) addPod(namespace string, typeCtx context.Context) cache.SharedIndexInformer {
	if ctxWithTags, err := tag.New(typeCtx, tag.Insert(skkube.KeyKind, "Pods")); err == nil {
		typeCtx = ctxWithTags
	}
	watch := c.client.CoreV1().Pods(namespace).Watch
	list := func(options metav1.ListOptions) (runtime.Object, error) {
		return c.client.CoreV1().Pods(namespace).List(c.ctx, options)
	}
	informer := skkube.NewSharedInformer(typeCtx, c.resyncDuration, &v1.Pod{}, list, watch)
	c.informers = append(c.informers, informer)
	lister := kubelisters.NewPodLister(informer.GetIndexer())
	c.podListers[namespace] = lister
	return informer
}

func (c *kubeCoreCaches) addService(namespace string, typeCtx context.Context) cache.SharedIndexInformer {
	if ctxWithTags, err := tag.New(typeCtx, tag.Insert(skkube.KeyKind, "Services")); err == nil {
		typeCtx = ctxWithTags
	}
	watch := c.client.CoreV1().Services(namespace).Watch
	list := func(options metav1.ListOptions) (runtime.Object, error) {
		return c.client.CoreV1().Services(namespace).List(c.ctx, options)
	}
	informer := skkube.NewSharedInformer(typeCtx, c.resyncDuration, &v1.Service{}, list, watch)
	c.informers = append(c.informers, informer)
	lister := kubelisters.NewServiceLister(informer.GetIndexer())
	c.serviceListers[namespace] = lister
	return informer
}

func (c *kubeCoreCaches) addConfigMap(namespace string, typeCtx context.Context) cache.SharedIndexInformer {
	if ctxWithTags, err := tag.New(typeCtx, tag.Insert(skkube.KeyKind, "ConfigMap")); err == nil {
		typeCtx = ctxWithTags
	}
	watch := c.client.CoreV1().ConfigMaps(namespace).Watch
	list := func(options metav1.ListOptions) (runtime.Object, error) {
		return c.client.CoreV1().ConfigMaps(namespace).List(c.ctx, options)
	}
	informer := skkube.NewSharedInformer(typeCtx, c.resyncDuration, &v1.ConfigMap{}, list, watch)
	c.informers = append(c.informers, informer)
	lister := kubelisters.NewConfigMapLister(informer.GetIndexer())
	c.configMapListers[namespace] = lister
	return informer
}

func (c *kubeCoreCaches) addSecret(namespace string, typeCtx context.Context) cache.SharedIndexInformer {
	if ctxWithTags, err := tag.New(typeCtx, tag.Insert(skkube.KeyKind, "Secrets")); err == nil {
		typeCtx = ctxWithTags
	}
	watch := c.client.CoreV1().Secrets(namespace).Watch
	list := func(options metav1.ListOptions) (runtime.Object, error) {
		return c.client.CoreV1().Secrets(namespace).List(c.ctx, options)
	}
	informer := skkube.NewSharedInformer(typeCtx, c.resyncDuration, &v1.Secret{}, list, watch)
	c.informers = append(c.informers, informer)
	lister := kubelisters.NewSecretLister(informer.GetIndexer())
	c.secretListers[namespace] = lister
	return informer
}

func (c *kubeCoreCaches) addNamespaceLister() {
	watch := c.client.CoreV1().Namespaces().Watch
	list := func(options metav1.ListOptions) (runtime.Object, error) {
		return c.client.CoreV1().Namespaces().List(c.ctx, options)
	}
	nsCtx := c.ctx
	if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyNamespaceKind, skkube.NotEmptyValue(metav1.NamespaceAll)), tag.Insert(skkube.KeyKind, "Namespaces")); err == nil {
		nsCtx = ctxWithTags
	}
	informer := skkube.NewSharedInformer(nsCtx, c.resyncDuration, &v1.Namespace{}, list, watch)
	c.informers = append(c.informers, informer)
	c.namespaceLister = kubelisters.NewNamespaceLister(informer.GetIndexer())
}

func (k *kubeCoreCaches) addNewNamespace(namespace string) []cache.SharedIndexInformer {
	nsCtx := k.ctx
	if ctxWithTags, err := tag.New(k.ctx, tag.Insert(skkube.KeyNamespaceKind, skkube.NotEmptyValue(namespace))); err == nil {
		nsCtx = ctxWithTags
	}
	podInformer := k.addPod(namespace, nsCtx)
	serviceInformer := k.addService(namespace, nsCtx)
	configMapInformer := k.addConfigMap(namespace, nsCtx)
	secretInformer := k.addSecret(namespace, nsCtx)
	return []cache.SharedIndexInformer{podInformer, serviceInformer, configMapInformer, secretInformer}
}

// RegisterNewNamespaceCache will create the cache informers for each resource type
// this will add the informer to the kube controller so that events can be watched.
func (k *kubeCoreCaches) RegisterNewNamespaceCache(namespace string) error {
	once, _ := k.registerNamespaceLock.LoadOrStore(namespace, &onceAndSent{Once: &sync.Once{}})
	onceFunc := once.(*onceAndSent)
	onceFunc.Once.Do(func() {
		informers := k.addNewNamespace(namespace)
		if err := k.kubeController.AddNewListOfInformers(informers); err != nil {
			onceFunc.Err = errors.Wrapf(err, "failed to add new list of informers to kube controller")
		}
	})
	return onceFunc.Err
}

// Deprecated: Use NamespacedPodLister instead
func (k *kubeCoreCaches) PodLister() kubelisters.PodLister {
	return k.podListers[metav1.NamespaceAll]
}

// Deprecated: Use NamespacedServiceLister instead
func (k *kubeCoreCaches) ServiceLister() kubelisters.ServiceLister {
	return k.serviceListers[metav1.NamespaceAll]
}

// Deprecated: Use NamespacedConfigMapLister instead
func (k *kubeCoreCaches) ConfigMapLister() kubelisters.ConfigMapLister {
	return k.configMapListers[metav1.NamespaceAll]
}

// Deprecated: Use NamespacedSecretLister instead
func (k *kubeCoreCaches) SecretLister() kubelisters.SecretLister {
	return k.secretListers[metav1.NamespaceAll]
}

// NamespaceLister() will return a non-null lister only if we watch all namespaces.
func (k *kubeCoreCaches) NamespaceLister() kubelisters.NamespaceLister {
	return k.namespaceLister
}

func (k *kubeCoreCaches) NamespacedPodLister(ns string) PodLister {
	if lister, ok := k.podListers[metav1.NamespaceAll]; ok {
		return lister.Pods(ns)
	}
	return k.podListers[ns]
}

func (k *kubeCoreCaches) NamespacedServiceLister(ns string) ServiceLister {
	if lister, ok := k.serviceListers[metav1.NamespaceAll]; ok {
		return lister.Services(ns)
	}
	return k.serviceListers[ns]
}

func (k *kubeCoreCaches) NamespacedConfigMapLister(ns string) ConfigMapLister {
	if lister, ok := k.configMapListers[metav1.NamespaceAll]; ok {
		return lister.ConfigMaps(ns)
	}
	return k.configMapListers[ns]
}

func (k *kubeCoreCaches) NamespacedSecretLister(ns string) SecretLister {
	if lister, ok := k.secretListers[metav1.NamespaceAll]; ok {
		return lister.Secrets(ns)
	}
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

func (k *kubeCoreCaches) IsClusterCache() {}

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
