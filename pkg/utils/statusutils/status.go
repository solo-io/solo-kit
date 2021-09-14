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

	setStatusForNamespace(r, status, podNamespace)
	return nil
}

func GetStatusForPodNamespace(r resources.InputResource) (*core.Status, error) {
	podNamespace := os.Getenv(PodNamespaceEnvName)
	if podNamespace == "" {
		return nil, errors.NewPodNamespaceErr()
	}

	return getStatusForNamespace(r, podNamespace), nil
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

	copyStatusForNamespace(source, destination, podNamespace)
	return nil
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

func setStatusForNamespace(r resources.InputResource, status *core.Status, namespace string) {
	statuses := r.GetNamespacedStatuses().GetStatuses()
	if statuses == nil {
		r.SetNamespacedStatuses(
			&core.NamespacedStatuses{
				Statuses: map[string]*core.Status{namespace: status},
			})
	} else {
		statuses[namespace] = status
	}
}

func getStatusForNamespace(r resources.InputResource, namespace string) *core.Status {
	statuses := r.GetNamespacedStatuses().GetStatuses()
	if statuses == nil {
		return nil
	}

	return statuses[namespace]
}

func updateStatusForNamespace(resource resources.InputResource, updateFunc func(status *core.Status) error, namespace string) error {
	statusForNamespace := getStatusForNamespace(resource, namespace)

	err := updateFunc(statusForNamespace)
	if err != nil {
		return err
	}

	setStatusForNamespace(resource, statusForNamespace, namespace)
	return nil
}

func copyStatusForNamespace(source, destination resources.InputResource, namespace string) {
	statusForNamespace := getStatusForNamespace(source, namespace)
	setStatusForNamespace(destination, statusForNamespace, namespace)
}
