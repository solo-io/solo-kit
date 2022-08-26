package namespace

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"k8s.io/client-go/kubernetes"
)

var _ resources.ResourceNamespaceLister = &kubeResourceNamespaceLister{}

// if we use the list thingy I guess we could try it out, if it does not work, then lets move forward with the base client
// solution, that uses the kubernetes interface...
func NewKubeResourceNamespaceLister(kube kubernetes.Interface, cache cache.KubeCoreCache) resources.ResourceNamespaceLister {
	return &kubeResourceNamespaceLister{
		namespace: NewNamespaceClient(kube, cache),
	}
}

type kubeResourceNamespaceLister struct {
	namespace skkube.KubeNamespaceClient
}

// GetNamespaceResourceList is the kubernetes implementation that returns the list of namespaces
func (kns *kubeResourceNamespaceLister) GetNamespaceResourceList(ctx context.Context, opts resources.ResourceNamespaceListOptions) (resources.ResourceNamespaceList, error) {
	namespaces, err := kns.namespace.List(clients.TranslateResourceNamespaceListToListOptions(opts))
	if err != nil {
		return nil, err
	}
	return convertNamespaceListToResourceNamespace(namespaces), nil
}

func (kns *kubeResourceNamespaceLister) GetNamespaceResourceWatch(ctx context.Context, opts resources.ResourceNamespaceWatchOptions) (chan resources.ResourceNamespaceList, <-chan error, error) {
	wopts := clients.WatchOpts{FieldSelectors: opts.FieldSelectors, ExpressionSelector: opts.ExpressionSelectors}
	// todo look that the namespace implementation to know exacactly what the channel of errors is returning.
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
				case resourceNamespaceChan <- convertNamespaceListToResourceNamespace(namespaceList):
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

func convertNamespaceListToResourceNamespace(namespaces skkube.KubeNamespaceList) resources.ResourceNamespaceList {
	l := resources.ResourceNamespaceList{}
	for _, ns := range namespaces {
		l = append(l, resources.ResourceNamespace{Name: ns.ObjectMeta.Name})
	}
	return l
}
