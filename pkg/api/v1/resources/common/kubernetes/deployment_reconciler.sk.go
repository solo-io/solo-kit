// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// Option to copy anything from the original to the desired before writing. Return value of false means don't update
type TransitionDeploymentFunc func(original, desired *Deployment) (bool, error)

type DeploymentReconciler interface {
	Reconcile(namespace string, desiredResources DeploymentList, transition TransitionDeploymentFunc, opts clients.ListOpts) error
}

func deploymentsToResources(list DeploymentList) resources.ResourceList {
	var resourceList resources.ResourceList
	for _, deployment := range list {
		resourceList = append(resourceList, deployment)
	}
	return resourceList
}

func NewDeploymentReconciler(client DeploymentClient, statusSetter resources.StatusSetter) DeploymentReconciler {
	return &deploymentReconciler{
		base: reconcile.NewReconciler(client.BaseClient(), statusSetter),
	}
}

type deploymentReconciler struct {
	base reconcile.Reconciler
}

func (r *deploymentReconciler) Reconcile(namespace string, desiredResources DeploymentList, transition TransitionDeploymentFunc, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "deployment_reconciler")
	var transitionResources reconcile.TransitionResourcesFunc
	if transition != nil {
		transitionResources = func(original, desired resources.Resource) (bool, error) {
			return transition(original.(*Deployment), desired.(*Deployment))
		}
	}
	return r.base.Reconcile(namespace, deploymentsToResources(desiredResources), transitionResources, opts)
}
