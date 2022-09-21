package customtype

import (
	"hash"
	"hash/fnv"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type MockCustomSpecHashType struct {
	*core.Metadata
	UID  string
	Spec MockCustomSpecHashTypeSpec
}

type MockCustomSpecHashTypeSpec struct {
}

func (m *MockCustomSpecHashTypeSpec) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	hasher.Write([]byte("mock"))
	return hasher.Sum64(), nil
}

func (m *MockCustomSpecHashType) Clone() *MockCustomSpecHashType {
	return &MockCustomSpecHashType{Metadata: m.Metadata.Clone().(*core.Metadata), Spec: m.Spec}
}

func (m *MockCustomSpecHashType) GetMetadata() *core.Metadata {
	return m.Metadata
}

func (m *MockCustomSpecHashType) SetMetadata(meta *core.Metadata) {
	m.Metadata = meta
}
