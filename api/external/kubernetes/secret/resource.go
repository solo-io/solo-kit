package secret

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	kubev1 "k8s.io/api/core/v1"
)

type Secret kubev1.Secret

func (p *Secret) Clone() *Secret {
	vp := kubev1.Secret(*p)
	copy := vp.DeepCopy()
	newP := Secret(*copy)
	return &newP
}

func (p *Secret) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *Secret) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *Secret) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}
