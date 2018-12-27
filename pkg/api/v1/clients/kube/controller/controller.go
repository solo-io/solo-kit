package controller

import (
	"fmt"
	"time"

	"github.com/solo-io/solo-kit/pkg/utils/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	name string

	kubeclientset kubernetes.Interface

	syncFuncs []cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// handler to call
	handler cache.ResourceEventHandler
}

// NewController returns a new controller
func NewController(
	controllerName string,
	kubeclient kubernetes.Interface,
	handler cache.ResourceEventHandler,
	informers ...cache.SharedInformer) *Controller {

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(log.Printf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})

	c := &Controller{
		name:      controllerName,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		recorder:  recorder,
		handler:   handler,
	}

	var hasSyncedFuncs []cache.InformerSynced
	for _, informer := range informers {
		hasSyncedFuncs = append(hasSyncedFuncs, informer.HasSynced)

		// Set up an event handler for when any informer's resources change
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.enqueueSync(added, nil, obj)
			},
			UpdateFunc: func(old, new interface{}) {
				c.enqueueSync(updated, old, new)
			},
			DeleteFunc: func(obj interface{}) {
				c.enqueueSync(deleted, nil, obj)
			},
		})
	}
	c.syncFuncs = hasSyncedFuncs

	return c
}

func (c *Controller) enqueueSync(t eventType, old, new interface{}) {
	e := &event{
		eventType: t,
		old:       old,
		new:       new,
	}
	// log the meta key for the obj
	// currently unused otherwise
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(e.new); err != nil {
		runtime.HandleError(err)
		return
	}
	// TODO: create multiple verbosity levels
	if false {
		log.Debugf("[%s] EVENT: %s: %s", c.name, e.eventType, key)
	}
	c.workqueue.AddRateLimited(e)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Debugf("Starting %v controller", c.name)

	// Wait for the caches to be synced before starting workers
	log.Debugf("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, []cache.InformerSynced(c.syncFuncs)...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Debugf("Starting workers")
	// Launch two workers to process resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Debugf("Started workers")
	<-stopCh
	log.Debugf("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var w *event
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if w, ok = obj.(*event); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected event type in workqueue but got %#v", obj))
			return nil
		}
		switch w.eventType {
		case added:
			c.handler.OnAdd(w.new)
		case updated:
			c.handler.OnUpdate(w.old, w.new)
		case deleted:
			c.handler.OnDelete(w.new)
		}

		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
	}

	return true
}
