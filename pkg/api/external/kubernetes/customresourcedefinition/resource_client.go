package customresourcedefinition

import (
	"context"
	"sort"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/solo-io/solo-kit/api/external/kubernetes/customresourcedefinition"

	"github.com/solo-io/solo-kit/pkg/api/shared"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/errors"

	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

type customResourceDefinitionResourceClient struct {
	cache   KubeCustomResourceDefinitionCache
	apiExts apiexts.Interface
	common.KubeCoreResourceClient
}

func newResourceClient(apiExts apiexts.Interface, cache KubeCustomResourceDefinitionCache) *customResourceDefinitionResourceClient {
	return &customResourceDefinitionResourceClient{
		cache:   cache,
		apiExts: apiExts,
		KubeCoreResourceClient: common.KubeCoreResourceClient{
			ResourceType: &skkube.CustomResourceDefinition{},
		},
	}
}

func NewCustomResourceDefinitionClient(apiExts apiexts.Interface, cache KubeCustomResourceDefinitionCache) skkube.CustomResourceDefinitionClient {
	resourceClient := newResourceClient(apiExts, cache)
	return skkube.NewCustomResourceDefinitionClientWithBase(resourceClient)
}

func FromKubeCustomResourceDefinition(customResourceDefinition *v1.CustomResourceDefinition) *skkube.CustomResourceDefinition {

	customResourceDefinitionCopy := customResourceDefinition.DeepCopy()
	kubeCustomResourceDefinition := customresourcedefinition.CustomResourceDefinition(*customResourceDefinitionCopy)
	resource := &skkube.CustomResourceDefinition{
		CustomResourceDefinition: kubeCustomResourceDefinition,
	}

	return resource
}

func ToKubeCustomResourceDefinition(resource resources.Resource) (*v1.CustomResourceDefinition, error) {
	customResourceDefinitionResource, ok := resource.(*skkube.CustomResourceDefinition)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to customResourceDefinition-only client", resources.Kind(resource))
	}

	customResourceDefinition := v1.CustomResourceDefinition(customResourceDefinitionResource.CustomResourceDefinition)

	return &customResourceDefinition, nil
}

var _ clients.ResourceClient = &customResourceDefinitionResourceClient{}

func (rc *customResourceDefinitionResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	customResourceDefinitionObj, err := rc.apiExts.ApiextensionsV1().CustomResourceDefinitions().Get(opts.Ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading customResourceDefinitionObj from kubernetes")
	}
	resource := FromKubeCustomResourceDefinition(customResourceDefinitionObj)

	if resource == nil {
		return nil, errors.Errorf("customResourceDefinitionObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *customResourceDefinitionResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	customResourceDefinitionObj, err := ToKubeCustomResourceDefinition(resource)
	if err != nil {
		return nil, err
	}

	original, err := rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{
		Ctx: opts.Ctx,
	})
	if original != nil && err == nil {
		if !opts.OverwriteExisting {
			return nil, errors.NewExistErr(meta)
		}
		if meta.ResourceVersion != original.GetMetadata().ResourceVersion {
			return nil, errors.NewResourceVersionErr(meta.Namespace, meta.Name, meta.ResourceVersion, original.GetMetadata().ResourceVersion)
		}
		if _, err := rc.apiExts.ApiextensionsV1().CustomResourceDefinitions().Update(opts.Ctx, customResourceDefinitionObj, metav1.UpdateOptions{}); err != nil {
			return nil, errors.Wrapf(err, "updating kube customResourceDefinitionObj %v", customResourceDefinitionObj.Name)
		}
	} else {
		if _, err := rc.apiExts.ApiextensionsV1().CustomResourceDefinitions().Create(opts.Ctx, customResourceDefinitionObj, metav1.CreateOptions{}); err != nil {
			return nil, errors.Wrapf(err, "creating kube customResourceDefinitionObj %v", customResourceDefinitionObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(customResourceDefinitionObj.Namespace, customResourceDefinitionObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *customResourceDefinitionResourceClient) ApplyStatus(statusClient resources.StatusClient, inputResource resources.InputResource, opts clients.ApplyStatusOpts) (resources.Resource, error) {
	name := inputResource.GetMetadata().GetName()
	namespace := inputResource.GetMetadata().GetNamespace()
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	data, err := shared.GetJsonPatchData(opts.Ctx, inputResource)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting status json patch data")
	}

	customResourceDefinitionObj, err := rc.apiExts.ApiextensionsV1().CustomResourceDefinitions().Patch(opts.Ctx, name, types.JSONPatchType, data, metav1.PatchOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "patching customResourceDefinitionObj status from kubernetes")
	}
	resource := FromKubeCustomResourceDefinition(customResourceDefinitionObj)

	if resource == nil {
		return nil, errors.Errorf("customResourceDefinitionObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *customResourceDefinitionResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(opts.Ctx, namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.apiExts.ApiextensionsV1().CustomResourceDefinitions().Delete(opts.Ctx, name, metav1.DeleteOptions{}); err != nil {
		return errors.Wrapf(err, "deleting customResourceDefinitionObj %v", name)
	}
	return nil
}

func (rc *customResourceDefinitionResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	customResourceDefinitionObjList, err := rc.cache.CustomResourceDefinitionLister().List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing customResourceDefinitions level")
	}
	var resourceList resources.ResourceList
	for _, customResourceDefinitionObj := range customResourceDefinitionObjList {
		resource := FromKubeCustomResourceDefinition(customResourceDefinitionObj)

		if resource == nil {
			continue
		}
		resourceList = append(resourceList, resource)
	}

	sort.SliceStable(resourceList, func(i, j int) bool {
		return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
	})

	return resourceList, nil
}

func (rc *customResourceDefinitionResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	return common.KubeResourceWatch(rc.cache, rc.List, namespace, opts)
}

func (rc *customResourceDefinitionResourceClient) exist(ctx context.Context, namespace, name string) bool {
	_, err := rc.apiExts.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	return err == nil
}
