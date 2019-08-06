package namespace

import (
	"sort"

	"github.com/solo-io/go-utils/errors"
	kubenamespace "github.com/solo-io/solo-kit/api/external/kubernetes/namespace"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	skerrors "github.com/solo-io/solo-kit/pkg/errors"
	kubev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type namespaceResourceClient struct {
	cache cache.KubeCoreCache
	Kube  kubernetes.Interface
	common.KubeCoreResourceClient
}

func newResourceClient(kube kubernetes.Interface, cache cache.KubeCoreCache) *namespaceResourceClient {
	return &namespaceResourceClient{
		cache: cache,
		Kube:  kube,
		KubeCoreResourceClient: common.KubeCoreResourceClient{
			ResourceType: &skkube.KubeNamespace{},
		},
	}
}

func NewNamespaceClient(kube kubernetes.Interface, cache cache.KubeCoreCache) skkube.KubeNamespaceClient {
	resourceClient := newResourceClient(kube, cache)
	return skkube.NewKubeNamespaceClientWithBase(resourceClient)
}

func FromKubeNamespace(namespace *kubev1.Namespace) *skkube.KubeNamespace {

	namespaceCopy := namespace.DeepCopy()
	kubeNamespace := kubenamespace.KubeNamespace(*namespaceCopy)
	resource := &skkube.KubeNamespace{
		KubeNamespace: kubeNamespace,
	}

	return resource
}

func ToKubeNamespace(resource resources.Resource) (*kubev1.Namespace, error) {
	namespaceResource, ok := resource.(*skkube.KubeNamespace)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to namespace-only client", resources.Kind(resource))
	}

	namespace := kubev1.Namespace(namespaceResource.KubeNamespace)

	return &namespace, nil
}

func (rc *namespaceResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	namespaceObj, err := rc.Kube.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, skerrors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading namespaceObj from kubernetes")
	}
	resource := FromKubeNamespace(namespaceObj)

	if resource == nil {
		return nil, errors.Errorf("namespaceObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *namespaceResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	namespaceObj, err := ToKubeNamespace(resource)
	if err != nil {
		return nil, err
	}

	original, err := rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{
		Ctx: opts.Ctx,
	})
	if original != nil && err == nil {
		if !opts.OverwriteExisting {
			return nil, skerrors.NewExistErr(meta)
		}
		if meta.ResourceVersion != original.GetMetadata().ResourceVersion {
			return nil, skerrors.NewResourceVersionErr(meta.Namespace, meta.Name, meta.ResourceVersion, original.GetMetadata().ResourceVersion)
		}
		if _, err := rc.Kube.CoreV1().Namespaces().Update(namespaceObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube namespaceObj %v", namespaceObj.Name)
		}
	} else {
		if _, err := rc.Kube.CoreV1().Namespaces().Create(namespaceObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube namespaceObj %v", namespaceObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(namespaceObj.Namespace, namespaceObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *namespaceResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return skerrors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.Kube.CoreV1().Namespaces().Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting namespaceObj %v", name)
	}
	return nil
}

func (rc *namespaceResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	if rc.cache.NamespaceLister() == nil {
		return nil, errors.New("to list namespaces you must watch all namespaces")
	}

	namespaceObjList, err := rc.cache.NamespaceLister().List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing namespaces level")
	}
	var resourceList resources.ResourceList
	for _, namespaceObj := range namespaceObjList {
		resource := FromKubeNamespace(namespaceObj)

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

func (rc *namespaceResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	return common.KubeResourceWatch(rc.cache, rc.List, namespace, opts)
}

func (rc *namespaceResourceClient) exist(namespace, name string) bool {
	_, err := rc.Kube.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	return err == nil
}
