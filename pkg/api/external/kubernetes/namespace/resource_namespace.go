package namespace

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

var _ resources.ResourceNamespaceLister = &kubeResourceNamespaceLister{}
var _ resources.ResourceNamespaceLister = &kubeClientResourceNamespaceLister{}

// NewKubeClientCacheResourceNamespaceLister will create a new resource namespace lister that requires the kubernestes
// client and cache.
func NewKubeClientCacheResourceNamespaceLister(kube kubernetes.Interface, cache cache.KubeCoreCache) resources.ResourceNamespaceLister {
	return &kubeResourceNamespaceLister{
		client: NewNamespaceClient(kube, cache),
	}
}

// NewKubeClientResourceNamespaceLister will create a new resource namespace lister that requires the kubernetes client
// interface.
func NewKubeClientResourceNamespaceLister(kube kubernetes.Interface) resources.ResourceNamespaceLister {
	return &kubeClientResourceNamespaceLister{
		kube: kube,
	}
}

type kubeResourceNamespaceLister struct {
	client skkube.KubeNamespaceClient
}

// GetResourceNamespaceList is the kubernetes implementation that returns the list of namespaces
func (kns *kubeResourceNamespaceLister) GetResourceNamespaceList(opts resources.ResourceNamespaceListOptions, filtered resources.ResourceNamespaceList) (resources.ResourceNamespaceList, error) {
	namespaces, err := kns.client.List(clients.TranslateResourceNamespaceListToListOptions(opts))
	if err != nil {
		return nil, err
	}
	converted := convertNamespaceListToResourceNamespace(namespaces)
	return kns.filter(converted, filtered), nil
}

// GetResourceNamespaceWatch returns a watch for events that occur on kube namespaces returning a list of all the namespaces
func (kns *kubeResourceNamespaceLister) GetResourceNamespaceWatch(opts resources.ResourceNamespaceWatchOptions, filtered resources.ResourceNamespaceList) (chan resources.ResourceNamespaceList, <-chan error, error) {
	ctx := opts.Ctx
	wopts := clients.TranslateResourceNamespaceListToWatchOptions(opts)
	namespaceChan, errorChan, err := kns.client.Watch(wopts)
	if err != nil {
		return nil, nil, err
	}

	resourceNamespaceChan := make(chan resources.ResourceNamespaceList)
	go func() {
		defer close(resourceNamespaceChan)
		for {
			select {
			case namespaceList := <-namespaceChan:
				select {
				case resourceNamespaceChan <- kns.filter(convertNamespaceListToResourceNamespace(namespaceList), filtered):
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return resourceNamespaceChan, errorChan, nil
}

func (kns *kubeResourceNamespaceLister) filter(namespaces resources.ResourceNamespaceList, filter resources.ResourceNamespaceList) resources.ResourceNamespaceList {
	filteredList := resources.ResourceNamespaceList{}
	for _, ns := range namespaces {
		add := true
		for _, wns := range filter {
			if ns.Name == wns.Name {
				add = false
				break
			}
		}
		if add {
			filteredList = append(filteredList, ns)
		}
	}
	return filteredList
}

func convertNamespaceListToResourceNamespace(namespaces skkube.KubeNamespaceList) resources.ResourceNamespaceList {
	l := make(resources.ResourceNamespaceList, len(namespaces))
	for i, ns := range namespaces {
		l[i] = resources.ResourceNamespace{Name: ns.ObjectMeta.Name}
	}
	return l
}

type kubeClientResourceNamespaceLister struct {
	kube kubernetes.Interface
}

// GetResourceNamespaceList is the kubernetes implementation that returns the list of namespaces
func (client *kubeClientResourceNamespaceLister) GetResourceNamespaceList(opts resources.ResourceNamespaceListOptions, filtered resources.ResourceNamespaceList) (resources.ResourceNamespaceList, error) {
	excludeNamespaces := client.getExcludeFieldSelector(filtered)
	namespaceList, err := client.kube.CoreV1().Namespaces().List(opts.Ctx, metav1.ListOptions{FieldSelector: excludeNamespaces, LabelSelector: opts.ExpressionSelector})
	if err != nil {
		return nil, err
	}
	return convertNamespaceListToResourceNamespaceList(namespaceList), nil
}

// GetResourceNamespaceWatch returns a watch for events that occur on kube namespaces returning a list of all the namespaces
func (client *kubeClientResourceNamespaceLister) GetResourceNamespaceWatch(opts resources.ResourceNamespaceWatchOptions, filtered resources.ResourceNamespaceList) (chan resources.ResourceNamespaceList, <-chan error, error) {
	excludeNamespaces := client.getExcludeFieldSelector(filtered)
	namespaceWatcher, err := client.kube.CoreV1().Namespaces().Watch(opts.Ctx, metav1.ListOptions{FieldSelector: excludeNamespaces, LabelSelector: opts.ExpressionSelector})
	if err != nil {
		return nil, nil, err
	}
	namespaceChan := namespaceWatcher.ResultChan()
	resourceNamespaceChan := make(chan resources.ResourceNamespaceList)
	errorChannel := make(chan error)
	go func() {
		for {
			select {
			case <-opts.Ctx.Done():
				return
			case event, ok := <-namespaceChan:
				if !ok {
					return
				}
				switch event.Type {
				case kubewatch.Error:
					errorChannel <- errors.Errorf("error with the event from watching namespaces: %v", event)
					return
				default:
					resourceNamespaceList, err := client.GetResourceNamespaceList(resources.ResourceNamespaceListOptions{
						Ctx:                opts.Ctx,
						ExpressionSelector: opts.ExpressionSelector,
					}, filtered)
					if err != nil {
						errorChannel <- errors.Wrap(err, "error getting the list of resource namespaces while watching")
						return
					}
					resourceNamespaceChan <- resourceNamespaceList
				}
			}
		}
	}()
	return resourceNamespaceChan, errorChannel, nil
}

func (client *kubeClientResourceNamespaceLister) getExcludeFieldSelector(filtered resources.ResourceNamespaceList) string {
	var buffer bytes.Buffer
	for i, rns := range filtered {
		ns := rns.Name
		if ns != "" {
			buffer.WriteString("metadata.name!=")
			buffer.WriteString(ns)
			if i < len(filtered)-1 {
				buffer.WriteByte(',')
			}
		}
	}
	return buffer.String()
}

func convertNamespaceListToResourceNamespaceList(namespaceList *kubev1.NamespaceList) resources.ResourceNamespaceList {
	resourceNamespaces := make(resources.ResourceNamespaceList, len(namespaceList.Items))
	for i, item := range namespaceList.Items {
		ns := item.Name
		resourceNamespaces[i] = resources.ResourceNamespace{Name: ns}
	}
	return resourceNamespaces
}
