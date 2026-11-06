package cache

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
	"github.com/solo-io/solo-kit/pkg/errors"
	"go.opencensus.io/tag"

	v1 "k8s.io/api/core/v1"
	kubelisters "k8s.io/client-go/listers/core/v1"

	skkube "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Resource names a kind of object the core cache can list and watch.
type Resource string

const (
	Pods       Resource = "Pods"
	Services   Resource = "Services"
	ConfigMaps Resource = "ConfigMaps"
	Secrets    Resource = "Secrets"
	// Namespaces are cluster-scoped, so the cache watches them only when it
	// watches all namespaces (metav1.NamespaceAll).
	Namespaces Resource = "Namespaces"
)

var allResources = []Resource{Pods, Services, ConfigMaps, Secrets, Namespaces}

// Option configures a core cache created by NewKubeCoreCacheWithOptions.
type Option func(*coreCacheOptions)

type coreCacheOptions struct {
	// resources is nil unless WithResources was given, meaning every resource.
	resources      map[Resource]bool
	fieldSelectors map[Resource]string
	transforms     map[Resource]cache.TransformFunc
}

// WithResources limits the cache to the given resources, so it neither lists,
// watches, nor needs RBAC for the others. The listers of a resource the cache does
// not watch return nil. Without this option the cache watches every resource, and
// Namespaces only when it watches all namespaces; naming Namespaces here while
// watching specific namespaces is an error.
func WithResources(resources ...Resource) Option {
	return func(o *coreCacheOptions) {
		o.resources = map[Resource]bool{}
		for _, r := range resources {
			o.resources[r] = true
		}
	}
}

// WithFieldSelector limits the objects of resource that the cache lists and
// watches to those matching selector. The API server applies it, so excluded
// objects are never sent or stored, and the listers do not return them. Each
// resource supports its own selectable fields; the API server rejects others,
// which leaves the cache unable to sync. A nil or empty selector removes the limit.
func WithFieldSelector(resource Resource, selector fields.Selector) Option {
	return func(o *coreCacheOptions) {
		if o.fieldSelectors == nil {
			o.fieldSelectors = map[Resource]string{}
		}
		o.fieldSelectors[resource] = ""
		if selector != nil {
			o.fieldSelectors[resource] = selector.String()
		}
	}
}

// WithSecretsFieldSelector is WithFieldSelector for Secrets, for example to
// exclude type!=helm.sh/release.v1. Secrets support selecting on metadata.name,
// metadata.namespace, and type only.
func WithSecretsFieldSelector(selector fields.Selector) Option {
	return WithFieldSelector(Secrets, selector)
}

// WithTransformFunc applies transform to each object of resource before the cache
// stores it, for example to drop fields the caller never reads and save memory.
// It follows the contract of client-go's cache.TransformFunc: it may modify and
// return the object it is given. A nil transform removes a previous one.
func WithTransformFunc(resource Resource, transform cache.TransformFunc) Option {
	return func(o *coreCacheOptions) {
		if o.transforms == nil {
			o.transforms = map[Resource]cache.TransformFunc{}
		}
		o.transforms[resource] = transform
	}
}

// watches reports whether the cache watches resource, given whether it watches
// all namespaces.
func (o *coreCacheOptions) watches(resource Resource, allNamespaces bool) bool {
	if resource == Namespaces && !allNamespaces {
		return false
	}
	return o.resources == nil || o.resources[resource]
}

func (o *coreCacheOptions) validate(allNamespaces bool) error {
	known := map[Resource]bool{}
	for _, r := range allResources {
		known[r] = true
	}
	if o.resources != nil {
		if len(o.resources) == 0 {
			return errors.Errorf("WithResources must name at least one resource")
		}
		for r := range o.resources {
			if !known[r] {
				return errors.Errorf("WithResources: unknown resource %q", r)
			}
		}
		if o.resources[Namespaces] && !allNamespaces {
			return errors.Errorf("WithResources: %s can only be watched with metav1.NamespaceAll", Namespaces)
		}
	}
	for r, selector := range o.fieldSelectors {
		if !known[r] {
			return errors.Errorf("WithFieldSelector: unknown resource %q", r)
		}
		if selector != "" && !o.watches(r, allNamespaces) {
			return errors.Errorf("WithFieldSelector: the cache does not watch %q", r)
		}
	}
	for r, transform := range o.transforms {
		if !known[r] {
			return errors.Errorf("WithTransformFunc: unknown resource %q", r)
		}
		if transform != nil && !o.watches(r, allNamespaces) {
			return errors.Errorf("WithTransformFunc: the cache does not watch %q", r)
		}
	}
	return nil
}

