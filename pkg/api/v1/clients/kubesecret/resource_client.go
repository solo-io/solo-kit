package kubesecret

import (
	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	"k8s.io/api/core/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"sort"
)

const annotationKey = "resource_kind"

func (rc *ResourceClient) fromKubeSecret(secret *v1.Secret) (resources.Resource, error) {
	resource := rc.NewResource()
	// not our secret
	// should be an error on a Read, ignored on a list
	if len(secret.ObjectMeta.Annotations) == 0 || secret.ObjectMeta.Annotations[annotationKey] != rc.Kind() {
		return nil, nil
	}
	// convert mapstruct to our object
	resourceMap, err := protoutils.MapStringStringToMapStringInterface(toStringStringMap(secret.Data))
	if err != nil {
		return nil, errors.Wrapf(err, "parsing secret data as map[string]interface{}")
	}

	if err := protoutils.UnmarshalMap(resourceMap, resource); err != nil {
		return nil, errors.Wrapf(err, "reading secret data into %v", rc.Kind())
	}
	resource.SetMetadata(kubeutils.FromKubeMeta(secret.ObjectMeta))
	return resource, nil
}

func toStringStringMap(input map[string][]byte) map[string]string {
	output := make(map[string]string)
	for k, v := range input {
		output[k] = string(v)
	}
	return output
}

func fromStringStringMap(input map[string]string) map[string][]byte {
	output := make(map[string][]byte)
	for k, v := range input {
		output[k] = []byte(v)
	}
	return output
}

func (rc *ResourceClient) toKubeSecret(resource resources.Resource) (*v1.Secret, error) {
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
	return &v1.Secret{
		ObjectMeta: meta,
		Data:       fromStringStringMap(resourceData),
	}, nil
}

func (rc *ResourceClient) fromPlainKubeSecret(secret *v1.Secret) (resources.Resource, error) {
	resource := rc.NewResource()
	// not our secret
	// should be an error on a Read, ignored on a list
	if len(secret.ObjectMeta.Annotations) == 0 || secret.ObjectMeta.Annotations[annotationKey] != rc.Kind() {
		return nil, nil
	}
	// only works for string fields
	resourceMap := make(map[string]interface{})
	for k, v := range secret.Data {
		resourceMap[k] = string(v)
	}
	if err := protoutils.UnmarshalMap(resourceMap, resource); err != nil {
		return nil, errors.Wrapf(err, "reading secret data into %v", rc.Kind())
	}
	resource.SetMetadata(kubeutils.FromKubeMeta(secret.ObjectMeta))
	return resource, nil
}

func (rc *ResourceClient) toPlainKubeSecret(resource resources.Resource) (*v1.Secret, error) {
	resourceMap, err := protoutils.MarshalMapEmitZeroValues(resource)
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling resource as map")
	}
	kubeSecretData := make(map[string][]byte)
	for k, v := range resourceMap {
		switch val := v.(type) {
		case string:
			kubeSecretData[k] = []byte(val)
		default:
			// TODO: handle other field types; for now the caller
			// must know this resource client only supports map[string]string style objects
		}
	}

	meta := kubeutils.ToKubeMeta(resource.GetMetadata())
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations[annotationKey] = rc.Kind()
	return &v1.Secret{
		ObjectMeta: meta,
		Data:       kubeSecretData,
	}, nil
}

type ResourceClient struct {
	apiexts      apiexts.Interface
	kube         kubernetes.Interface
	ownerLabel   string
	resourceName string
	resourceType resources.Resource
	kubeCache    cache.KubeCoreCache
	// should we marshal/unmarshal these secrets assuming their structure is map[string]string ?
	plainSecrets bool
}

func NewResourceClient(kube kubernetes.Interface, resourceType resources.Resource, plainSecrets bool, kubeCache cache.KubeCoreCache) (*ResourceClient, error) {
	return &ResourceClient{
		kube:         kube,
		resourceName: reflect.TypeOf(resourceType).String(),
		resourceType: resourceType,
		plainSecrets: plainSecrets,
		kubeCache:    kubeCache,
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

	secret, err := rc.kube.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading secret from kubernetes")
	}
	var resource resources.Resource
	if rc.plainSecrets {
		resource, err = rc.fromPlainKubeSecret(secret)
		if err != nil {
			return nil, err
		}
	} else {
		resource, err = rc.fromKubeSecret(secret)
		if err != nil {
			return nil, err
		}
	}
	if resource == nil {
		return nil, errors.Errorf("secret %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := proto.Clone(resource).(resources.Resource)
	clone.SetMetadata(meta)

	var secret *v1.Secret
	var err error
	if rc.plainSecrets {
		secret, err = rc.toPlainKubeSecret(resource.(resources.Resource))
		if err != nil {
			return nil, err
		}
	} else {
		secret, err = rc.toKubeSecret(resource.(resources.Resource))
		if err != nil {
			return nil, err
		}
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
		if _, err := rc.kube.CoreV1().Secrets(secret.Namespace).Update(secret); err != nil {
			return nil, errors.Wrapf(err, "updating kube secret %v", secret.Name)
		}
	} else {
		if _, err := rc.kube.CoreV1().Secrets(secret.Namespace).Create(secret); err != nil {
			return nil, errors.Wrapf(err, "creating kube secret %v", secret.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(secret.Namespace, secret.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr(namespace, name)
		}
		return nil
	}

	if err := rc.kube.CoreV1().Secrets(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting secret %v", name)
	}
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	secretList, err := rc.kubeCache.SecretLister().Secrets(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing secrets in %v", namespace)
	}
	var resourceList resources.ResourceList
	for _, secret := range secretList {
		var resource resources.Resource
		if rc.plainSecrets {
			resource, err = rc.fromPlainKubeSecret(secret)
			if err != nil {
				return nil, err
			}
		} else {
			resource, err = rc.fromKubeSecret(secret)
			if err != nil {
				return nil, err
			}
		}
		// not our resource, ignore it
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
	watch := rc.kubeCache.Subscribe()

	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)

	// prevent flooding the channel with duplicates
	var previous *resources.ResourceList
	updateResourceList := func() {
		list, err := rc.List(namespace, clients.ListOpts{
			Ctx:      opts.Ctx,
			Selector: opts.Selector,
		})
		if err != nil {
			errs <- err
			return
		}
		if previous != nil {
			if list.Equal(*previous) {
				return
			}
		}
		previous = &list
		resourcesChan <- list
	}

	go func() {
		defer rc.kubeCache.Unsubscribe(watch)
		defer close(resourcesChan)
		defer close(errs)

		// watch should open up with an initial read
		updateResourceList()
		for {
			select {
			case _, ok := <-watch:
				if !ok {
					return
				}
				updateResourceList()
			case <-opts.Ctx.Done():
				return
			}
		}
	}()

	return resourcesChan, errs, nil
}

func (rc *ResourceClient) exist(namespace, name string) bool {
	_, err := rc.kube.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
