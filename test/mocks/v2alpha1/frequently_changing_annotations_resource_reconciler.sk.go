// Code generated by solo-kit. DO NOT EDIT.

//Generated by pkg/code-generator/codegen/templates/resource_reconciler_template.go
package v2alpha1

import (
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// Option to copy anything from the original to the desired before writing. Return value of false means don't update
type TransitionFrequentlyChangingAnnotationsResourceFunc func(original, desired *FrequentlyChangingAnnotationsResource) (bool, error)

type FrequentlyChangingAnnotationsResourceReconciler interface {
	Reconcile(namespace string, desiredResources FrequentlyChangingAnnotationsResourceList, transition TransitionFrequentlyChangingAnnotationsResourceFunc, opts clients.ListOpts) error
}

func frequentlyChangingAnnotationsResourcesToResources(list FrequentlyChangingAnnotationsResourceList) resources.ResourceList {
	var resourceList resources.ResourceList
	for _, frequentlyChangingAnnotationsResource := range list {
		resourceList = append(resourceList, frequentlyChangingAnnotationsResource)
	}
	return resourceList
}

func NewFrequentlyChangingAnnotationsResourceReconciler(client FrequentlyChangingAnnotationsResourceClient, statusSetter resources.StatusSetter) FrequentlyChangingAnnotationsResourceReconciler {
	return &frequentlyChangingAnnotationsResourceReconciler{
		base: reconcile.NewReconciler(client.BaseClient(), statusSetter),
	}
}

type frequentlyChangingAnnotationsResourceReconciler struct {
	base reconcile.Reconciler
}

func (r *frequentlyChangingAnnotationsResourceReconciler) Reconcile(namespace string, desiredResources FrequentlyChangingAnnotationsResourceList, transition TransitionFrequentlyChangingAnnotationsResourceFunc, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "frequentlyChangingAnnotationsResource_reconciler")
	var transitionResources reconcile.TransitionResourcesFunc
	if transition != nil {
		transitionResources = func(original, desired resources.Resource) (bool, error) {
			return transition(original.(*FrequentlyChangingAnnotationsResource), desired.(*FrequentlyChangingAnnotationsResource))
		}
	}
	return r.base.Reconcile(namespace, frequentlyChangingAnnotationsResourcesToResources(desiredResources), transitionResources, opts)
}
