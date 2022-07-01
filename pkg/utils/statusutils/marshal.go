package statusutils

import (
	"github.com/golang/protobuf/proto"
	v1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

var _ resources.StatusUnmarshaler = new(NamespacedStatusesUnmarshaler)
var _ resources.StatusUnmarshaler = new(SingleStatusUnmarshaler)

type NamespacedStatusesUnmarshaler struct {
	unmarshalMapToProto func(m map[string]interface{}, into proto.Message) error
}

func NewNamespacedStatusesUnmarshaler(
	unmarshalMapToProto func(m map[string]interface{}, into proto.Message) error) *NamespacedStatusesUnmarshaler {
	return &NamespacedStatusesUnmarshaler{
		unmarshalMapToProto: unmarshalMapToProto,
	}
}

func (n *NamespacedStatusesUnmarshaler) UnmarshalStatus(resourceStatus v1.Status, into resources.InputResource) {
	statusToSet := &core.NamespacedStatuses{}
	if glooStatus := n.convertResourceStatusToGlooStatus(resourceStatus); glooStatus != nil {
		statusToSet = glooStatus
	}

	into.SetNamespacedStatuses(statusToSet)
}

func (n *NamespacedStatusesUnmarshaler) convertResourceStatusToGlooStatus(resourceStatus v1.Status) *core.NamespacedStatuses {
	if resourceStatus == nil {
		return nil
	}
	typedStatus := core.NamespacedStatuses{}
	if err := n.unmarshalMapToProto(resourceStatus, &typedStatus); err != nil {
		return nil
	}
	return &typedStatus
}

type SingleStatusUnmarshaler struct {
	unmarshalMapToProto func(m map[string]interface{}, into proto.Message) error
}

func (s *SingleStatusUnmarshaler) UnmarshalStatus(status v1.Status, into resources.InputResource) {
	statusToSet := &core.Status{}
	if glooStatus := s.convertResourceStatusToGlooStatus(status); glooStatus != nil {
		statusToSet = glooStatus
	}

	into.SetStatus(statusToSet)
}

func (s *SingleStatusUnmarshaler) convertResourceStatusToGlooStatus(resourceStatus v1.Status) *core.Status {
	if resourceStatus == nil {
		return nil
	}
	typedStatus := core.Status{}
	if err := s.unmarshalMapToProto(resourceStatus, &typedStatus); err != nil {
		return nil
	}
	return &typedStatus
}
