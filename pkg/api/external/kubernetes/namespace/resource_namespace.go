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
var _ resources.ResourceNamespaceLister = &kubeResourceNamespaceClient{}

// if we use the list thingy I guess we could try it out, if it does not work, then lets move forward with the base client
// solution, that uses the kubernetes interface...
func NewKubeResourceNamespaceLister(kube kubernetes.Interface, cache cache.KubeCoreCache) resources.ResourceNamespaceLister {
	return &kubeResourceNamespaceLister{
		namespace: NewNamespaceClient(kube, cache),
	}
}

func NewKubeClientResourceNamespaceLister(kube kubernetes.Interface) resources.ResourceNamespaceLister {
	return &kubeResourceNamespaceClient{
		kube: kube,
	}
}

type kubeResourceNamespaceLister struct {
	namespace skkube.KubeNamespaceClient
}

// GetNamespaceResourceList is the kubernetes implementation that returns the list of namespaces
func (kns *kubeResourceNamespaceLister) GetNamespaceResourceList(opts resources.ResourceNamespaceListOptions, filtered resources.ResourceNamespaceList) (resources.ResourceNamespaceList, error) {
	namespaces, err := kns.namespace.List(clients.TranslateResourceNamespaceListToListOptions(opts))
	if err != nil {
		return nil, err
	}
	converted := convertNamespaceListToResourceNamespace(namespaces)
	return kns.filter(converted, filtered), nil
}

// GetNamespaceResourceWatch returns a watch for events that occur on kube namespaces returning a list of all the namespaces
func (kns *kubeResourceNamespaceLister) GetNamespaceResourceWatch(opts resources.ResourceNamespaceWatchOptions, filtered resources.ResourceNamespaceList, errs chan error) (chan resources.ResourceNamespaceList, <-chan error, error) {
	ctx := opts.Ctx
	wopts := clients.TranslateResourceNamespaceListToWatchOptions(opts)
	namespaceChan, errorChan, err := kns.namespace.Watch(wopts)
	if err != nil {
		return nil, nil, err
	}

	resourceNamespaceChan := make(chan resources.ResourceNamespaceList)
	go func() {
		for {
			select {
			case namespaceList := <-namespaceChan:
				select {
				case resourceNamespaceChan <- kns.filter(convertNamespaceListToResourceNamespace(namespaceList), filtered):
				case <-ctx.Done():
					close(resourceNamespaceChan)
					return
				}
			case <-ctx.Done():
				close(resourceNamespaceChan)
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
	l := resources.ResourceNamespaceList{}
	for _, ns := range namespaces {
		l = append(l, resources.ResourceNamespace{Name: ns.ObjectMeta.Name})
	}
	return l
}

type kubeResourceNamespaceClient struct {
	kube kubernetes.Interface
}

// GetNamespaceResourceList is the kubernetes implementation that returns the list of namespaces
func (client *kubeResourceNamespaceClient) GetNamespaceResourceList(opts resources.ResourceNamespaceListOptions, filtered resources.ResourceNamespaceList) (resources.ResourceNamespaceList, error) {
	excludeNamespaces := client.getExcludeFieldSelector(filtered)
	namespaceList, err := client.kube.CoreV1().Namespaces().List(opts.Ctx, metav1.ListOptions{FieldSelector: excludeNamespaces, LabelSelector: opts.ExpressionSelector})
	if err != nil {
		return nil, err
	}
	return convertNamespaceListToResourceNamespaceList(namespaceList), nil
}

// GetNamespaceResourceWatch returns a watch for events that occur on kube namespaces returning a list of all the namespaces
func (client *kubeResourceNamespaceClient) GetNamespaceResourceWatch(opts resources.ResourceNamespaceWatchOptions, filtered resources.ResourceNamespaceList, errs chan error) (chan resources.ResourceNamespaceList, <-chan error, error) {
	excludeNamespaces := client.getExcludeFieldSelector(filtered)
	namespaceWatcher, err := client.kube.CoreV1().Namespaces().Watch(opts.Ctx, metav1.ListOptions{FieldSelector: excludeNamespaces, LabelSelector: opts.ExpressionSelector})
	if err != nil {
		return nil, nil, err
	}
	namespaceChan := namespaceWatcher.ResultChan()
	resourceNamespaceChan := make(chan resources.ResourceNamespaceList)
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
					errs <- errors.Errorf("error with the event from watching namespaces: %v", event)
					return
				default:
					resourceNamespaceList, err := client.GetNamespaceResourceList(resources.ResourceNamespaceListOptions{
						Ctx:                opts.Ctx,
						FieldSelectors:     excludeNamespaces,
						ExpressionSelector: opts.ExpressionSelector,
					}, filtered)
					if err != nil {
						errs <- errors.Wrap(err, "error getting the list of namespaces while watching")
						return
					}
					resourceNamespaceChan <- resourceNamespaceList
				}
			}
		}
	}()
	// TODO-JAKE do we need a <- chan error
	return resourceNamespaceChan, nil, nil
}

func (client *kubeResourceNamespaceClient) getExcludeFieldSelector(filtered resources.ResourceNamespaceList) string {
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
	resourceNamespaces := resources.ResourceNamespaceList{}
	for _, item := range namespaceList.Items {
		ns := item.Name
		resourceNamespaces = append(resourceNamespaces, resources.ResourceNamespace{Name: ns})
	}
	return resourceNamespaces
}
