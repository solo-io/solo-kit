package factory

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/configmap"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/consul"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/file"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/vault"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type NewResourceClientParams struct {
	ResourceType resources.Resource
	Token        string
}

// TODO(ilackarms): more opts validation
func newResourceClient(ctx context.Context, factory ResourceClientFactory, params NewResourceClientParams) (clients.ResourceClient, error) {
	resourceType := params.ResourceType
	switch opts := factory.(type) {
	case *KubeResourceClientFactory:
		kubeCfg := opts.Cfg
		if kubeCfg == nil {
			return nil, errors.Errorf("must provide a rest.Config for the kube resource client")
		}
		if params.Token != "" {
			kubeCfg.BearerToken = strings.TrimPrefix(params.Token, "Bearer ")
		}
		inputResource, ok := params.ResourceType.(resources.InputResource)
		if !ok {
			return nil, errors.Errorf("the kubernetes crd client can only be used for input resources, received type %v", resources.Kind(resourceType))
		}
		if opts.Crd.Version.Type == nil {
			return nil, errors.Errorf("must provide a crd for the kube resource client")
		}
		if opts.SharedCache == nil {
			return nil, errors.Errorf("must provide a shared cache for the kube resource client")
		}

		// Validate namespace whitelist:
		// 1. If no namespace list was provided, default to all namespaces
		// 2. Error if namespace list contains the empty string plus other values
		namespaceWhitelist := opts.NamespaceWhitelist
		if len(namespaceWhitelist) == 0 {
			namespaceWhitelist = []string{metav1.NamespaceAll}
		}
		if len(namespaceWhitelist) > 1 && stringutils.ContainsString(metav1.NamespaceAll, namespaceWhitelist) {
			return nil, fmt.Errorf("the kube resource client namespace list must contain either "+
				"the empty string (all namespaces) or multiple non-empty strings. Found both: %v", namespaceWhitelist)
		}

		// Create clientset for crd
		crdClient, err := versioned.NewForConfig(kubeCfg, opts.Crd)
		if err != nil {
			return nil, errors.Wrapf(err, "creating crd client")
		}

		for _, ns := range namespaceWhitelist {
			// verify the crd is registered and we have List permission
			// for the given namespaces
			if _, err := crdClient.ResourcesV1().Resources(ns).List(ctx, metav1.ListOptions{}); err != nil {
				return nil, errors.Wrapf(err, "list check failed")
			}
		}

		client := kube.NewResourceClient(
			opts.Crd,
			crdClient,
			opts.SharedCache,
			inputResource,
			namespaceWhitelist,
			opts.ResyncPeriod,
		)
		return wrapper.NewClusterResourceClient(client, opts.Cluster), nil

	case *ConsulResourceClientFactory:
		versionedResource, ok := params.ResourceType.(resources.VersionedResource)
		if !ok {
			return nil, errors.Errorf("the consul storage client can only be used for resources which implement the resources.VersionedResource interface resources, received type %v", resources.Kind(resourceType))
		}
		return consul.NewResourceClient(opts.Consul, opts.RootKey, versionedResource), nil
	case *FileResourceClientFactory:
		return file.NewResourceClient(opts.RootDir, resourceType), nil
	case *MemoryResourceClientFactory:
		return memory.NewResourceClient(opts.Cache, resourceType), nil
	case *KubeConfigMapClientFactory:
		if opts.Cache == nil {
			return nil, errors.Errorf("invalid opts, configmap client requires a kube core cache")
		}
		if opts.CustomConverter != nil {
			return configmap.NewResourceClientWithConverter(opts.Clientset, resourceType, opts.Cache, opts.CustomConverter)
		}
		client, err := configmap.NewResourceClient(opts.Clientset, resourceType, opts.Cache, opts.PlainConfigmaps)
		if err != nil {
			return nil, err
		}
		return wrapper.NewClusterResourceClient(client, opts.Cluster), nil
	case *KubeSecretClientFactory:
		if opts.Cache == nil {
			return nil, errors.Errorf("invalid opts, secret client requires a kube core cache")
		}
		if opts.SecretConverter != nil {
			return kubesecret.NewResourceClientWithSecretConverter(opts.Clientset, resourceType, opts.Cache, opts.SecretConverter)
		}
		client, err := kubesecret.NewResourceClient(opts.Clientset, resourceType, opts.PlainSecrets, opts.Cache)
		if err != nil {
			return nil, err
		}
		return wrapper.NewClusterResourceClient(client, opts.Cluster), nil
	case *VaultSecretClientFactory:
		versionedResource, ok := params.ResourceType.(resources.VersionedResource)
		if !ok {
			return nil, errors.Errorf("the vault storage client can only be used for resources which implement the resources.VersionedResource interface resources, received type %v", resources.Kind(resourceType))
		}
		return vault.NewResourceClient(opts.Vault, opts.RootKey, versionedResource), nil
	}
	panic("unsupported type " + reflect.TypeOf(factory).Name())
}

