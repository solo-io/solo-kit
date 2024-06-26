package kube

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/api/shared"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned"
	v1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/specutils"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	MCreates        = stats.Int64("kube/creates", "The number of creates", "1")
	CreateCountView = &view.View{
		Name:        "kube/creates-count",
		Measure:     MCreates,
		Description: "The number of create calls",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			KeyKind,
		},
	}
	MUpdates        = stats.Int64("kube/updates", "The number of updates", "1")
	UpdateCountView = &view.View{
		Name:        "kube/updates-count",
		Measure:     MUpdates,
		Description: "The number of update calls",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			KeyKind,
		},
	}

	MDeletes        = stats.Int64("kube/deletes", "The number of deletes", "1")
	DeleteCountView = &view.View{
		Name:        "kube/deletes-count",
		Measure:     MDeletes,
		Description: "The number of delete calls",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			KeyKind,
		},
	}

	KeyOpKind, _ = tag.NewKey("op")

	MInFlight       = stats.Int64("kube/req_in_flight", "The number of requests in flight", "1")
	InFlightSumView = &view.View{
		Name:        "kube/req-in-flight",
		Measure:     MInFlight,
		Description: "The number of requests in flight",
		Aggregation: view.Sum(),
		TagKeys: []tag.Key{
			KeyOpKind,
			KeyKind,
		},
	}

	MEvents         = stats.Int64("kube/events", "The number of events", "1")
	EventsCountView = &view.View{
		Name:        "kube/events-count",
		Measure:     MEvents,
		Description: "The number of events sent from kuberenets to us",
		Aggregation: view.Count(),
	}
)

func init() {
	view.Register(CreateCountView, UpdateCountView, DeleteCountView, InFlightSumView, EventsCountView)
}

// Kuberenetes specific write options.
// Allows modifing a resource just before it is written.
// This allows to make kubernetes specific changes, like adding an owner reference.
type KubeWriteOpts struct {
	PreWriteCallback func(r *v1.Resource)
}

func (*KubeWriteOpts) StorageWriteOptsTag() {}

var _ clients.StorageWriteOpts = new(KubeWriteOpts)

// lazy start in list & watch
// register informers in register
type ResourceClient struct {
	crd                       crd.Crd
	crdClientset              versioned.Interface
	resourceName              string
	resourceType              resources.InputResource
	sharedCache               SharedCache
	namespaceWhitelist        []string // Will contain at least metaV1.NamespaceAll ("")
	resyncPeriod              time.Duration
	resourceStatusUnmarshaler resources.StatusUnmarshaler
}

func NewResourceClient(
	crd crd.Crd,
	clientset versioned.Interface,
	sharedCache SharedCache,
	resourceType resources.InputResource,
	namespaceWhitelist []string,
	resyncPeriod time.Duration,
	resourceStatusUnmarshaler resources.StatusUnmarshaler,
) *ResourceClient {

	typeof := reflect.TypeOf(resourceType)
	resourceName := strings.Replace(typeof.String(), "*", "", -1)
	resourceName = strings.Replace(resourceName, ".", "", -1)

	return &ResourceClient{
		crd:                       crd,
		crdClientset:              clientset,
		resourceName:              resourceName,
		resourceType:              resourceType,
		sharedCache:               sharedCache,
		namespaceWhitelist:        namespaceWhitelist,
		resyncPeriod:              resyncPeriod,
		resourceStatusUnmarshaler: resourceStatusUnmarshaler,
	}
}

var _ clients.ResourceClient = &ResourceClient{}

func (rc *ResourceClient) Kind() string {
	return resources.Kind(rc.resourceType)
}

func (rc *ResourceClient) NewResource() resources.Resource {
	return resources.Clone(rc.resourceType)
}

// Registers the client with the shared cache. The cache will create a dedicated informer to list and
// watch resources of kind rc.Kind() in the namespaces given in rc.namespaceWhitelist.
func (rc *ResourceClient) Register() error {
	return rc.sharedCache.Register(rc)
}

