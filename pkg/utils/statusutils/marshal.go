package statusutils

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-multierror"
	v1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
)

var _ resources.StatusUnmarshaler = new(NamespacedStatusesUnmarshaler)

type NamespacedStatusesUnmarshaler struct {
	unmarshalMapToProto func(m map[string]interface{}, into proto.Message) error
	statusClient        resources.StatusClient
}

func NewNamespacedStatusesUnmarshaler(statusReporterNamespace string) *NamespacedStatusesUnmarshaler {
	return &NamespacedStatusesUnmarshaler{
		unmarshalMapToProto: protoutils.UnmarshalMapToProto,
		statusClient:        NewNamespacedStatusesClient(statusReporterNamespace),
	}
}

func (i *NamespacedStatusesUnmarshaler) UnmarshalStatus(resourceStatus v1.Status, into resources.InputResource) error {
	// Always initialize status to empty, before it was empty by default, as it was a non-pointer value.
	i.statusClient.SetStatus(into, &core.Status{})

	updateStatusFunc := func(status *core.Status) error {
		if status == nil {
			return nil
		}
		typedStatus := core.Status{}
		if err := i.unmarshalMapToProto(resourceStatus, &typedStatus); err != nil {
			return err
		}
		*status = typedStatus
		return nil
	}

	updateNamespacedStatusesFunc := func(status *core.NamespacedStatuses) error {
		if status == nil {
			return nil
		}
		typedStatus := core.NamespacedStatuses{}
		if err := i.unmarshalMapToProto(resourceStatus, &typedStatus); err != nil {
			return err
		}
		*status = typedStatus
		return nil
	}

	// Unmarshal the status from the Resource
	// To support Resources that have Statuses either of type core.Status or core.NamespacedStatuses
	//	we perform this unmarshalling in a couple of steps:
	//
	// 1. Attempt to unmarshal the status as a core.NamespacedStatus. Resources will be persisted with this type
	//	moving forward, so we attempt this unmarshalling first.
	// 2. If we are successful, complete
	// 3. If we are not successful, attempt to unmarshal the status as a core.Status.
	// 4. If we are successful, update the Status for this statusReporterNamespace
	// 5. If we are not successful, an error has occurred.
	if namespacedStatusesErr := UpdateNamespacedStatuses(into, updateNamespacedStatusesFunc); namespacedStatusesErr != nil {
		// If unmarshalling NamespacedStatuses failed, the resource likely has a Status instead.
		statusErr := UpdateStatus(into, updateStatusFunc, i.statusClient)
		if statusErr != nil {
			// There's actually something wrong if either status can't be unmarshalled.
			var multiErr *multierror.Error
			multiErr = multierror.Append(multiErr, namespacedStatusesErr)
			multiErr = multierror.Append(multiErr, statusErr)
			return multiErr
		}
	}

	return nil
}

func UpdateStatus(resource resources.InputResource, updateFunc func(status *core.Status) error, statusClient resources.StatusClient) error {
	status := statusClient.GetStatus(resource)

	err := updateFunc(status)
	if err != nil {
		return err
	}

	statusClient.SetStatus(resource, status)
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
