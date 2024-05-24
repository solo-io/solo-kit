package statusutils

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

var _ resources.PolicyStatusClient = new(PolicyStatusClient)

// InputResources support multiple statuses, each set by a particular controller
// Each controller should only update its own status, so we expose a client
// with simple Get/Set capabilities. This way, the consumers of this client
// do not need to be aware of the statusReporterNamespace.
type PolicyStatusClient struct{}

func NewPolicyStatusClient() *PolicyStatusClient {
	return &PolicyStatusClient{}
}

func (s *PolicyStatusClient) GetPolicyStatus(resource resources.PolicyResource) *core.PolicyStatus {
	return resource.GetPolicyStatus()
}

func (s *PolicyStatusClient) SetPolicyStatus(resource resources.PolicyResource, status *core.PolicyStatus) {
	resource.SetPolicyStatus(status)
}
