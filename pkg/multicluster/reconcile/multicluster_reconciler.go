package reconcile

import (
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"go.uber.org/multierr"
)

type MultiClusterReconciler interface {
	reconcile.Reconciler
	AddClusterClient(cluster string, client clients.ResourceClient)
	RemoveClusterClient(cluster string)
}

/*
A MultiClusterReconciler takes a map of cluster names to resource clients
It can then perform reconciles against lists of desiredResources with resources
desired across multiple clusters
*/
type multiClusterReconciler struct {
	rcs    map[string]clients.ResourceClient
	access sync.RWMutex
}

func (r *multiClusterReconciler) AddClusterClient(cluster string, client clients.ResourceClient) {
	r.access.Lock()
	r.rcs[cluster] = client
	r.access.Unlock()
}

func (r *multiClusterReconciler) RemoveClusterClient(cluster string) {
	r.access.Lock()
	delete(r.rcs, cluster)
	r.access.Unlock()
}

func NewMultiClusterReconciler(rcs map[string]clients.ResourceClient) MultiClusterReconciler {
	return &multiClusterReconciler{rcs: rcs}
}

func (r *multiClusterReconciler) Reconcile(namespace string, desiredResources resources.ResourceList, transitionFunc reconcile.TransitionResourcesFunc, opts clients.ListOpts) error {
	byCluster := desiredResources.ByCluster()

	r.access.RLock()
	defer r.access.RUnlock()

	var errs error
	for cluster, desiredForCluster := range byCluster {
		rc, ok := r.rcs[cluster]
		if !ok {
			return errors.Errorf("no client found for cluster %v", cluster)
		}
		if err := reconcile.NewReconciler(rc).Reconcile(namespace, desiredForCluster, transitionFunc, opts); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "reconciling cluster %v", cluster))
		}
	}
	return errs
}
