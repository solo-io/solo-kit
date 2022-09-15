package kube

import (
	"context"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
	v1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/rest"

	"github.com/solo-io/go-utils/contextutils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/solo-io/solo-kit/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	MLists   = stats.Int64("kube/lists", "The number of lists", "1")
	MWatches = stats.Int64("kube/watches", "The number of watches", "1")

	KeyKind, _          = tag.NewKey("kind")
	KeyNamespaceKind, _ = tag.NewKey("ns")

	ListCountView = &view.View{
		Name:        "kube/lists-count",
		Measure:     MLists,
		Description: "The number of list calls",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			KeyKind,
			KeyNamespaceKind,
		},
	}
	WatchCountView = &view.View{
		Name:        "kube/watches-count",
		Measure:     MWatches,
		Description: "The number of watch calls",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			KeyKind,
			KeyNamespaceKind,
		},
	}
)

func init() {
	view.Register(ListCountView, WatchCountView)
}

type SharedCache interface {
	clustercache.ClusterCache

	// Registers the client with the shared cache
	Register(rc *ResourceClient) error
	// RegisterNewNamespace will register the client with a new namespace
	RegisterNewNamespace(namespace string, rc *ResourceClient) error
	// Starts all informers in the factory's registry. Must be idempotent.
	Start()
	// Returns a lister for resources of the given type in the given namespace.
	GetLister(namespace string, obj runtime.Object) (ResourceLister, error)
	// Returns a channel that will receive notifications on changes to resources
	// managed by clients that have registered with the factory.
	// Clients must specify a size for the buffer of the channel they will
	// receive notifications on.
	AddWatch(bufferSize uint) <-chan v1.Resource
	// Removed the given channel from the watches added to the cache
	RemoveWatch(c <-chan v1.Resource)
}

// start uses this context, runs until context gets cancelled
func NewKubeCache(ctx context.Context) SharedCache {
	return &ResourceClientSharedInformerFactory{
		ctx:           ctx,
		defaultResync: 12 * time.Hour,
		registry:      newInformerRegistry(),
		watchTimeout:  time.Second,
	}
}

var _ clustercache.NewClusterCacheForConfig = NewKubeSharedCacheForConfig

func NewKubeSharedCacheForConfig(ctx context.Context, cluster string, restConfig *rest.Config) clustercache.ClusterCache {
	return NewKubeCache(ctx)
}

// The ResourceClientSharedInformerFactory creates a SharedIndexInformer for each of the clients that register with it
// and, when started, creates a kubernetes controller that distributes notifications for changes to the watches that
// have been added to the factory.
// All direct operations on the ResourceClientSharedInformerFactory are synchronized.
// Note it might be best to use this as a Singleton instead of a Structure
// if a Singleton suggestion is proposed, then Register will no long be needed,
// and we can just use Register New Namespace.  Also this would clear up some mutexes, that are being used.
type ResourceClientSharedInformerFactory struct {
	// Contains all the informers managed by this factory
	registry *informerRegistry

	// Default value for how often the informers will resync their caches
	defaultResync time.Duration

	// Indicates whether the factory is started
	started bool

	// the context that was passed to the constructor for the factory.
	// if the context is cancelled, all goroutines started by this cache should be cancelled
	ctx context.Context

	// This allows Start() to be called multiple times safely.
	factoryStarter sync.Once

	// Listeners that need to be notified when a res
	cacheUpdatedWatchers []chan v1.Resource

	// Determines how long the controller will wait for a watch channel to accept an event before aborting the delivery
	watchTimeout time.Duration

	// kubeController is the controller used to watch for events on the informers. It can be used to add new informers too.
	kubeController *controller.Controller

	// registerNamespaceLock is a map of string(namespace) -> sync.Once. It is used to register a new namespace.
	registerNamespaceLock sync.Map

	// registryLock is used when adding or getting information from the registry
	registryLock sync.Mutex
	// startingLock is used when checking isRunning or to lock starting of the SharedInformer
	startingLock              sync.Mutex
	cacheUpdatedWatchersMutex sync.Mutex
}

func NotEmptyValue(ns string) string {
	if ns == "" {
		return "<all>"
	}
	return ns
}

