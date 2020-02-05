package reconcile

import (
	"context"

	"k8s.io/client-go/util/retry"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
)

// Option to copy anything from the original to the desired before writing
type TransitionResourcesFunc func(original, desired resources.Resource) (bool, error)

type Reconciler interface {
	Reconcile(namespace string, desiredResources resources.ResourceList, transitionFunc TransitionResourcesFunc, opts clients.ListOpts) error
}

type reconciler struct {
	rc clients.ResourceClient
}

func NewReconciler(resourceClient clients.ResourceClient) Reconciler {
	return &reconciler{
		rc: resourceClient,
	}
}

func (r *reconciler) Reconcile(namespace string, desiredResources resources.ResourceList, transition TransitionResourcesFunc, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "reconciler")
	originalResources, err := r.rc.List(namespace, opts)
	if err != nil {
		return err
	}
	for _, desired := range desiredResources {
		if err := r.syncResource(opts.Ctx, desired, originalResources, transition); err != nil {
			return errors.Wrapf(err, "reconciling resource %v", desired.GetMetadata().Name)
		}
	}
	// delete unused
	for _, original := range originalResources {
		unused := findResource(original.GetMetadata().Namespace, original.GetMetadata().Name, desiredResources) == nil
		if unused {
			if err := deleteStaleResource(opts.Ctx, r.rc, original); err != nil {
				return errors.Wrapf(err, "deleting stale resource %v", original.GetMetadata().Name)
			}
		}
	}

	return nil
}

func (r *reconciler) syncResource(ctx context.Context, desired resources.Resource, originalResources resources.ResourceList, transition TransitionResourcesFunc) error {
	var overwriteExisting, alreadyAttemptedUpdate bool
	original := findResource(desired.GetMetadata().Namespace, desired.GetMetadata().Name, originalResources)

	return errors.RetryOnConflict(retry.DefaultBackoff, func() error {
		var err error
		original, err = refreshOriginalResource(ctx, r.rc, original, alreadyAttemptedUpdate)
		if err != nil {
			return err
		}

		if original != nil {
			// this is an update: update resource version, set status to 0, needs to be re-processed
			overwriteExisting = true
			resources.UpdateMetadata(desired, func(meta *core.Metadata) {
				meta.ResourceVersion = original.GetMetadata().ResourceVersion
			})
			if desiredInput, ok := desired.(resources.InputResource); ok {
				desiredInput.SetStatus(core.Status{})
			}
			if transition == nil {
				transition = defaultTransition
			}
			needsUpdate, err := transition(original, desired)
			if err != nil {
				return err
			}
			if !needsUpdate {
				return nil
			}
		}
		_, err = r.rc.Write(desired, clients.WriteOpts{Ctx: ctx, OverwriteExisting: overwriteExisting})
		alreadyAttemptedUpdate = true
		return err
	})
}

// default transition policy: only perform an update if the Hash has changed
func defaultTransition(original, desired resources.Resource) (b bool, e error) {
	equal, ok := hashutils.HashableEqual(original, desired)
	if ok {
		return !equal, nil
	}

	// default behavior: perform the update if one if the objects are not hashable
	return true, nil
}

func refreshOriginalResource(ctx context.Context, client clients.ResourceClient, original resources.Resource, alreadyAttemptedUpdate bool) (resources.Resource, error) {
	if alreadyAttemptedUpdate {
		var err error
		original, err = client.Read(original.GetMetadata().Namespace, original.GetMetadata().Name, clients.ReadOpts{Ctx: ctx})
		if err != nil {
			return nil, err
		}
	}
	return original, nil
}

func deleteStaleResource(ctx context.Context, rc clients.ResourceClient, original resources.Resource) error {
	return rc.Delete(original.GetMetadata().Namespace, original.GetMetadata().Name, clients.DeleteOpts{
		Ctx:            ctx,
		IgnoreNotExist: true,
	})
}

func findResource(namespace, name string, rss resources.ResourceList) resources.Resource {
	for _, resource := range rss {
		if resource.GetMetadata().Namespace == namespace && resource.GetMetadata().Name == name {
			return resource
		}
	}
	return nil
}
