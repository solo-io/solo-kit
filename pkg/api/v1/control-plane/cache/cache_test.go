package cache

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestIDHash uses ID field as the node hash.
type TestIDHash struct{}

// ID uses the node ID field
func (TestIDHash) ID(node *envoy_config_core_v3.Node) string {
	if node == nil {
		return ""
	}
	return node.Id
}

var _ = Describe("Control Plane Cache", func() {

	It("returns sane values for NewStatusInfo", func() {
		node := &envoy_config_core_v3.Node{Id: "test"}
		info := newStatusInfo(node)

		Expect(info.GetNode()).To(Equal(node))

		Expect(info.GetNumWatches()).To(Equal(0))

		Expect(info.GetLastWatchRequestTime().IsZero()).To(BeTrue())
	})

	It("returns sane values for GetStatusKeys", func() {
		cache := NewSnapshotCache(false, TestIDHash{}, nil)

		keys := cache.GetStatusKeys()
		Expect(len(keys)).To(Equal(0))

		cache.SetSnapshot("test", nil)
		cache.CreateWatch(Request{
			VersionInfo: "",
			Node: &envoy_config_core_v3.Node{
				Id: "test",
			},
		})
		keys = cache.GetStatusKeys()
		Expect(keys).To(Equal([]string{"test"}))
	})
})
