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

type StatusStore struct {
	namespace string
}

func NewStatusStore(namespace string) *StatusStore {
	return &StatusStore{namespace: namespace}
}

func (s *StatusStore) GetStatus(resource resources.InputResource) *core.Status {
	return resource.GetStatusForNamespace(s.namespace)
}

func (s *StatusStore) SetStatus(resource resources.InputResource, status *core.Status) {
	resource.SetStatusForNamespace(s.namespace, status)
}

func GetStatusReporterNamespaceFromEnv() (string, error) {
	podNamespace := os.Getenv(PodNamespaceEnvName)
	if podNamespace == "" {
		return podNamespace, errors.NewPodNamespaceErr()
	}
	return podNamespace, nil
}

func SetStatusForNamespace(r resources.InputResource, namespace string, status *core.Status) {
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

func GetStatusForNamespace(r resources.InputResource, namespace string) *core.Status {
	statuses := r.GetNamespacedStatuses().GetStatuses()
	if statuses == nil {
		return nil
	}

	return statuses[namespace]
}

func CopyStatusForNamespace(source, destination resources.InputResource, namespace string) {
	SetStatusForNamespace(destination, namespace, GetStatusForNamespace(source, namespace))
}

func UpdateStatusForNamespace(resource resources.InputResource, updateFunc func(status *core.Status) error, namespace string) error {
	statusForNamespace := GetStatusForNamespace(resource, namespace)

	err := updateFunc(statusForNamespace)
	if err != nil {
		return err
	}

	SetStatusForNamespace(resource, namespace, statusForNamespace)
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
