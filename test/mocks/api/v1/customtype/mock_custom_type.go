package customtype

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type MockCustomType struct {
	meta core.Metadata
}

func (m *MockCustomType) Clone() *MockCustomType {
	return &MockCustomType{meta: m.meta}
}

func (m *MockCustomType) GetMetadata() *core.Metadata {
	return m.meta
}

func (m *MockCustomType) SetMetadata(meta *core.Metadata) {
	m.meta = meta
}

func (m *MockCustomType) Equal(that interface{}) bool {
	return reflect.DeepEqual(m, that)
}
