// Code generated by solo-kit. DO NOT EDIT.

//Source: pkg/code-generator/codegen/templates/resource_reconciler_template.go
package v1

import (
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// Option to copy anything from the original to the desired before writing. Return value of false means don't update
type TransitionKubeConfigFunc func(original, desired *KubeConfig) (bool, error)

type KubeConfigReconciler interface {
	Reconcile(namespace string, desiredResources KubeConfigList, transition TransitionKubeConfigFunc, opts clients.ListOpts) error
}

func kubeConfigsToResources(list KubeConfigList) resources.ResourceList {
	var resourceList resources.ResourceList
	for _, kubeConfig := range list {
		resourceList = append(resourceList, kubeConfig)
	}
	return resourceList
}

func NewKubeConfigReconciler(client KubeConfigClient, statusSetter resources.StatusSetter) KubeConfigReconciler {
	return &kubeConfigReconciler{
		base: reconcile.NewReconciler(client.BaseClient(), statusSetter),
	}
}

type kubeConfigReconciler struct {
	base reconcile.Reconciler
}

func (r *kubeConfigReconciler) Reconcile(namespace string, desiredResources KubeConfigList, transition TransitionKubeConfigFunc, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "kubeConfig_reconciler")
	var transitionResources reconcile.TransitionResourcesFunc
	if transition != nil {
		transitionResources = func(original, desired resources.Resource) (bool, error) {
			return transition(original.(*KubeConfig), desired.(*KubeConfig))
		}
	}
	return r.base.Reconcile(namespace, kubeConfigsToResources(desiredResources), transitionResources, opts)
}
