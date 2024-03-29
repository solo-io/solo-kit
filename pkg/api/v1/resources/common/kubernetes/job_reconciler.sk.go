// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// Option to copy anything from the original to the desired before writing. Return value of false means don't update
type TransitionJobFunc func(original, desired *Job) (bool, error)

type JobReconciler interface {
	Reconcile(namespace string, desiredResources JobList, transition TransitionJobFunc, opts clients.ListOpts) error
}

func jobsToResources(list JobList) resources.ResourceList {
	var resourceList resources.ResourceList
	for _, job := range list {
		resourceList = append(resourceList, job)
	}
	return resourceList
}

func NewJobReconciler(client JobClient, statusSetter resources.StatusSetter) JobReconciler {
	return &jobReconciler{
		base: reconcile.NewReconciler(client.BaseClient(), statusSetter),
	}
}

type jobReconciler struct {
	base reconcile.Reconciler
}

func (r *jobReconciler) Reconcile(namespace string, desiredResources JobList, transition TransitionJobFunc, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "job_reconciler")
	var transitionResources reconcile.TransitionResourcesFunc
	if transition != nil {
		transitionResources = func(original, desired resources.Resource) (bool, error) {
			return transition(original.(*Job), desired.(*Job))
		}
	}
	return r.base.Reconcile(namespace, jobsToResources(desiredResources), transitionResources, opts)
}
