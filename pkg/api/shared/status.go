package shared

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/golang/protobuf/jsonpb"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// the patch metadata is capped at 2kb of data, for more see https://github.com/solo-io/solo-kit/issues/523
	// 2147483647 bytes max k8s get from etcd / 2kb per status patch ~= 1 million resources
	MaxStatusBytes = 2048
)

var (
	// only public for unit tests!
	DisableMaxStatusSize = false

	NoNamespacedStatusesError = func(inputResource resources.InputResource) error {
		return errors.Errorf("no namespaced statuses found on input resource %s.%s (%T)",
			inputResource.GetMetadata().GetNamespace(),
			inputResource.GetMetadata().GetName(),
			inputResource)
	}
	StatusReporterNamespaceError = func(err error) error {
		return errors.Wrapf(err, "getting status reporter namespace")
	}
	NamespacedStatusNotFoundError = func(inputResource resources.InputResource, namespace string) error {
		return errors.Errorf("input resource %s.%s (%T) does not contain status for namespace %s",
			inputResource.GetMetadata().GetNamespace(),
			inputResource.GetMetadata().GetName(),
			inputResource,
			namespace)
	}
	ResourceMarshalError = func(err error, inputResource resources.InputResource) error {
		return errors.Wrapf(err, "marshalling input resource %s.%s (%T)",
			inputResource.GetMetadata().GetNamespace(),
			inputResource.GetMetadata().GetName(),
			inputResource)
	}
	PatchTooLargeError = func(data []byte) error {
		return errors.Errorf("patch is too large (%v bytes), max is %v bytes", len(data), MaxStatusBytes)
	}
)

func init() {
	DisableMaxStatusSize = os.Getenv("DISABLE_MAX_STATUS_SIZE") == "true"
}

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

// GetJsonPatchData returns the status json patch data for the input resource.
// Prefer using json patch for single api call status updates when supported (e.g. k8s) to avoid ratelimiting
// to the k8s apiserver (e.g. https://github.com/solo-io/gloo/blob/a083522af0a4ce22f4d2adf3a02470f782d5a865/projects/gloo/api/v1/settings.proto#L337-L350)
func GetJsonPatchData(ctx context.Context, inputResource resources.InputResource) ([]byte, error) {
	namespacedStatuses := inputResource.GetNamespacedStatuses().GetStatuses()
	if len(namespacedStatuses) == 0 {
		return nil, NoNamespacedStatusesError(inputResource)
	}

	// try to get the status corresponding to the pod namespace
	ns, err := statusutils.GetStatusReporterNamespaceFromEnv()
	if err != nil {
		return nil, StatusReporterNamespaceError(err)
	}
	status, ok := namespacedStatuses[ns]
	if !ok {
		return nil, NamespacedStatusNotFoundError(inputResource, ns)
	}

	buf := &bytes.Buffer{}
	var marshaller jsonpb.Marshaler
	marshaller.EnumsAsInts = false  // prefer jsonpb over encoding/json marshaller since it renders enum as string not int (i.e., state is human-readable)
	marshaller.EmitDefaults = false // keep status as small as possible
	err = marshaller.Marshal(buf, status)
	if err != nil {
		return nil, ResourceMarshalError(err, inputResource)
	}

	bytes := buf.Bytes()
	patch := fmt.Sprintf(`[{"op": "replace", "path": "/status/statuses/%s", "value": %s}]`, ns, string(bytes)) // only replace our status so other reporters are not affected (e.g. blue-green of gloo)
	data := []byte(patch)

	if !DisableMaxStatusSize && len(data) > MaxStatusBytes {
		if contextutils.GetLogLevel() == zapcore.DebugLevel {
			contextutils.LoggerFrom(ctx).Debugf("status patch is too large, will not apply: %s", data)
		} else {
			contextutils.LoggerFrom(ctx).Warnw("status patch is too large, will not apply", zap.Int("status_bytes", len(data)))
		}
		return nil, PatchTooLargeError(data)
	}

	return data, nil
}
