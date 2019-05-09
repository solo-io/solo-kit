package util

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func ZeroResourceVersion(resource resources.Resource) {
	resources.UpdateMetadata(resource, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
}

func ZeroResourceVersions(resources ...resources.Resource) {
	for _, res := range resources {
		ZeroResourceVersion(res)
	}
}
