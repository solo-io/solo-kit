package kube

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/solo-io/solo-kit/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	MLists   = stats.Int64("kube/lists", "The number of lists", "1")
	MWatches = stats.Int64("kube/lists", "The number of watches", "1")

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
		Description: "The number of list calls",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			KeyKind,
			KeyNamespaceKind,
		},
	}
)

func init() {
	view.Register(ListCountView)
}

type ResourceLister interface {
	List(selector labels.Selector) (ret []*v1.Resource, err error)
}

type ResourceClientSharedInformerFactory struct {
	initError error

	lock          sync.Mutex
	defaultResync time.Duration

	informers map[reflect.Type]map[string]cache.SharedIndexInformer

	started bool

	// This allows Start() to be called multiple times safely.
	factoryStarter sync.Once
}

func NewResourceClientSharedInformerFactory() *ResourceClientSharedInformerFactory {
	return &ResourceClientSharedInformerFactory{
		defaultResync: 12 * time.Hour,
		informers:     make(map[reflect.Type]map[string]cache.SharedIndexInformer),
	}
}
func (f *ResourceClientSharedInformerFactory) InitErr() error {
	return f.initError
}

// Creates a new SharedIndexInformer and adds it to the factory's informer registry.
// NOTE: Currently we cannot share informers between resource clients, because the listWatch functions are configured
// with the client's specific token. Hence, we must enforce a one-to-one relationship between informers and clients.
func (f *ResourceClientSharedInformerFactory) Register(rc *ResourceClient) error {
	ctx := context.TODO()
	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName)); err == nil {
		ctx = ctxWithTags
	}

	informerType := reflect.TypeOf(rc.crd.Type)
	namespaces := rc.namespaces // will always contain at least one element

	resyncPeriod := f.defaultResync
	if rc.resyncPeriod != 0 {
		resyncPeriod = rc.resyncPeriod
	}

	// Create a shared informer for each of the given namespaces.
	// NOTE: We do not distinguish between the value "" (all namespaces) and a regular namespace here.
	for _, ns := range namespaces {

		// To nip configuration errors in the bud, error if the registry already contains an informer for the given resource/namespace.
		if forResourceType, exists := f.informers[informerType]; exists {
			if _, exists := forResourceType[ns]; exists {
				return errors.Errorf("Shared cache already contains informer for resource [%v] and namespace [%v]", informerType, ns)
			}
		}

		if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyNamespaceKind, ns)); err == nil {
			ctx = ctxWithTags
		}

		list := rc.kube.ResourcesV1().Resources(ns).List
		watch := rc.kube.ResourcesV1().Resources(ns).Watch
		sharedInformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyOpKind, "list")); err == nil {
						ctx = ctxWithTags
					}
					stats.Record(ctx, MLists.M(1), MInFlight.M(1))
					defer stats.Record(ctx, MInFlight.M(-1))
					return list(options)
				},
				WatchFunc: func(options metav1.ListOptions) (kubewatch.Interface, error) {
					if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyOpKind, "watch")); err == nil {
						ctx = ctxWithTags
					}

					stats.Record(ctx, MWatches.M(1), MInFlight.M(1))
					defer stats.Record(ctx, MInFlight.M(-1))
					return watch(options)
				},
			},
			&v1.Resource{}, // TODO(yuval-k): can we make this rc.crd.Type ?
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)

		f.registerInformer(ctx, ns, informerType, sharedInformer)
	}
	return nil
}

// Add the given informer to the factory's internal informer registry
func (f *ResourceClientSharedInformerFactory) registerInformer(ctx context.Context, namespace string, informerType reflect.Type, newInformer cache.SharedIndexInformer) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.started {
		contextutils.LoggerFrom(ctx).DPanic("can't register informer after factory has started. This may change in the future.")
	}

	// Initialize nested map if it does not already exist
	if _, exists := f.informers[informerType]; !exists {
		f.informers[informerType] = make(map[string]cache.SharedIndexInformer)
	}

	f.informers[informerType][namespace] = newInformer
	return
}

// Starts all informers in the factory's registry (if they have not yet been started) and configures the factory to call
// the given updateCallback function whenever any of the resources associated with the informers changes.
func (f *ResourceClientSharedInformerFactory) Start(ctx context.Context, kubeClient kubernetes.Interface, updateCallback func(v1.Resource)) {

	// Guarantees that the factory will be started at most once
	f.factoryStarter.Do(func() {

		// Collect all registered informers
		var sharedInformers []cache.SharedInformer
		for _, informersByNamespace := range f.informers {
			for _, informer := range informersByNamespace {
				sharedInformers = append(sharedInformers, informer)
			}
		}

		// Initialize a new kubernetes controller
		kubeController := controller.NewController("solo-resource-controller", kubeClient,
			controller.NewLockingCallbackHandler(updateCallback), sharedInformers...)

		// Start the controller
		runResult := make(chan error, 1)
		go func() {
			// If there is a problem with the ListWatch, the Run method might wait indefinitely for the informer caches
			// to sync, so we start it in a goroutine to be able to timeout.
			runResult <- kubeController.Run(2, ctx.Done())
		}()

		// Fail if the caches have not synchronized after 10 seconds. This prevents the controller from hanging forever.
		var err error
		select {
		case err = <-runResult:
		case <-time.After(10 * time.Second):
			err = errors.Errorf("timed out while waiting for informer caches to sync")
		}

		// If initError is non-nil, the kube resource client will panic
		if err != nil {
			f.initError = errors.Wrapf(err, "failed to start kuberenetes controller")
		}

		// Mark the factory as started
		f.started = true
	})
}

func (f *ResourceClientSharedInformerFactory) GetLister(namespace string, obj runtime.Object) (ResourceLister, error) {
	informer := f.getInformer(namespace, obj)
	if informer == nil {
		// TODO: improve error message
		return nil, errors.Errorf("no lister has been registered for ObjectKind %v", obj.GetObjectKind())
	}
	return &resourceLister{indexer: informer.GetIndexer()}, nil

}

// Retrieve the informer for the given resourceType/namespace pair from the factory's registry
func (f *ResourceClientSharedInformerFactory) getInformer(namespace string, obj runtime.Object) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informerType := reflect.TypeOf(obj)

	// Look up by resource type
	if forResourceType, exists := f.informers[informerType]; exists {
		return forResourceType[namespace]
	}
	return nil
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
