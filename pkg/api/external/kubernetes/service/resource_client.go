package service

import (
	"sort"

	kubeservice "github.com/solo-io/solo-kit/api/external/kubernetes/service"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	kubev1 "k8s.io/api/core/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type serviceResourceClient struct {
	cache cache.KubeCoreCache
	Kube  kubernetes.Interface
	common.KubeCoreResourceClient
}

func newResourceClient(kube kubernetes.Interface, cache cache.KubeCoreCache) *serviceResourceClient {
	return &serviceResourceClient{
		cache: cache,
		Kube:  kube,
		KubeCoreResourceClient: common.KubeCoreResourceClient{
			ResourceType: &skkube.Service{},
		},
	}
}

func NewServiceClient(kube kubernetes.Interface, cache cache.KubeCoreCache) skkube.ServiceClient {
	resourceClient := newResourceClient(kube, cache)
	return skkube.NewServiceClientWithBase(resourceClient)
}

func FromKubeService(service *kubev1.Service) *skkube.Service {

	serviceCopy := service.DeepCopy()
	kubeService := kubeservice.Service(*serviceCopy)
	resource := &skkube.Service{
		Service: kubeService,
	}

	return resource
}

func ToKubeService(resource resources.Resource) (*kubev1.Service, error) {
	serviceResource, ok := resource.(*skkube.Service)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to service-only client", resources.Kind(resource))
	}

	service := kubev1.Service(serviceResource.Service)

	return &service, nil
}

var _ clients.ResourceClient = &serviceResourceClient{}

func (rc *serviceResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	serviceObj, err := rc.Kube.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading serviceObj from kubernetes")
	}
	resource := FromKubeService(serviceObj)

	if resource == nil {
		return nil, errors.Errorf("serviceObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *serviceResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	serviceObj, err := ToKubeService(resource)
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
		if _, err := rc.Kube.CoreV1().Services(meta.Namespace).Update(serviceObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube serviceObj %v", serviceObj.Name)
		}
	} else {
		if _, err := rc.Kube.CoreV1().Services(meta.Namespace).Create(serviceObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube serviceObj %v", serviceObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(serviceObj.Namespace, serviceObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *serviceResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.Kube.CoreV1().Services(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting serviceObj %v", name)
	}
	return nil
}

func (rc *serviceResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	if rc.cache.NamespacedServiceLister(namespace) == nil {
		return nil, errors.Errorf("namespaces is not watched")
	}
	serviceObjList, err := rc.cache.NamespacedServiceLister(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing services level")
	}
	var resourceList resources.ResourceList
	for _, serviceObj := range serviceObjList {
		resource := FromKubeService(serviceObj)

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

func (rc *serviceResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	return common.KubeResourceWatch(rc.cache, rc.List, namespace, opts)
}

func (rc *serviceResourceClient) exist(namespace, name string) bool {
	_, err := rc.Kube.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