func (rc *ResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	if err := rc.validateNamespace(namespace); err != nil {
		return nil, err
	}

	ctx := opts.Ctx

	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName), tag.Insert(KeyOpKind, "read")); err == nil {
		ctx = ctxWithTags
	}

	stats.Record(ctx, MInFlight.M(1))
	resourceCrd, err := rc.crdClientset.ResourcesV1().Resources(namespace).Get(ctx, name, metav1.GetOptions{})
	stats.Record(ctx, MInFlight.M(-1))
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading resource from kubernetes")
	}
	if !rc.matchesClientGVK(*resourceCrd) {
		return nil, errors.Errorf("cannot read %v resource with %v client", resourceCrd.GroupVersionKind().String(), rc.crd.GroupVersionKind().String())
	}
	resource, err := rc.convertCrdToResource(resourceCrd)
	if err != nil {
		return nil, errors.Wrapf(err, "converting output crd")
	}
	return resource, nil
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	if err := rc.validateNamespace(meta.Namespace); err != nil {
		return nil, err
	}

	// mutate and return clone
	clone := resources.Clone(resource).(resources.InputResource)
	clone.SetMetadata(meta)
	resourceCrd, err := rc.crd.KubeResource(clone)
	if err != nil {
		return nil, err
	}

	ctx := opts.Ctx
	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName), tag.Insert(KeyOpKind, "write")); err == nil {
		ctx = ctxWithTags
	}

	if opts.StorageWriteOpts != nil {
		if kubeOpts, ok := opts.StorageWriteOpts.(*KubeWriteOpts); ok {
			kubeOpts.PreWriteCallback(resourceCrd)
		}
	}

	if rc.exist(ctx, meta.Namespace, meta.Name) {
		if !opts.OverwriteExisting {
			return nil, errors.NewExistErr(meta)
		}
		stats.Record(ctx, MUpdates.M(1), MInFlight.M(1))
		defer stats.Record(ctx, MInFlight.M(-1))
		if _, updateErr := rc.crdClientset.ResourcesV1().Resources(meta.Namespace).Update(ctx, resourceCrd, metav1.UpdateOptions{}); updateErr != nil {

			original, err := rc.crdClientset.ResourcesV1().Resources(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
			if err == nil {
				if apierrors.IsConflict(updateErr) {
					return nil, errors.NewResourceVersionErr(meta.Namespace, meta.Name, "", meta.ResourceVersion)
				}
				return nil, errors.Wrapf(updateErr, "updating kube resource %v:%v (want %v)", resourceCrd.Name, resourceCrd.ResourceVersion, original.ResourceVersion)
			}

			if apierrors.IsConflict(updateErr) {
				return nil, errors.NewResourceVersionErr(meta.Namespace, meta.Name, original.ObjectMeta.ResourceVersion, meta.ResourceVersion)
			}
			return nil, errors.Wrapf(updateErr, "updating kube resource %v", resourceCrd.Name)
		}
	} else {
		stats.Record(ctx, MCreates.M(1), MInFlight.M(1))
		defer stats.Record(ctx, MInFlight.M(-1))
		if _, err := rc.crdClientset.ResourcesV1().Resources(meta.Namespace).Create(ctx, resourceCrd, metav1.CreateOptions{}); err != nil {
			if apierrors.IsAlreadyExists(err) {
				return nil, errors.NewExistErr(meta)
			}
			return nil, errors.Wrapf(err, "creating kube resource %v", resourceCrd.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {

	if err := rc.validateNamespace(namespace); err != nil {
		return err
	}

	opts = opts.WithDefaults()

	ctx := opts.Ctx

	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName), tag.Insert(KeyOpKind, "delete")); err == nil {
		ctx = ctxWithTags
	}
	stats.Record(ctx, MDeletes.M(1))

	if !rc.exist(ctx, namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr(namespace, name)
		}
		return nil
	}

	stats.Record(ctx, MInFlight.M(1))
	defer stats.Record(ctx, MInFlight.M(-1))
	if err := rc.crdClientset.ResourcesV1().Resources(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return errors.Wrapf(err, "deleting resource %v", name)
	}
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	if err := rc.validateNamespace(namespace); err != nil {
		return nil, err
	}

	// Will have no effect if the factory is already running
	rc.sharedCache.Start()

	lister, err := rc.sharedCache.GetLister(namespace, rc.crd.Version.Type)
	if err != nil {
		return nil, err
	}

	labelSelector, err := kubeutils.ToLabelSelector(opts)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing label selector")
	}

	allResources, err := lister.List(labelSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "listing resources in %v", namespace)
	}

	var listedResources []*v1.Resource
	if namespace != "" {
		for _, r := range allResources {
			if r.ObjectMeta.Namespace == namespace {
				listedResources = append(listedResources, r)
			}
		}
	} else {
		listedResources = allResources
	}

	var resourceList resources.ResourceList
	for _, resourceCrd := range listedResources {
		if !rc.matchesClientGVK(*resourceCrd) {
			continue
		}
		resource, err := rc.convertCrdToResource(resourceCrd)
		if err != nil {
			return nil, errors.Wrapf(err, "converting output crd")
		}
		resourceList = append(resourceList, resource)
	}

	sort.SliceStable(resourceList, func(i, j int) bool {
		return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
	})

	return resourceList, nil
}

func (rc *ResourceClient) ApplyStatus(statusClient resources.StatusClient, inputResource resources.InputResource, opts clients.ApplyStatusOpts) (resources.Resource, error) {
	name := inputResource.GetMetadata().GetName()
	namespace := inputResource.GetMetadata().GetNamespace()
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	if err := rc.validateNamespace(namespace); err != nil {
		return nil, err
	}

	ctx := opts.Ctx

	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName), tag.Insert(KeyOpKind, "patch")); err == nil {
		ctx = ctxWithTags
	}

	data, err := shared.GetJsonPatchData(ctx, inputResource)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting status json patch data")
	}

	stats.Record(ctx, MInFlight.M(1))
	resourceCrd, err := rc.crdClientset.ResourcesV1().Resources(namespace).Patch(ctx, name, types.JSONPatchType, data, metav1.PatchOptions{})
	stats.Record(ctx, MInFlight.M(-1))
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "patching resource status from kubernetes")
	}
	if !rc.matchesClientGVK(*resourceCrd) {
		return nil, errors.Errorf("cannot patch %v resource with %v client", resourceCrd.GroupVersionKind().String(), rc.crd.GroupVersionKind().String())
	}
	resource, err := rc.convertCrdToResource(resourceCrd)
	if err != nil {
		return nil, errors.Wrapf(err, "converting output crd")
	}
	return resource, nil
}