// https://golang.org/doc/faq#generics
type ResourceClientFactory interface {
	NewResourceClient(ctx context.Context, params NewResourceClientParams) (clients.ResourceClient, error)
}

// Clients built with this factory will be able to access only resources the given namespace list. If no value is provided,
// clients will be able to access resources in all namespaces.
type KubeResourceClientFactory struct {
	Crd                crd.Crd
	Cfg                *rest.Config
	SharedCache        kube.SharedCache
	NamespaceWhitelist []string
	ResyncPeriod       time.Duration
	// the cluster that these resources belong to
	// all resources written and read by the resource client
	// will be marked with this cluster
	Cluster string
}

func (f *KubeResourceClientFactory) NewResourceClient(ctx context.Context, params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(ctx, f, params)
}

type ConsulResourceClientFactory struct {
	Consul  *api.Client
	RootKey string
}

func (f *ConsulResourceClientFactory) NewResourceClient(ctx context.Context, params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(ctx, f, params)
}

type FileResourceClientFactory struct {
	RootDir string
}

func (f *FileResourceClientFactory) NewResourceClient(ctx context.Context, params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(ctx, f, params)
}

type MemoryResourceClientFactory struct {
	Cache memory.InMemoryResourceCache
}

func (f *MemoryResourceClientFactory) NewResourceClient(ctx context.Context, params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(ctx, f, params)
}

type KubeConfigMapClientFactory struct {
	Clientset kubernetes.Interface
	Cache     cache.KubeCoreCache
	// set this  to true if resource fields are all strings
	// resources will be stored as plain kubernetes configmaps without serializing/deserializing as objects
	PlainConfigmaps bool
	// a custom handler to define how configmaps are serialized/deserialized out of resources
	// if set, Plain is ignored
	CustomConverter configmap.ConfigMapConverter
	// the cluster that these resources belong to
	// all resources written and read by the resource client
	// will be marked with this cluster
	Cluster string
}

func (f *KubeConfigMapClientFactory) NewResourceClient(ctx context.Context, params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(ctx, f, params)
}

type KubeSecretClientFactory struct {
	Clientset kubernetes.Interface
	// set this  to true if resource fields are all strings
	// resources will be stored as plain kubernetes secrets without serializing/deserializing as objects
	PlainSecrets    bool
	SecretConverter kubesecret.SecretConverter
	Cache           cache.KubeCoreCache
	// the cluster that these resources belong to
	// all resources written and read by the resource client
	// will be marked with this cluster
	Cluster string
}

func (f *KubeSecretClientFactory) NewResourceClient(ctx context.Context, params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(ctx, f, params)
}

type VaultSecretClientFactory struct {
	Vault   *vaultapi.Client
	RootKey string
}

func (f *VaultSecretClientFactory) NewResourceClient(ctx context.Context, params NewResourceClientParams) (clients.ResourceClient, error) {
	return newResourceClient(ctx, f, params)
}
