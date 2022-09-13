package shared

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

func ApplyStatus(rc clients.ResourceClient, statusClient resources.StatusClient, inputResource resources.InputResource, opts clients.ApplyStatusOpts) (resources.Resource, error) {
	name := inputResource.GetMetadata().GetName()
	namespace := inputResource.GetMetadata().GetNamespace()
	res, err := rc.Read(namespace, name, clients.ReadOpts{
		Ctx:     opts.Ctx,
		Cluster: opts.Cluster,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "error reading before applying status")
	}

	inputRes, ok := res.(resources.InputResource)
	if !ok {
		return nil, errors.Errorf("error converting resource of type %T to input resource to apply status", res)
	}

	statusClient.SetStatus(inputRes, statusClient.GetStatus(inputResource))
	updatedRes, err := rc.Write(inputRes, clients.WriteOpts{
		Ctx:               opts.Ctx,
		OverwriteExisting: true,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "error writing to apply status")
	}
	return updatedRes, nil
}
