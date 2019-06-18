package deployment

import (
	"sort"

	kubedeployment "github.com/solo-io/solo-kit/api/external/kubernetes/deployment"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type deploymentResourceClient struct {
	cache cache.KubeDeploymentCache
	Kube  kubernetes.Interface
	common.KubeCoreResourceClient
}

func newResourceClient(kube kubernetes.Interface, cache cache.KubeDeploymentCache) *deploymentResourceClient {
	return &deploymentResourceClient{
		cache: cache,
		Kube:  kube,
		KubeCoreResourceClient: common.KubeCoreResourceClient{
			ResourceType: &skkube.Deployment{},
		},
	}
}

func NewDeploymentClient(kube kubernetes.Interface, cache cache.KubeDeploymentCache) skkube.DeploymentClient {
	resourceClient := newResourceClient(kube, cache)
	return skkube.NewDeploymentClientWithBase(resourceClient)
}

func FromKubeDeployment(deployment *appsv1.Deployment) *skkube.Deployment {

	deploymentCopy := deployment.DeepCopy()
	kubeDeployment := kubedeployment.Deployment(*deploymentCopy)
	resource := &skkube.Deployment{
		Deployment: kubeDeployment,
	}

	return resource
}

func ToKubeDeployment(resource resources.Resource) (*appsv1.Deployment, error) {
	deploymentResource, ok := resource.(*skkube.Deployment)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to deployment-only client", resources.Kind(resource))
	}

	deployment := appsv1.Deployment(deploymentResource.Deployment)

	return &deployment, nil
}

var _ clients.ResourceClient = &deploymentResourceClient{}

func (rc *deploymentResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	deploymentObj, err := rc.Kube.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading deploymentObj from kubernetes")
	}
	resource := FromKubeDeployment(deploymentObj)

	if resource == nil {
		return nil, errors.Errorf("deploymentObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *deploymentResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	deploymentObj, err := ToKubeDeployment(resource)
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
		if _, err := rc.Kube.AppsV1().Deployments(meta.Namespace).Update(deploymentObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube deploymentObj %v", deploymentObj.Name)
		}
	} else {
		if _, err := rc.Kube.AppsV1().Deployments(meta.Namespace).Create(deploymentObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube deploymentObj %v", deploymentObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(deploymentObj.Namespace, deploymentObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *deploymentResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.Kube.AppsV1().Deployments(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting deploymentObj %v", name)
	}
	return nil
}

func (rc *deploymentResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	deploymentObjList, err := rc.cache.DeploymentLister().Deployments(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing deployments level")
	}
	var resourceList resources.ResourceList
	for _, deploymentObj := range deploymentObjList {
		resource := FromKubeDeployment(deploymentObj)

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

func (rc *deploymentResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	return common.KubeResourceWatch(rc.cache, rc.List, namespace, opts)
}

func (rc *deploymentResourceClient) exist(namespace, name string) bool {
	_, err := rc.Kube.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
