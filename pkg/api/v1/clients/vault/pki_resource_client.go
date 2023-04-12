package vault

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

/**
Open Questions:
- Namespace concept is not available in Vault. How do production users separate certs or identify groupings? Will we just list all?
- How does rotation work? Do we need to worry about that? I imagine we don't, as we will just reference a cert by ref
- Secrets in Gloo have a name/namespace concept. We need some clear mapping of cert to name/namespace

TODO:
- Query Certificate API instead of returning empty list
- Determine how PrivateKeys will be pulled and if they are necessary
- Standup integration tests
*/

var _ clients.ResourceClient = new(PkiResourceClient)

var (
	readOnlyError       = errors.New("PKI ResourceClient is read-only")
	notImplementedError = errors.New("PKI ResourceClient is a WIP and this feature is not yet implemented")
)

type PkiResourceClient struct {
	resourceType resources.VersionedResource

	client          *api.Client
	secretConverter SecretConverter
}

type PKIResourceClientOptions struct {
	Client          *api.Client
	SecretConverter SecretConverter
}

func NewPkiResourceClient(resourceType resources.VersionedResource, options PKIResourceClientOptions) *PkiResourceClient {
	return &PkiResourceClient{
		resourceType:    resourceType,
		client:          options.Client,
		secretConverter: options.SecretConverter,
	}
}

func (p PkiResourceClient) Kind() string {
	return resources.Kind(p.resourceType)
}

func (p PkiResourceClient) NewResource() resources.Resource {
	return resources.Clone(p.resourceType)
}

// Register is a no-op.
// Deprecated: As outlined in the ResourceClient interface,
// Register is only necessary for the kubernetes resource client
func (p PkiResourceClient) Register() error {
	return nil
}

func (p PkiResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	panic(notImplementedError)
}

func (p PkiResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	// 1. Extract the secrets from the Vault store
	vaultSecretList, err := p.listSecrets(opts.Ctx)
	if err != nil {
		return nil, err
	}

	// 2. Convert the secrets to the Gloo resource format
	resourceList, conversionErr := p.convertSecrets(opts.Ctx, vaultSecretList)
	if conversionErr != nil {
		return nil, conversionErr
	}

	// 3. Sort the resources for idempotence
	sort.SliceStable(resourceList, func(i, j int) bool {
		return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
	})
	return resourceList, nil
}

func (p PkiResourceClient) listSecrets(ctx context.Context) (SecretList, error) {
	return SecretList{}, nil
}

func (p PkiResourceClient) convertSecrets(ctx context.Context, vaultSecretList SecretList) (resources.ResourceList, error) {
	var resourceList resources.ResourceList
	for _, vaultSecret := range vaultSecretList {
		resource, conversionErr := p.secretConverter.FromSecret(ctx, vaultSecret)
		if conversionErr != nil {
			switch conversionErr.(type) {
			case *UnrecoverableConversionError:
				// This should rarely (if ever) be used
				// Ideally invalid secrets do not half execution, and instead are processed
				return nil, conversionErr
			default:
				contextutils.LoggerFrom(ctx).Warnf("Failed to convert VaultSecret to GlooSecret: %v", conversionErr)
				continue
			}
		}
		resourceList = append(resourceList, resource)
	}

	return resourceList, nil
}

func (p PkiResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()
	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)

	listOpts := clients.ListOpts{
		Ctx: opts.Ctx,

		// These advanced selectors are available to resources which support label selection
		// This is not supported for Vault
		Selector:           opts.Selector,
		ExpressionSelector: opts.ExpressionSelector,
	}

	go func() {
		// watch should open up with an initial read
		initialResourceList, initialResourceListErr := p.List(namespace, listOpts)
		if initialResourceListErr != nil {
			errs <- initialResourceListErr
			return
		}
		resourcesChan <- initialResourceList
		for {
			select {
			case <-time.After(opts.RefreshRate):
				resourceList, resourceListErr := p.List(namespace, listOpts)
				if resourceListErr != nil {
					errs <- resourceListErr
				}
				resourcesChan <- resourceList
			case <-opts.Ctx.Done():
				close(resourcesChan)
				close(errs)
				return
			}
		}
	}()

	return resourcesChan, errs, nil
}

func (p PkiResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	panic(readOnlyError)
}

func (p PkiResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	panic(readOnlyError)
}

func (p PkiResourceClient) ApplyStatus(statusClient resources.StatusClient, inputResource resources.InputResource, opts clients.ApplyStatusOpts) (resources.Resource, error) {
	panic(readOnlyError)
}
