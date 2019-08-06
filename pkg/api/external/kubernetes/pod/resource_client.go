package pod

import (
	"sort"

	kubepod "github.com/solo-io/solo-kit/api/external/kubernetes/pod"
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

type podResourceClient struct {
	cache cache.KubeCoreCache
	Kube  kubernetes.Interface
	common.KubeCoreResourceClient
}

func newResourceClient(kube kubernetes.Interface, cache cache.KubeCoreCache) *podResourceClient {
	return &podResourceClient{
		cache: cache,
		Kube:  kube,
		KubeCoreResourceClient: common.KubeCoreResourceClient{
			ResourceType: &skkube.Pod{},
		},
	}
}

func NewPodClient(kube kubernetes.Interface, cache cache.KubeCoreCache) skkube.PodClient {
	resourceClient := newResourceClient(kube, cache)
	return skkube.NewPodClientWithBase(resourceClient)
}

func FromKubePod(pod *kubev1.Pod) *skkube.Pod {

	podCopy := pod.DeepCopy()
	kubePod := kubepod.Pod(*podCopy)
	resource := &skkube.Pod{
		Pod: kubePod,
	}

	return resource
}

func ToKubePod(resource resources.Resource) (*kubev1.Pod, error) {
	podResource, ok := resource.(*skkube.Pod)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to pod-only client", resources.Kind(resource))
	}

	pod := kubev1.Pod(podResource.Pod)

	return &pod, nil
}

var _ clients.ResourceClient = &podResourceClient{}

func (rc *podResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	podObj, err := rc.Kube.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading podObj from kubernetes")
	}
	resource := FromKubePod(podObj)

	if resource == nil {
		return nil, errors.Errorf("podObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *podResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	podObj, err := ToKubePod(resource)
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
		if _, err := rc.Kube.CoreV1().Pods(meta.Namespace).Update(podObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube podObj %v", podObj.Name)
		}
	} else {
		if _, err := rc.Kube.CoreV1().Pods(meta.Namespace).Create(podObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube podObj %v", podObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(podObj.Namespace, podObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *podResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.Kube.CoreV1().Pods(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting podObj %v", name)
	}
	return nil
}

func (rc *podResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	if rc.cache.NamespacedPodLister(namespace) == nil {
		return nil, errors.Errorf("namespaces is not watched")
	}
	podObjList, err := rc.cache.NamespacedPodLister(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing pods level")
	}
	var resourceList resources.ResourceList
	for _, podObj := range podObjList {
		resource := FromKubePod(podObj)

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

func (rc *podResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	return common.KubeResourceWatch(rc.cache, rc.List, namespace, opts)
}

func (rc *podResourceClient) exist(namespace, name string) bool {
	_, err := rc.Kube.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
