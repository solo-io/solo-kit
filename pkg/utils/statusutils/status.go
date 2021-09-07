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

func SetStatusForPodNamespace(r resources.InputResource, status *core.Status) error {
	podNamespace := os.Getenv(PodNamespaceEnvName)
	if podNamespace == "" {
		return errors.NewPodNamespaceErr()
	}

	return setStatusForNamespace(r, status, podNamespace)
}

func GetStatusForPodNamespace(r resources.InputResource) (*core.Status, error) {
	podNamespace := os.Getenv(PodNamespaceEnvName)
	if podNamespace == "" {
		return nil, errors.NewPodNamespaceErr()
	}

	return getStatusForNamespace(r, podNamespace)
}

func UpdateStatusForPodNamespace(resource resources.InputResource, updateFunc func(status *core.Status) error) error {
	podNamespace := os.Getenv(PodNamespaceEnvName)
	if podNamespace == "" {
		return errors.NewPodNamespaceErr()
	}

	return updateStatusForNamespace(resource, updateFunc, podNamespace)
}

func CopyStatusForPodNamespace(source, destination resources.InputResource) error {
	podNamespace := os.Getenv(PodNamespaceEnvName)
	if podNamespace == "" {
		return errors.NewPodNamespaceErr()
	}

	return copyStatusForNamespace(source, destination, podNamespace)
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

func setStatusForNamespace(r resources.InputResource, status *core.Status, namespace string) error {
	statuses := r.GetNamespacedStatuses().GetStatuses()
	if statuses == nil {
		r.SetNamespacedStatuses(
			&core.NamespacedStatuses{
				Statuses: map[string]*core.Status{namespace: status},
			})
	} else {
		statuses[namespace] = status
	}

	return nil
}

func getStatusForNamespace(r resources.InputResource, namespace string) (*core.Status, error) {
	statuses := r.GetNamespacedStatuses().GetStatuses()
	if statuses == nil {
		return nil, nil
	}

	return statuses[namespace], nil
}

func updateStatusForNamespace(resource resources.InputResource, updateFunc func(status *core.Status) error, namespace string) error {
	statusForNamespace, err := getStatusForNamespace(resource, namespace)
	if err != nil {
		return err
	}
	err = updateFunc(statusForNamespace)
	if err != nil {
		return err
	}
	return setStatusForNamespace(resource, statusForNamespace, namespace)
}

func copyStatusForNamespace(source, destination resources.InputResource, namespace string) error {
	statusForNamespace, err := getStatusForNamespace(source, namespace)
	if err != nil {
		return err
	}
	return setStatusForNamespace(destination, statusForNamespace, namespace)
}
