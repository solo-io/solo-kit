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
	rc           clients.ResourceClient
	statusSetter resources.StatusSetter
}

func NewReconciler(resourceClient clients.ResourceClient, statusSetter resources.StatusSetter) Reconciler {
	return &reconciler{
		rc:           resourceClient,
		statusSetter: statusSetter,
	}
}

func (r *reconciler) Reconcile(namespace string, desiredResources resources.ResourceList, transition TransitionResourcesFunc, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "reconciler")
	originalResources, err := r.rc.List(namespace, opts)
	if err != nil {
		return err
	}
	originalResourcesMap := resourceListToResourceMap(originalResources)
	for _, desired := range desiredResources {
		if err := r.syncResource(opts.Ctx, desired, originalResourcesMap, transition); err != nil {
			return errors.Wrapf(err, "reconciling resource %v", desired.GetMetadata().Name)
		}
	}
	// delete unused
	desiredResourcesMap := resourceListToResourceMap(desiredResources)
	for _, original := range originalResources {
		_, ok := desiredResourcesMap[original.GetMetadata().Ref().Key()]
		if !ok {
			if err := deleteStaleResource(opts.Ctx, r.rc, original); err != nil {
				return errors.Wrapf(err, "deleting stale resource %v", original.GetMetadata().Name)
			}
		}
	}

	return nil
}

func (r *reconciler) syncResource(ctx context.Context, desired resources.Resource, originalResourcesMap map[string]resources.Resource, transition TransitionResourcesFunc) error {
	original, _ := originalResourcesMap[desired.GetMetadata().Ref().Key()]
	return errors.RetryOnConflict(retry.DefaultBackoff, func() error {
		var err error
		original, err = attemptSyncResource(ctx, desired, original, r.rc, transition, r.statusSetter)
		return err
	})
}

func attemptSyncResource(ctx context.Context, desired, original resources.Resource, client clients.ResourceClient, transition TransitionResourcesFunc, statusSetter resources.StatusSetter) (resources.Resource, error) {
	err := writeDesiredResource(ctx, desired, original, client, transition, statusSetter)
	if err == nil {
		return original, nil
	}

	updatedOriginal, readErr := refreshOriginalResource(ctx, client, desired)
	if readErr != nil {
		if errors.IsResourceVersion(err) {
			// we don't want to return the unwrapped resource version writeErr if we also had a read error
			// otherwise we could get into infinite retry loop if reads repeatedly failed (e.g., no read RBAC)
			return original, errors.Wrapf(err, "unable to read updated resource, no reason to retry resource version conflict; readErr: %v", readErr)
		}
		return original, err
	}

	return updatedOriginal, err
}

func writeDesiredResource(ctx context.Context, desired, original resources.Resource, client clients.ResourceClient, transition TransitionResourcesFunc, statusSetter resources.StatusSetter) error {
	if original != nil {
		// this is an update: update resource version, set status to 0, needs to be re-processed
		desired = updateDesiredResourceVersionAndStatus(desired, original, statusSetter)
		if transition == nil {
			transition = defaultTransition
		}
		needsUpdate, err := transition(original, desired)
		if err != nil {
			return err // default transition will never error
		}
		if !needsUpdate {
			return nil
		}
	}
	_, writeErr := client.Write(desired, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true})
	return writeErr
}

func updateDesiredResourceVersionAndStatus(desired, original resources.Resource, statusSetter resources.StatusSetter) resources.Resource {
	resources.UpdateMetadata(desired, func(meta *core.Metadata) {
		meta.ResourceVersion = original.GetMetadata().ResourceVersion
	})
	if desiredInput, ok := desired.(resources.InputResource); ok {
		statusSetter.SetStatus(desiredInput, &core.Status{})
	}
	return desired
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

func refreshOriginalResource(ctx context.Context, client clients.ResourceClient, desired resources.Resource) (resources.Resource, error) {
	updatedOriginal, err := client.Read(desired.GetMetadata().Namespace, desired.GetMetadata().Name, clients.ReadOpts{Ctx: ctx})
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnf("unable to read updated resource %v to get updated resource version; %v", desired.GetMetadata().Ref(), err.Error())
		return nil, err
	}
	return updatedOriginal, nil
}

func deleteStaleResource(ctx context.Context, rc clients.ResourceClient, original resources.Resource) error {
	return rc.Delete(original.GetMetadata().Namespace, original.GetMetadata().Name, clients.DeleteOpts{
		Ctx:            ctx,
		IgnoreNotExist: true,
	})
}

func resourceListToResourceMap(rss resources.ResourceList) map[string]resources.Resource {
	resourceMap := map[string]resources.Resource{}
	for _, resource := range rss {
		resourceMap[resource.GetMetadata().Ref().Key()] = resource
	}
	return resourceMap
}