// newInformer returns an informer for resource that applies the resource's
// field selector and transform and records solo-kit's list and watch metrics.
func (o *coreCacheOptions) newInformer(ctx context.Context, resource Resource, resyncDuration time.Duration, objType runtime.Object, client kubernetes.Interface,
	list func(context.Context, metav1.ListOptions) (runtime.Object, error),
	watch func(context.Context, metav1.ListOptions) (kubewatch.Interface, error),
) (cache.SharedIndexInformer, error) {
	selector, hasSelector := o.fieldSelectors[resource]
	informer := skkube.NewSharedInformerForClient(ctx, resyncDuration, objType, client,
		func(options metav1.ListOptions) (runtime.Object, error) {
			if hasSelector {
				options.FieldSelector = selector
			}
			return list(ctx, options)
		},
		func(watchCtx context.Context, options metav1.ListOptions) (kubewatch.Interface, error) {
			if hasSelector {
				options.FieldSelector = selector
			}
			return watch(watchCtx, options)
		},
	)
	if transform := o.transforms[resource]; transform != nil {
		if err := informer.SetTransform(transform); err != nil {
			return nil, err
		}
	}
	return informer, nil
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

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

var _ KubeCoreCache = &kubeCoreCaches{}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewKubeCoreCache(ctx context.Context, client kubernetes.Interface) (*kubeCoreCaches, error) {
	resyncDuration := 12 * time.Hour
	return NewKubeCoreCacheWithOptions(ctx, client, resyncDuration, []string{metav1.NamespaceAll})
}

// NewKubeCoreCacheWithOptions creates a core cache of the given namespaces. By
// default it watches Pods, Services, ConfigMaps, and Secrets in each namespace, and
// Namespaces when it watches all namespaces; opts can narrow and configure that.
// The cache runs until ctx is done, and the constructor returns once it has synced.
func NewKubeCoreCacheWithOptions(ctx context.Context, client kubernetes.Interface, resyncDuration time.Duration, namespacesToWatch []string, opts ...Option) (*kubeCoreCaches, error) {
	var cacheOptions coreCacheOptions
	for _, opt := range opts {
		opt(&cacheOptions)
	}

	if len(namespacesToWatch) == 0 {
		namespacesToWatch = []string{metav1.NamespaceAll}
	}

	if len(namespacesToWatch) > 1 {
		if stringutils.ContainsString(metav1.NamespaceAll, namespacesToWatch) {
			return nil, errors.Errorf("if metav1.NamespaceAll is provided, it must be the only one. namespaces provided %v", namespacesToWatch)
		}
	}
	allNamespaces := len(namespacesToWatch) == 1 && namespacesToWatch[0] == metav1.NamespaceAll
	if err := cacheOptions.validate(allNamespaces); err != nil {
		return nil, err
	}

	var informers []cache.SharedIndexInformer
	// addInformer tags ctx with the metric kind of resource, creates its informer,
	// and registers it to run.
	addInformer := func(ctx context.Context, resource Resource, metricKind string, objType runtime.Object,
		list func(context.Context, metav1.ListOptions) (runtime.Object, error),
		watch func(context.Context, metav1.ListOptions) (kubewatch.Interface, error),
	) (cache.Indexer, error) {
		typeCtx := ctx
		if ctxWithTags, err := tag.New(ctx, tag.Insert(skkube.KeyKind, metricKind)); err == nil {
			typeCtx = ctxWithTags
		}
		informer, err := cacheOptions.newInformer(typeCtx, resource, resyncDuration, objType, client, list, watch)
		if err != nil {
			return nil, err
		}
		informers = append(informers, informer)
		return informer.GetIndexer(), nil
	}

	pods := map[string]kubelisters.PodLister{}
	services := map[string]kubelisters.ServiceLister{}
	configMaps := map[string]kubelisters.ConfigMapLister{}
	secrets := map[string]kubelisters.SecretLister{}

	for _, nsToWatch := range namespacesToWatch {
		nsToWatch := nsToWatch
		nsCtx := ctx
		if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyNamespaceKind, skkube.NotEmptyValue(nsToWatch))); err == nil {
			nsCtx = ctxWithTags
		}

		if cacheOptions.watches(Pods, allNamespaces) {
			c := client.CoreV1().Pods(nsToWatch)
			indexer, err := addInformer(nsCtx, Pods, "Pods", &v1.Pod{},
				func(ctx context.Context, o metav1.ListOptions) (runtime.Object, error) { return c.List(ctx, o) }, c.Watch)
			if err != nil {
				return nil, err
			}
			pods[nsToWatch] = kubelisters.NewPodLister(indexer)
		}
		if cacheOptions.watches(Services, allNamespaces) {
			c := client.CoreV1().Services(nsToWatch)
			indexer, err := addInformer(nsCtx, Services, "Services", &v1.Service{},
				func(ctx context.Context, o metav1.ListOptions) (runtime.Object, error) { return c.List(ctx, o) }, c.Watch)
			if err != nil {
				return nil, err
			}
			services[nsToWatch] = kubelisters.NewServiceLister(indexer)
		}
		if cacheOptions.watches(ConfigMaps, allNamespaces) {
			c := client.CoreV1().ConfigMaps(nsToWatch)
			indexer, err := addInformer(nsCtx, ConfigMaps, "ConfigMap", &v1.ConfigMap{},
				func(ctx context.Context, o metav1.ListOptions) (runtime.Object, error) { return c.List(ctx, o) }, c.Watch)
			if err != nil {
				return nil, err
			}
			configMaps[nsToWatch] = kubelisters.NewConfigMapLister(indexer)
		}
		if cacheOptions.watches(Secrets, allNamespaces) {
			c := client.CoreV1().Secrets(nsToWatch)
			indexer, err := addInformer(nsCtx, Secrets, "Secrets", &v1.Secret{},
				func(ctx context.Context, o metav1.ListOptions) (runtime.Object, error) { return c.List(ctx, o) }, c.Watch)
			if err != nil {
				return nil, err
			}
			secrets[nsToWatch] = kubelisters.NewSecretLister(indexer)
		}
	}

	var namespaceLister kubelisters.NamespaceLister
	if cacheOptions.watches(Namespaces, allNamespaces) {
		nsCtx := ctx
		if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyNamespaceKind, skkube.NotEmptyValue(metav1.NamespaceAll))); err == nil {
			nsCtx = ctxWithTags
		}
		c := client.CoreV1().Namespaces()
		indexer, err := addInformer(nsCtx, Namespaces, "Namespaces", &v1.Namespace{},
			func(ctx context.Context, o metav1.ListOptions) (runtime.Object, error) { return c.List(ctx, o) }, c.Watch)
		if err != nil {
			return nil, err
		}
		namespaceLister = kubelisters.NewNamespaceLister(indexer)
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
