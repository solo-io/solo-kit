package memory

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/shared"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"k8s.io/apimachinery/pkg/labels"
)

const separator = ";"

type InMemoryResourceCache interface {
	Get(key string) (resources.Resource, bool)
	Delete(key string)
	Set(key string, resource resources.Resource)
	List(prefix string) resources.ResourceList
	Subscribe() chan struct{}
	Unsubscribe(subscription chan struct{})
}

type inMemoryResourceCache struct {
	store       map[string]resources.Resource
	lock        sync.RWMutex
	subscribers []chan struct{}
}

func (c *inMemoryResourceCache) signalUpdate() {
	for _, subscription := range c.subscribers {
		select {
		case subscription <- struct{}{}:
		default:
			// already in signaled state, nothing to do
		}
	}
}

func (c *inMemoryResourceCache) Get(key string) (resources.Resource, bool) {
	c.lock.RLock()
	resource, ok := c.store[key]
	c.lock.RUnlock()
	return resource, ok
}

func (c *inMemoryResourceCache) Delete(key string) {
	c.lock.Lock()
	delete(c.store, key)
	c.signalUpdate()
	c.lock.Unlock()
}

func (c *inMemoryResourceCache) Set(key string, resource resources.Resource) {
	c.lock.Lock()
	c.store[key] = resource
	c.signalUpdate()
	c.lock.Unlock()
}

func (c *inMemoryResourceCache) List(prefix string) resources.ResourceList {
	c.lock.RLock()
	defer c.lock.RUnlock()
	ress := make(resources.ResourceList, 0, len(c.store))
	for key, resource := range c.store {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		ress = append(ress, resource)
	}
	return ress
}

func (c *inMemoryResourceCache) Subscribe() chan struct{} {
	subscription := make(chan struct{}, 1)
	c.lock.Lock()
	c.subscribers = append(c.subscribers, subscription)
	c.lock.Unlock()
	return subscription
}

func (c *inMemoryResourceCache) Unsubscribe(subscription chan struct{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for i, sub := range c.subscribers {
		if sub == subscription {
			c.subscribers = append(c.subscribers[:i], c.subscribers[i+1:]...)
			return
		}
	}
}

func NewInMemoryResourceCache() InMemoryResourceCache {
	return &inMemoryResourceCache{
		store: make(map[string]resources.Resource),
	}
}

type ResourceClient struct {
	resourceType resources.Resource
	cache        InMemoryResourceCache
}

func NewResourceClient(cache InMemoryResourceCache, resourceType resources.Resource) *ResourceClient {
	return &ResourceClient{
		cache:        cache,
		resourceType: resourceType,
	}
}

var _ clients.ResourceClient = &ResourceClient{}

func (rc *ResourceClient) Kind() string {
	return resources.Kind(rc.resourceType)
}

func (rc *ResourceClient) NewResource() resources.Resource {
	return resources.Clone(rc.resourceType)
}

func (rc *ResourceClient) Register() error {
	return nil
}

func (rc *ResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()
	resource, ok := rc.cache.Get(rc.key(namespace, name))
	if !ok {
		return nil, errors.NewNotExistErr(namespace, name)
	}

	// avoid data races
	clone := resources.Clone(resource)
	return clone, nil
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}

	key := rc.key(resource.GetMetadata().GetNamespace(), resource.GetMetadata().GetName())

	original, err := rc.Read(
		resource.GetMetadata().GetNamespace(),
		resource.GetMetadata().GetName(),
		clients.ReadOpts{},
	)
	if original != nil && err == nil {
		if !opts.OverwriteExisting {
			return nil, errors.NewExistErr(resource.GetMetadata())
		}
		if resource.GetMetadata().GetResourceVersion() != original.GetMetadata().GetResourceVersion() {
			return nil, errors.NewResourceVersionErr(
				resource.GetMetadata().GetNamespace(),
				resource.GetMetadata().GetName(),
				resource.GetMetadata().GetResourceVersion(),
				original.GetMetadata().GetResourceVersion(),
			)
		}
	}

	// mutate and return clone
	newResource := resources.Clone(resource)
	// initialize or increment resource version
	newResource.GetMetadata().ResourceVersion = newOrIncrementResourceVer(resource.GetMetadata().GetResourceVersion())

	rc.cache.Set(key, newResource)

	return resources.Clone(newResource), nil
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	key := rc.key(namespace, name)
	_, ok := rc.cache.Get(key)
	if !ok {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr(namespace, name)
		}
		return nil
	}

	rc.cache.Delete(key)
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()
	cachedResources := rc.cache.List(rc.Prefix(namespace))

	var labelSelector labels.Selector
	var err error
	if kubeutils.HasSelector(opts) {
		labelSelector, err = kubeutils.ToLabelSelector(opts)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing label selector")
		}
	}

	var resourceList resources.ResourceList
	for _, resource := range cachedResources {
		if labelSelector == nil || labelSelector.Matches(labels.Set(resource.GetMetadata().Labels)) {
			clone := resources.Clone(resource)
			resourceList = append(resourceList, clone)
		}
	}

	sort.Stable(resourceList)

	return resourceList, nil
}

func (rc *ResourceClient) ApplyStatus(statusClient resources.StatusClient, inputResource resources.InputResource, opts clients.ApplyStatusOpts) (resources.Resource, error) {
	return shared.ApplyStatus(rc, statusClient, inputResource, opts)
}

func (rc *ResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()
	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)
	updateResourceList := func() {
		list, err := rc.List(namespace, clients.ListOpts{
			Ctx:                opts.Ctx,
			Selector:           opts.Selector,
			ExpressionSelector: opts.ExpressionSelector,
		})
		if err != nil {
			errs <- err
			return
		}
		resourcesChan <- list.FilterByKind(rc.Kind())
	}

	subscription := rc.cache.Subscribe()
	go func() {
		updateResourceList()
		for {
			select {
			case <-time.After(opts.RefreshRate):
				updateResourceList()
			case <-subscription:
				updateResourceList()
			case <-opts.Ctx.Done():
				rc.cache.Unsubscribe(subscription)
				close(resourcesChan)
				close(errs)
				return
			}
		}
	}()

	return resourcesChan, errs, nil
}

func (rc *ResourceClient) Prefix(namespace string) string {
	return rc.Kind() + separator + namespace
}

func (rc *ResourceClient) key(namespace, name string) string {
	return rc.Prefix(namespace) + separator + name
}

// util methods
func newOrIncrementResourceVer(resourceVersion string) string {
	curr, err := strconv.Atoi(resourceVersion)
	if err != nil {
		curr = 1
	}
	return fmt.Sprintf("%v", curr+1)
}
