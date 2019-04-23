package v1

import (
	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"k8s.io/client-go/tools/clientcmd/api"
	"reflect"
)

// represents a kubernetes rest.Config for a local or remote cluster
// multicluster RestConfigs are a custom solo-kit resource
// which are parsed from kubernetes secrets
type KubeConfig struct {
	Metadata core.Metadata
	api.Config
}

func (c *KubeConfig) GetMetadata() core.Metadata {
	return c.Metadata
}

func (c *KubeConfig) SetMetadata(meta core.Metadata) {
	c.Metadata = meta
}

func (c *KubeConfig) Equal(that interface{}) bool {
	return reflect.DeepEqual(c, that)
}

func (c *KubeConfig) Clone() *KubeConfig {
	meta := proto.Clone(&c.Metadata).(*core.Metadata)
	innerClone := c.Config.DeepCopy()
	clone := KubeConfig{Metadata: *meta, Config: *innerClone}
	return &clone
}