// NewSharedInformer creates a new Shared Index Informer with the list and watch template functions.
func NewSharedInformer(ctx context.Context, resyncPeriod time.Duration, objType runtime.Object,
	listFunc func(options metav1.ListOptions) (runtime.Object, error),
	watchFunc func(context.Context, metav1.ListOptions) (kubewatch.Interface, error)) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				listCtx := ctx
				if ctxWithTags, err := tag.New(listCtx, tag.Insert(KeyOpKind, "list")); err == nil {
					listCtx = ctxWithTags
				}
				stats.Record(listCtx, MLists.M(1), MInFlight.M(1))
				defer stats.Record(listCtx, MInFlight.M(-1))
				return listFunc(options)
			},
			WatchFunc: func(options metav1.ListOptions) (kubewatch.Interface, error) {
				watchCtx := ctx
				if ctxWithTags, err := tag.New(watchCtx, tag.Insert(KeyOpKind, "watch")); err == nil {
					watchCtx = ctxWithTags
				}

				stats.Record(watchCtx, MWatches.M(1), MInFlight.M(1))
				defer stats.Record(watchCtx, MInFlight.M(-1))
				return watchFunc(ctx, options)
			},
		},
		objType,
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
}

// Creates a new SharedIndexInformer and adds it to the factory's informer registry.
// This method is meant to be called once per rc namespace set, please call Register New Namespace when adding new namespaces.
// NOTE: Currently we cannot share informers between resource clients, because the listWatch functions are configured
// with the client's specific token. Hence, we must enforce a one-to-one relationship between informers and clients.
func (f *ResourceClientSharedInformerFactory) Register(rc *ResourceClient) error {
	// because we do not know that we have started yet or not, we need to make a lock
	// once we know, we can proceed
	f.startingLock.Lock()
	ctx := f.ctx
	if f.started {
		f.startingLock.Unlock()
		for _, ns := range rc.namespaceWhitelist {
			if err := f.RegisterNewNamespace(ns, rc); err != nil {
				return err
			}
		}
	} else {
		defer f.startingLock.Unlock()
		if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName)); err == nil {
			ctx = ctxWithTags
		}

		// Create a shared informer for each of the given namespaces.
		// NOTE: We do not distinguish between the value "" (all namespaces) and a regular namespace here.
		// TODO-JAKE there is a special rule, when the namespaces is [""], we need an option
		// to be able to change the Informer so that it does not watch on "", but only those namespaces
		// that we want to watch.  This only happens if the client(user) is setting the namespaceLabelSelectors field in the API.
		f.registryLock.Lock()
		defer f.registryLock.Unlock()

		for _, ns := range rc.namespaceWhitelist {
			// we want to make sure that we have registered this namespace
			f.registerNamespaceLock.LoadOrStore(ns, &sync.Once{})
			if _, err := f.addNewNamespaceToRegistry(ctx, ns, rc); err != nil {
				return err
			}
		}
	}
	return nil
}

// addNewNamespaceToRegistry will create a watch for the resource client type and namespace
func (f *ResourceClientSharedInformerFactory) addNewNamespaceToRegistry(ctx context.Context, ns string, rc *ResourceClient) (cache.SharedIndexInformer, error) {
	nsCtx := ctx
	resourceType := reflect.TypeOf(rc.crd.Version.Type)

	// get the resync value
	resyncPeriod := f.defaultResync
	if rc.resyncPeriod != 0 {
		resyncPeriod = rc.resyncPeriod
	}

	// To nip configuration errors in the bud, error if the registry already contains an informer for the given resource/namespace.
	if f.registry.get(resourceType, ns) != nil {
		return nil, errors.Errorf("Shared cache already contains informer for resource [%v] and namespace [%v]", resourceType, ns)
	}
	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyNamespaceKind, NotEmptyValue(ns))); err == nil {
		nsCtx = ctxWithTags
	}

	listResourceFunc := rc.crdClientset.ResourcesV1().Resources(ns).List
	list := func(options metav1.ListOptions) (runtime.Object, error) {
		return listResourceFunc(ctx, options)
	}
	watch := rc.crdClientset.ResourcesV1().Resources(ns).Watch
	sharedInformer := NewSharedInformer(nsCtx, resyncPeriod, &v1.Resource{}, list, watch)
	f.registry.add(resourceType, ns, sharedInformer)
	return sharedInformer, nil
}

