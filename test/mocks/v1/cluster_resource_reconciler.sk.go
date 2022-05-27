// Code generated by solo-kit. DO NOT EDIT.

//Generated by pkg/code-generator/codegen/templates/resource_reconciler_template.go
package v1

import (
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// Option to copy anything from the original to the desired before writing. Return value of false means don't update
type TransitionClusterResourceFunc func(original, desired *ClusterResource) (bool, error)

type ClusterResourceReconciler interface {
	Reconcile(namespace string, desiredResources ClusterResourceList, transition TransitionClusterResourceFunc, opts clients.ListOpts) error
}

func clusterResourcesToResources(list ClusterResourceList) resources.ResourceList {
	var resourceList resources.ResourceList
	for _, clusterResource := range list {
		resourceList = append(resourceList, clusterResource)
	}
	return resourceList
}

func NewClusterResourceReconciler(client ClusterResourceClient, statusSetter resources.StatusSetter) ClusterResourceReconciler {
	return &clusterResourceReconciler{
		base: reconcile.NewReconciler(client.BaseClient(), statusSetter),
	}
}

type clusterResourceReconciler struct {
	base reconcile.Reconciler
}

func (r *clusterResourceReconciler) Reconcile(namespace string, desiredResources ClusterResourceList, transition TransitionClusterResourceFunc, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "clusterResource_reconciler")
	var transitionResources reconcile.TransitionResourcesFunc
	if transition != nil {
		transitionResources = func(original, desired resources.Resource) (bool, error) {
			return transition(original.(*ClusterResource), desired.(*ClusterResource))
		}
	}
	return r.base.Reconcile(namespace, clusterResourcesToResources(desiredResources), transitionResources, opts)
}
