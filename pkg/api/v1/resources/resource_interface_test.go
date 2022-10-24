package resources_test

import (
	"hash"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

var _ = Describe("ResourceInterface", func() {

	It("should sort resources", func() {
		rl := ResourceList{NewMockResource("ns1", "foo-2"), NewMockResource("ns1", "foo-1")}

		sort.Sort(rl)
		Expect(rl[0].GetMetadata().Name).To(Equal("foo-1"))
	})

	It("should sort input resources", func() {
		rl := InputResourceList{NewMockResource("ns1", "foo-2"), NewMockResource("ns1", "foo-1")}

		sort.Sort(rl)
		Expect(rl[0].GetMetadata().Name).To(Equal("foo-1"))
	})

})

func NewMockResource(ns, name string) InputResource {
	return &mockResources{Ns: ns, Name: name}
}

type mockResources struct {
	Ns   string
	Name string
}

func (m *mockResources) GetMetadata() *core.Metadata {
	return &core.Metadata{
		Name:      m.Name,
		Namespace: m.Ns,
	}
}

func (m *mockResources) Hash(hasher hash.Hash64) (uint64, error) {
	// not implement
	return 0, nil
}

func (m *mockResources) MustHash() uint64 {
	// not implemented
	return 0
}

func (m *mockResources) SetMetadata(meta *core.Metadata) {
	// Not need in this test
}
func (m *mockResources) Equal(that interface{}) bool {
	if r, ok := that.(*mockResources); ok {
		return *r == *m
	}
	return false
}

// Deprecated
func (m *mockResources) GetStatus() *core.Status {
	// Not needed in this test
	panic("implement me")
}

// Deprecated
func (m *mockResources) SetStatus(status *core.Status) {
	// Not needed in this test
	panic("implement me")
}

func (m *mockResources) GetNamespacedStatuses() *core.NamespacedStatuses {
	// Not needed in this test
	panic("implement me")
}

func (m *mockResources) SetNamespacedStatuses(namespacedStatuses *core.NamespacedStatuses) {
	// Not needed in this test
	panic("implement me")
}
