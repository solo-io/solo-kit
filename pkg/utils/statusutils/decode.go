package statusutils

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type UnmarshalMapToProto func(m map[string]interface{}, into proto.Message) error

func UnmarshalInputResourceStatus(resourceStatus map[string]interface{}, into resources.InputResource, unmarshalMapToProto UnmarshalMapToProto) error {
	// Always initialize status to empty, before it was empty by default, as it was a non-pointer value.
	if err := into.SetStatusForNamespace(&core.Status{}); err != nil {
		return err
	}

	updateStatusFunc := func(status *core.Status) error {
		if status == nil {
			return nil
		}
		typedStatus := core.Status{}
		if err := unmarshalMapToProto(resourceStatus, &typedStatus); err != nil {
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
		if err := unmarshalMapToProto(resourceStatus, &typedStatus); err != nil {
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
	// 4. If we are successful, update the Status for this namespace
	// 5. If we are not successful, an error has occurred.
	if namespacedStatusesErr := UpdateNamespacedStatuses(into, updateNamespacedStatusesFunc); namespacedStatusesErr != nil {
		// If unmarshalling NamespacedStatuses failed, the resource likely has a Status instead.
		statusErr := UpdateStatusForNamespace(into, updateStatusFunc)
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
