package statusutils

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

var _ resources.StatusClient = new(NamespacedStatusesClient)
var _ resources.StatusClient = new(NoOpStatusClient)

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
	getStatusForNamespace(resource, s.statusReporterNamespace)
}

func (s *NamespacedStatusesClient) SetStatus(resource resources.InputResource, status *core.Status) {
	setStatusForNamespace(resource, status, s.statusReporterNamespace)
}

type NoOpStatusClient struct {
}

func NewNoOpStatusClient() *NoOpStatusClient {
	return &NoOpStatusClient{}
}

func (n *NoOpStatusClient) GetStatus(resource resources.InputResource) *core.Status {
	return nil
}

func (n *NoOpStatusClient) SetStatus(resource resources.InputResource, status *core.Status) {
}

func setStatusForNamespace(resource resources.InputResource, status *core.Status, namespace string) {
	statuses := resource.GetNamespacedStatuses().GetStatuses()
	if statuses == nil {
		resource.SetNamespacedStatuses(
			&core.NamespacedStatuses{
				Statuses: map[string]*core.Status{namespace: status},
			})
	} else {
		statuses[namespace] = status
	}
}

func getStatusForNamespace(resource resources.InputResource, namespace string) *core.Status {
	statuses := resource.GetNamespacedStatuses().GetStatuses()
	if statuses == nil {
		return nil
	}

	return statuses[namespace]
}

// These code is only used to support the deprecated SetStatus and GetStatus
// methods on an InputResource

const singleStatusNamespace = ""

func GetSingleStatusInNamespacedStatuses(resource resources.InputResource) *core.Status {
	return getStatusForNamespace(resource, singleStatusNamespace)
}

func SetSingleStatusInNamespacedStatuses(resource resources.InputResource, status *core.Status) {
	setStatusForNamespace(resource, status, singleStatusNamespace)
}
