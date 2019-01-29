package configmap

import (
	"reflect"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	v1 "k8s.io/api/core/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const annotationKey = "resource_kind"

func (rc *ResourceClient) fromKubeConfigMap(configMap *v1.ConfigMap) (resources.Resource, error) {
	resource := rc.NewResource()
	// not our configMap
	// should be an error on a Read, ignored on a list
	if len(configMap.ObjectMeta.Annotations) == 0 || configMap.ObjectMeta.Annotations[annotationKey] != rc.Kind() {
		return nil, nil
	}
	// convert mapstruct to our object
	resourceMap, err := protoutils.MapStringStringToMapStringInterface(configMap.Data)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing configmap data as map[string]interface{}")
	}

	if err := protoutils.UnmarshalMap(resourceMap, resource); err != nil {
		return nil, errors.Wrapf(err, "reading configmap data into %v", rc.Kind())
	}
	resource.SetMetadata(kubeutils.FromKubeMeta(configMap.ObjectMeta))

	return resource, nil
}

func (rc *ResourceClient) toKubeConfigMap(resource resources.Resource) (*v1.ConfigMap, error) {
	resourceMap, err := protoutils.MarshalMap(resource)
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling resource as map")
	}
	resourceData, err := protoutils.MapStringInterfaceToMapStringString(resourceMap)
	if err != nil {
		return nil, errors.Wrapf(err, "internal err: converting resource map to map[string]string")
	}
	// metadata moves over to kube style
	delete(resourceData, "metadata")
	meta := kubeutils.ToKubeMeta(resource.GetMetadata())
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations[annotationKey] = rc.Kind()
	return &v1.ConfigMap{
		ObjectMeta: meta,
		Data:       resourceData,
	}, nil
}

type ResourceClient struct {
	apiexts      apiexts.Interface
	kube         kubernetes.Interface
	ownerLabel   string
	resourceName string
	resourceType resources.Resource
}

func NewResourceClient(kube kubernetes.Interface, resourceType resources.Resource) (*ResourceClient, error) {
	return &ResourceClient{
		kube:         kube,
		resourceName: reflect.TypeOf(resourceType).String(),
		resourceType: resourceType,
	}, nil
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
	namespace = clients.DefaultNamespaceIfEmpty(namespace)

	configMap, err := rc.kube.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading configMap from kubernetes")
	}
	resource, err := rc.fromKubeConfigMap(configMap)
	if err != nil {
		return nil, err
	}
	if resource == nil {
		return nil, errors.Errorf("configMap %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()
	meta.Namespace = clients.DefaultNamespaceIfEmpty(meta.Namespace)

	// mutate and return clone
	clone := proto.Clone(resource).(resources.Resource)
	clone.SetMetadata(meta)
	configMap, err := rc.toKubeConfigMap(resource)
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
		if _, err := rc.kube.CoreV1().ConfigMaps(configMap.Namespace).Update(configMap); err != nil {
			return nil, errors.Wrapf(err, "updating kube configMap %v", configMap.Name)
		}
	} else {
		if _, err := rc.kube.CoreV1().ConfigMaps(configMap.Namespace).Create(configMap); err != nil {
			return nil, errors.Wrapf(err, "creating kube configMap %v", configMap.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(configMap.Namespace, configMap.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr(namespace, name)
		}
		return nil
	}

	if err := rc.kube.CoreV1().ConfigMaps(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting configMap %v", name)
	}
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()
	namespace = clients.DefaultNamespaceIfEmpty(namespace)

	configMapList, err := rc.kube.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(opts.Selector).String(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "listing configMaps in %v", namespace)
	}
	var resourceList resources.ResourceList
	for _, configMap := range configMapList.Items {
		resource, err := rc.fromKubeConfigMap(&configMap)
		if err != nil {
			return nil, err
		}
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

func (rc *ResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()
	namespace = clients.DefaultNamespaceIfEmpty(namespace)
	watch, err := rc.kube.CoreV1().ConfigMaps(namespace).Watch(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(opts.Selector).String(),
	})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "initiating kube watch in %v", namespace)
	}
	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)
	updateResourceList := func() {
		list, err := rc.List(namespace, clients.ListOpts{
			Ctx:      opts.Ctx,
			Selector: opts.Selector,
		})
		if err != nil {
			errs <- err
			return
		}
		resourcesChan <- list
	}

	go func() {
		// watch should open up with an initial read
		updateResourceList()
		for {
			select {
			case <-time.After(opts.RefreshRate):
				updateResourceList()
			case event := <-watch.ResultChan():
				switch event.Type {
				case kubewatch.Error:
					errs <- errors.Errorf("error during watch: %v", event)
				default:
					updateResourceList()
				}
			case <-opts.Ctx.Done():
				watch.Stop()
				close(resourcesChan)
				close(errs)
				return
			}
		}
	}()

	return resourcesChan, errs, nil
}

func (rc *ResourceClient) exist(namespace, name string) bool {
	_, err := rc.kube.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