func (f *ResourceClientSharedInformerFactory) IsClusterCache() {}

var cacheSyncTimeout = func() time.Duration {
	timeout := time.Minute
	if timeoutOverride := os.Getenv("SOLOKIT_CACHE_SYNC_TIMEOUT"); timeoutOverride != "" {
		override, err := time.ParseDuration(timeoutOverride)
		if err != nil {
			panic(err)
		}
		timeout = override
	}
	return timeout
}()

// Starts all informers in the factory's registry (if they have not yet been started) and configures the factory to call
// the updateCallback function whenever any of the resources associated with the informers changes.
func (f *ResourceClientSharedInformerFactory) Start() {
	// if starting and managing the locks becomes to much a problem, we can start the factory
	// when the factory is instantiated. That way we will not have multiple processes trying to Start() this structure
	// which should be a singleton.

	// Guarantees that the factory will be started at most once
	f.factoryStarter.Do(func() {
		f.startingLock.Lock()
		defer f.startingLock.Unlock()
		ctx := f.ctx

		// Collect all registered informers

		sharedInformers := f.registry.list()

		// Initialize a new kubernetes controller
		f.kubeController = controller.NewController("solo-resource-controller",
			controller.NewLockingCallbackHandler(f.updatedOccurred), sharedInformers...)

		// Start the controller
		runResult := make(chan error, 1)
		go func() {
			// If there is a problem with the ListWatch, the Run method might wait indefinitely for the informer caches
			// to sync, so we start it in a goroutine to be able to timeout.
			runResult <- f.kubeController.Run(2, ctx.Done())
		}()

		// Fail if the caches have not synchronized after 10 seconds. This prevents the controller from hanging forever.
		var err error
		select {
		case <-ctx.Done():
			return
		case err = <-runResult:
		case <-time.After(cacheSyncTimeout):
			err = errors.Errorf("timed out while waiting for informer caches to sync")
		}

		// If err is non-nil, the kube resource client will panic
		if err != nil {
			contextutils.LoggerFrom(ctx).Panicf("failed to start kube shared informer factory: %v", err)
		}

		// Mark the factory as started
		f.started = true
	})
}

// RegisterNewNamespace is used when the resource client is running. This will add a new namespace to the
// kube controller so that events can be received.
func (f *ResourceClientSharedInformerFactory) RegisterNewNamespace(namespace string, rc *ResourceClient) error {
	f.registryLock.Lock()
	defer f.registryLock.Unlock()

	// because this is an exposed function, Register New Namespace could be called at any time
	// in that event, we want to make sure that the cache has started, the reason is because
	// we have to initiallize the default namespaces as well as this new namespace
	f.Start()
	// we should only register a namespace once and only once
	once, loaded := f.registerNamespaceLock.LoadOrStore(namespace, &sync.Once{})
	if !loaded {
		return nil
	}
	once.(*sync.Once).Do(func() {
		ctx := f.ctx
		if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName)); err == nil {
			ctx = ctxWithTags
		}
		informer, err := f.addNewNamespaceToRegistry(ctx, namespace, rc)
		if err != nil {
			// TODO-JAKE not sure if we want to panic here or to return an error.  But I feel like panics are ok.
			contextutils.LoggerFrom(ctx).Panicf("failed to add new namespace to registry: %v", err)
		}
		if err := f.kubeController.AddNewInformer(informer); err != nil {
			contextutils.LoggerFrom(ctx).Panicf("failed to add new informer to kube controller: %v", err)
		}

	})
	return nil
}

func (f *ResourceClientSharedInformerFactory) GetLister(namespace string, obj runtime.Object) (ResourceLister, error) {
	f.registryLock.Lock()
	defer f.registryLock.Unlock()

	// If the factory (and hence the informers) have not been started, the list operations will be meaningless.
	// Will not happen in our current use of this, but since this method is public it's worth having this check.
	if !f.started {
		return nil, errors.Errorf("cannot get lister for non-running informer")
	}

	// Check if we have informer for this particular namespace
	informer := f.registry.get(reflect.TypeOf(obj), namespace)

	// Check if we have an informer for all namespaces
	if informer == nil {
		informer = f.registry.get(reflect.TypeOf(obj), metav1.NamespaceAll)
	}

	if informer == nil {
		return nil, errors.Errorf("no informer has been registered for ObjectKind %v and namespace %v. "+
			"Make sure that you called Register() on your ResourceClient.", obj.GetObjectKind(), namespace)
	}
	return &resourceLister{indexer: informer.GetIndexer()}, nil
}

