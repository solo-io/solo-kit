package external

// This is a workaround to verify that all the generated proto files that are not used in this repository are valid
import (
	_ "github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/core"
	_ "github.com/solo-io/solo-kit/pkg/api/external/envoy/type"
)
