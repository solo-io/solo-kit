package statusutils

import (
	"os"

	"github.com/solo-io/solo-kit/pkg/errors"
)

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
