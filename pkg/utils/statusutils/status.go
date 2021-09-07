package statusutils

import (
	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
)

// The name of the environment variable used by resource reporters
// to associate a resource status with the appropriate controller namespace
const PodNamespaceEnvName = "POD_NAMESPACE"

func SetStatusForNamespace(r resources.InputResource, status *core.Status) error {
	podNamespace := os.Getenv(PodNamespaceEnvName)
	if podNamespace == "" {
		return errors.NewPodNamespaceErr()
	}

	statuses := r.GetNamespacedStatuses().GetStatuses()
	if statuses == nil {
		r.SetNamespacedStatuses(
			&core.NamespacedStatuses{
				Statuses: map[string]*core.Status{podNamespace: status},
			})
	} else {
		statuses[podNamespace] = status
	}

	return nil
}

func GetStatusForNamespace(r resources.InputResource) (*core.Status, error) {
	podNamespace := os.Getenv(PodNamespaceEnvName)
	if podNamespace == "" {
		return nil, errors.NewPodNamespaceErr()
	}

	statuses := r.GetNamespacedStatuses().GetStatuses()
	if statuses == nil {
		return nil, nil
	}

	return statuses[podNamespace], nil
}

func UpdateStatusForNamespace(resource resources.InputResource, updateFunc func(status *core.Status) error) error {
	statusForNamespace, err := resource.GetStatusForNamespace()
	if err != nil {
		return err
	}
	err = updateFunc(statusForNamespace)
	if err != nil {
		return err
	}
	return resource.SetStatusForNamespace(statusForNamespace)
}

func UpdateNamespacedStatuses(resource resources.InputResource, updateFunc func(namespacedStatuses *core.NamespacedStatuses) error) error {
	namespacedStatuses := resource.GetNamespacedStatuses()
	err := updateFunc(namespacedStatuses)
	if err != nil {
		return err
	}
	resource.SetNamespacedStatuses(namespacedStatuses)
	return nil
}

func CopyStatusForNamespace(source, destination resources.InputResource) error {
	statusForNamespace, err := source.GetStatusForNamespace()
	if err != nil {
		return err
	}
	return destination.SetStatusForNamespace(statusForNamespace)
}
