package clients

import (
	"context"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

//go:generate mockgen -destination=./mocks/client_interface.go -source client_interface.go -package mocks

const DefaultNamespace = "default"

var DefaultRefreshRate = time.Second * 30

func DefaultNamespaceIfEmpty(namespace string) string {
	if namespace == "" {
		return DefaultNamespace
	}
	return namespace
}

type ResourceWatch func(ctx context.Context) (<-chan resources.ResourceList, <-chan error, error)

type ResourceWatcher interface {
	Watch(namespace string, opts WatchOpts) (<-chan resources.ResourceList, <-chan error, error)
}

type ResourceClient interface {
	Kind() string
	NewResource() resources.Resource
	// Deprecated: implemented only by the kubernetes resource client. Will be removed from the interface.
	Register() error
	Read(namespace, name string, opts ReadOpts) (resources.Resource, error)
	Write(resource resources.Resource, opts WriteOpts) (resources.Resource, error)
	Delete(namespace, name string, opts DeleteOpts) error
	List(namespace string, opts ListOpts) (resources.ResourceList, error)
	ResourceWatcher
}

type ResourceClients map[string]ResourceClient

func (r ResourceClients) Add(rcs ...ResourceClient) {
	for _, rc := range rcs {
		r[rc.Kind()] = rc
	}
}

func (r ResourceClients) ForResource(resource resources.Resource) (ResourceClient, error) {
	return r.ForKind(resources.Kind(resource))
}

func (r ResourceClients) ForKind(kind string) (ResourceClient, error) {
	rc, ok := r[kind]
	if !ok {
		return nil, errors.Errorf("no resource client registered for kind %v", kind)
	}
	return rc, nil
}

type ReadOpts struct {
	Ctx     context.Context
	Cluster string
}

func (o ReadOpts) WithDefaults() ReadOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	return o
}

type StorageWriteOpts interface {
	StorageWriteOptsTag()
}

type WriteOpts struct {
	Ctx               context.Context
	OverwriteExisting bool

	// Implementation dependant write opts
	StorageWriteOpts StorageWriteOpts
}

func (o WriteOpts) WithDefaults() WriteOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	return o
}

type DeleteOpts struct {
	Ctx            context.Context
	IgnoreNotExist bool
	Cluster        string
}

func (o DeleteOpts) WithDefaults() DeleteOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	return o
}

type ListOpts struct {
	Ctx     context.Context
	Cluster string

	// Equality-based label requirements
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#equality-based-requirement
	// Equality-based requirements allow filtering by label keys and values.
	// Matching objects must satisfy all of the specified label constraints,
	// though they may have additional labels as well.
	// Example:
	//	{product: edge} would return all objects with a label key equal to
	//	product and label value equal to edge
	// If both ExpressionSelector and Selector are defined, ExpressionSelector is preferred
	Selector map[string]string
	// Set-based label requirements
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#set-based-requirement
	// Set-based label requirements allow filtering keys according to a set of values.
	// Three kinds of operators are supported: in,notin and exists (only the key identifier).
	// Set-based requirements can be mixed with equality-based requirements.
	// Example:
	//	"product in (edge, mesh),version=v1" would return all objects that match ALL of the following:
	//	 (1) the label key equal to product and value equal to edge or mesh
	//	 (2) the label key equal to version and the value equal to v1
	// If both ExpressionSelector and Selector are defined, ExpressionSelector is preferred
	ExpressionSelector string
	// TODO-JAKE add in Field Selectors
	FieldSelectors string
}

func (o ListOpts) WithDefaults() ListOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	return o
}

// TODO-JAKE do we want to combine the WatchOpts, ListOpts, and ResourceNamespaceOpts???

// RefreshRate is currently ignored by the Kubernetes ResourceClient implementation.
// To achieve a similar behavior you can use the KubeResourceClientFactory.ResyncPeriod field. The difference is that it
// will apply to all the watches started by clients built with the factory.
type WatchOpts struct {
	Ctx context.Context

	// Equality-based label requirements
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#equality-based-requirement
	// Equality-based requirements allow filtering by label keys and values.
	// Matching objects must satisfy all of the specified label constraints,
	// though they may have additional labels as well.
	// Example:
	//	{product: edge} would return all objects with a label key equal to
	//	product and label value equal to edge
	// If both ExpressionSelector and Selector are defined, ExpressionSelector is preferred
	Selector map[string]string
	// Set-based label requirements
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#set-based-requirement
	// Set-based label requirements allow filtering keys according to a set of values.
	// Three kinds of operators are supported: in,notin and exists (only the key identifier).
	// Set-based requirements can be mixed with equality-based requirements.
	// Example:
	//	"product in (edge, mesh),version=v1" would return all objects that match ALL of the following:
	//	 (1) the label key equal to product and value equal to edge or mesh
	//	 (2) the label key equal to version and the value equal to v1
	// If both ExpressionSelector and Selector are defined, ExpressionSelector is preferred
	ExpressionSelector string
	// JAKE-TODO
	FieldSelectors string
	RefreshRate    time.Duration
	// Cluster is ignored by aggregated watches, but is respected by multi cluster clients.
	Cluster string
}

func (o WatchOpts) WithDefaults() WatchOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	if o.RefreshRate == 0 {
		o.RefreshRate = DefaultRefreshRate
	}
	return o
}

func TranslateWatchOptsIntoListOpts(wopts WatchOpts) ListOpts {
	clopts := ListOpts{Ctx: wopts.Ctx, FieldSelectors: wopts.FieldSelectors, ExpressionSelector: wopts.ExpressionSelector, Selector: wopts.Selector}
	return clopts
}

// TODO-JAKE maybe they should be the same type of options?
// TranslateResourceNamespaceListToListOptions translates the resource namespace list options to List Options
func TranslateResourceNamespaceListToListOptions(lopts resources.ResourceNamespaceListOptions) ListOpts {
	clopts := ListOpts{Ctx: lopts.Ctx, ExpressionSelector: lopts.ExpressionSelector}
	return clopts
}

func TranslateResourceNamespaceListToWatchOptions(wopts resources.ResourceNamespaceWatchOptions) WatchOpts {
	clopts := WatchOpts{Ctx: wopts.Ctx, ExpressionSelector: wopts.ExpressionSelector}
	return clopts
}