// Adds a watch with the given buffer size to the factory.
func (f *ResourceClientSharedInformerFactory) AddWatch(bufferSize uint) <-chan v1.Resource {
	f.cacheUpdatedWatchersMutex.Lock()
	defer f.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan v1.Resource, bufferSize)
	f.cacheUpdatedWatchers = append(f.cacheUpdatedWatchers, c)
	return c
}

// Removes the given watch to the factory.
// A call to this method should be deferred passing the channel returned by AddWatch wherever a watch is added.
func (f *ResourceClientSharedInformerFactory) RemoveWatch(c <-chan v1.Resource) {
	f.cacheUpdatedWatchersMutex.Lock()
	defer f.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range f.cacheUpdatedWatchers {
		if cacheUpdated == c {
			f.cacheUpdatedWatchers = append(f.cacheUpdatedWatchers[:i], f.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

// Not part of the interface (used for testing)
func (f *ResourceClientSharedInformerFactory) Informers() []cache.SharedIndexInformer {
	return f.registry.list()
}

// Not part of the interface (used for testing)
func (f *ResourceClientSharedInformerFactory) IsRunning() bool {
	return f.started
}

// This function will be called when an event occurred for the given resource.
// NOTE: as described in the doc for SharedInformer, we should NOT depend on the contents of the cache exactly matching
// the resource in the notification we received in handler functions. The cache might be MORE fresh than the notification.
func (f *ResourceClientSharedInformerFactory) updatedOccurred(resource v1.Resource) {
	stats.Record(f.ctx, MEvents.M(1))
	f.cacheUpdatedWatchersMutex.Lock()
	defer f.cacheUpdatedWatchersMutex.Unlock()
	for _, watcher := range f.cacheUpdatedWatchers {
		// Attempt delivery asynchronously
		go func(watchChan chan v1.Resource, res v1.Resource) {
			select {
			case <-f.ctx.Done():
				return
			case watchChan <- resource:
			case <-time.After(f.watchTimeout):
				contextutils.LoggerFrom(f.ctx).Debugf("timed out after waiting for %v "+
					"for watch to receive event on resource %v", f.watchTimeout, resource)
			}
		}(watcher, resource)
	}
}

type ResourceLister interface {
	List(selector labels.Selector) (ret []*v1.Resource, err error)
}

type resourceLister struct {
	indexer cache.Indexer
}

func (s *resourceLister) List(selector labels.Selector) (ret []*v1.Resource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Resource))
	})
	return ret, err
}

// Provides convenience methods to add and get elements from the underlying map of maps.
// Operations are NOT thread-safe.
type informerRegistry struct {
	informers map[reflect.Type]map[string]cache.SharedIndexInformer
}

func newInformerRegistry() *informerRegistry {
	return &informerRegistry{
		informers: make(map[reflect.Type]map[string]cache.SharedIndexInformer, 1),
	}
}

// Adds the given informer to the registry. Overwrites existing entries matching the given resourceType and namespace.
func (r *informerRegistry) add(resourceType reflect.Type, namespace string, informer cache.SharedIndexInformer) {

	// Initialize nested map if it does not already exist
	if _, exists := r.informers[resourceType]; !exists {
		r.informers[resourceType] = make(map[string]cache.SharedIndexInformer)
	}

	r.informers[resourceType][namespace] = informer
}

// Retrieve the informer for the given resourceType and namespace.
func (r *informerRegistry) get(resourceType reflect.Type, namespace string) cache.SharedIndexInformer {
	if forResourceType, exists := r.informers[resourceType]; exists {
		return forResourceType[namespace]
	}
	return nil
}

// Returns all the informers contained in the registry.
func (r *informerRegistry) list() []cache.SharedIndexInformer {
	var sharedInformers []cache.SharedIndexInformer
	for _, forResourceType := range r.informers {
		for _, informer := range forResourceType {
			sharedInformers = append(sharedInformers, informer)
		}
	}
	return sharedInformers
}