func (rc *ResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {

	if err := rc.validateNamespace(namespace); err != nil {
		return nil, nil, err
	}

	rc.sharedCache.Start()

	opts = opts.WithDefaults()
	resourcesChan := make(chan resources.ResourceList, 1)
	errs := make(chan error)
	ctx := opts.Ctx

	updateResourceList := func() {
		list, err := rc.List(namespace, clients.ListOpts{
			Ctx:                ctx,
			Selector:           opts.Selector,
			ExpressionSelector: opts.ExpressionSelector,
		})
		if err != nil {
			errs <- err
			return
		}
		select {
		case resourcesChan <- list:
		default:
		Drainloop:
			for {
				select {
				case <-resourcesChan:
				default:
					break Drainloop
				}
			}
			// this will not block as we drained the channel.
			resourcesChan <- list
		}
	}
	// watch should open up with an initial read
	cacheUpdated := rc.sharedCache.AddWatch(10)

	go func(watchedNamespace string) {
		defer rc.sharedCache.RemoveWatch(cacheUpdated)
		defer close(resourcesChan)
		defer close(errs)

		// Perform an initial list operation

		timer := time.NewTicker(time.Second)
		defer timer.Stop()

		// watch should open up with an initial read
		updateResourceList()
		update := false
		for {
			select {
			case resource := <-cacheUpdated:

				// Only notify watchers if the updated resource is in the watched
				// namespace and its kind matches the one of the resource clientz
				if matchesTargetNamespace(watchedNamespace, resource.ObjectMeta.Namespace) && rc.matchesClientGVK(resource) {
					update = true
				}
			case <-timer.C:
				if update {
					updateResourceList()
					update = false
				}
			case <-ctx.Done():
				return
			}
		}

	}(namespace)

	return resourcesChan, errs, nil
}

// Checks whether the group version kind of the given resource matches that of the client's underlying CRD:
func (rc *ResourceClient) matchesClientGVK(resource v1.Resource) bool {
	return resource.GroupVersionKind().String() == rc.crd.GroupVersionKind().String()
}

func (rc *ResourceClient) exist(ctx context.Context, namespace, name string) bool {

	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName), tag.Upsert(KeyOpKind, "get")); err == nil {
		ctx = ctxWithTags
	}

	stats.Record(ctx, MInFlight.M(1))
	defer stats.Record(ctx, MInFlight.M(-1))

	_, err := rc.crdClientset.ResourcesV1().Resources(namespace).Get(ctx, name, metav1.GetOptions{}) // TODO(yuval-k): check error for real
	return err == nil

}

func (rc *ResourceClient) convertCrdToResource(resourceCrd *v1.Resource) (resources.Resource, error) {
	resource := rc.NewResource()
	resource.SetMetadata(kubeutils.FromKubeMeta(resourceCrd.ObjectMeta, true))

	if customResource, ok := resource.(resources.CustomInputResource); ok {
		// Handle custom spec/status unmarshalling

		if resourceCrd.Spec != nil {
			if err := customResource.UnmarshalSpec(*resourceCrd.Spec); err != nil {
				return nil, errors.Wrapf(err, "unmarshalling crd spec on custom resource %v in namespace %v into %v",
					resourceCrd.Name, resourceCrd.Namespace, rc.resourceName)
			}
		}
		customResource.UnmarshalStatus(resourceCrd.Status, rc.resourceStatusUnmarshaler)

	} else {
		// Default unmarshalling

		if withStatus, ok := resource.(resources.InputResource); ok {
			rc.resourceStatusUnmarshaler.UnmarshalStatus(resourceCrd.Status, withStatus)
		}
		if resourceCrd.Spec != nil {
			if err := specutils.UnmarshalSpecMapToResource(*resourceCrd.Spec, resource); err != nil {
				return nil, errors.Wrapf(err, "reading crd spec on resource %v in namespace %v into %v", resourceCrd.Name, resourceCrd.Namespace, rc.resourceName)
			}
		}
	}
	return resource, nil
}

// Check whether the given namespace is in the whitelist or we allow all namespaces
func (rc *ResourceClient) validateNamespace(namespace string) error {
	if !stringutils.ContainsAny([]string{namespace, metav1.NamespaceAll}, rc.namespaceWhitelist) {
		return errors.Errorf("this client was not configured to access resources in the [%v] namespace. "+
			"Allowed namespaces are %v", namespace, rc.namespaceWhitelist)
	}
	return nil
}

func matchesTargetNamespace(targetNs, resourceNs string) bool {
	// "" == all namespaces are valid
	if targetNs == "" {
		return true
	}
	return targetNs == resourceNs
}
