package statusutils

import (
	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
)

var _ resources.StatusClient = new(NamespacedStatusesClient)

// InputResources support multiple statuses, each set by a particular controller
// Each controller should only update its own status, so we expose a client
// with simple Get/Set capabilities. This way, the consumers of this client
// do not need to be aware of the statusReporterNamespace.
type NamespacedStatusesClient struct {
	statusReporterNamespace string
}

func NewNamespacedStatusesClient(namespace string) *NamespacedStatusesClient {
	return &NamespacedStatusesClient{statusReporterNamespace: namespace}
}

func (s *NamespacedStatusesClient) GetStatus(resource resources.InputResource) *core.Status {
	return resource.GetStatusForNamespace(s.statusReporterNamespace)
}

func (s *NamespacedStatusesClient) SetStatus(resource resources.InputResource, status *core.Status) {
	resource.SetStatusForNamespace(s.statusReporterNamespace, status)
}

// The name of the environment variable used by resource reporters
// to associate a resource status with the appropriate controller statusReporterNamespace
const PodNamespaceEnvName = "POD_NAMESPACE"

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
