package customtype

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type MockCustomType struct {
	meta core.Metadata
}

func (m MockCustomType) Clone() MockCustomType {
	return m
}

func (m *MockCustomType) Reset() {}

func (m *MockCustomType) String() string {
	return fmt.Sprintf("%v", m)
}

func (m *MockCustomType) ProtoMessage() {}

func (m *MockCustomType) GetMetadata() core.Metadata {
	return m.meta
}

func (m *MockCustomType) SetMetadata(meta core.Metadata) {
	m.meta = meta
}

func (m *MockCustomType) Equal(that interface{}) bool {
	return m.meta.Equal(that)
}
