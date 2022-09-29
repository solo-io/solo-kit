package shared

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// the patch metadata is capped at 2kb of data, for more see https://github.com/solo-io/solo-kit/issues/523
	// 2147483647 bytes max k8s get from etcd / 2kb per status patch ~= 1 million resources
	MaxStatusBytes = 2048
)

// ApplyStatus is used by clients that don't support patch updates to resource statuses (e.g. consul, files, in-memory)
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

// GetJsonPatchData returns the json patch data for the input resource.
// Prefer using json patch for single api call status updates when supported (e.g. k8s) to avoid ratelimiting
// to the k8s apiserver (e.g. https://github.com/solo-io/gloo/blob/a083522af0a4ce22f4d2adf3a02470f782d5a865/projects/gloo/api/v1/settings.proto#L337-L350)
func GetJsonPatchData(ctx context.Context, inputResource resources.InputResource) ([]byte, error) {
	namespacedStatuses := inputResource.GetNamespacedStatuses().GetStatuses()
	if len(namespacedStatuses) != 1 {
		// we only expect our namespace to report here; we don't want to blow away statuses from other reporters
		return nil, errors.Errorf("unexpected number of namespaces in input resource: %v", len(inputResource.GetNamespacedStatuses().GetStatuses()))
	}
	ns := ""
	for loopNs := range inputResource.GetNamespacedStatuses().GetStatuses() {
		ns = loopNs
	}
	status := inputResource.GetNamespacedStatuses().GetStatuses()[ns]

	buf := &bytes.Buffer{}
	var marshaller jsonpb.Marshaler
	marshaller.EnumsAsInts = false  // prefer jsonpb over encoding/json marshaller since it renders enum as string not int (i.e., state is human-readable)
	marshaller.EmitDefaults = false // keep status as small as possible
	err := marshaller.Marshal(buf, status)
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling input resource")
	}

	bytes := buf.Bytes()
	patch := fmt.Sprintf(`[{"op": "replace", "path": "/status/statuses/%s", "value": %s}]`, ns, string(bytes)) // only replace our status so other reporters are not affected (e.g. blue-green of gloo)
	data := []byte(patch)

	if len(data) > MaxStatusBytes {
		if contextutils.GetLogLevel() == zapcore.DebugLevel {
			contextutils.LoggerFrom(ctx).Warnf("status patch is too large, will not apply: %s", data)
		} else {
			contextutils.LoggerFrom(ctx).Warnw("status patch is too large, will not apply", zap.Int("status_bytes", len(data)))
		}
		return nil, errors.Errorf("patch is too large (%v bytes), max is %v bytes", len(data), MaxStatusBytes)
	}

	return data, nil
}
