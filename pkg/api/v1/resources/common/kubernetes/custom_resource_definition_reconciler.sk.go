// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// Option to copy anything from the original to the desired before writing. Return value of false means don't update
type TransitionCustomResourceDefinitionFunc func(original, desired *CustomResourceDefinition) (bool, error)

type CustomResourceDefinitionReconciler interface {
	Reconcile(namespace string, desiredResources CustomResourceDefinitionList, transition TransitionCustomResourceDefinitionFunc, opts clients.ListOpts) error
}

func customResourceDefinitionsToResources(list CustomResourceDefinitionList) resources.ResourceList {
	var resourceList resources.ResourceList
	for _, customResourceDefinition := range list {
		resourceList = append(resourceList, customResourceDefinition)
	}
	return resourceList
}

func NewCustomResourceDefinitionReconciler(client CustomResourceDefinitionClient, statusSetter resources.StatusSetter) CustomResourceDefinitionReconciler {
	return &customResourceDefinitionReconciler{
		base: reconcile.NewReconciler(client.BaseClient(), statusSetter),
	}
}

type customResourceDefinitionReconciler struct {
	base reconcile.Reconciler
}

func (r *customResourceDefinitionReconciler) Reconcile(namespace string, desiredResources CustomResourceDefinitionList, transition TransitionCustomResourceDefinitionFunc, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "customResourceDefinition_reconciler")
	var transitionResources reconcile.TransitionResourcesFunc
	if transition != nil {
		transitionResources = func(original, desired resources.Resource) (bool, error) {
			return transition(original.(*CustomResourceDefinition), desired.(*CustomResourceDefinition))
		}
	}
	return r.base.Reconcile(namespace, customResourceDefinitionsToResources(desiredResources), transitionResources, opts)
}
