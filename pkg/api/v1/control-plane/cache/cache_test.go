package cache

import (
	envoy_api_v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestIDHash uses ID field as the node hash.
type TestIDHash struct{}

// ID uses the node ID field
func (TestIDHash) ID(node *envoy_api_v2_core.Node) string {
	if node == nil {
		return ""
	}
	return node.Id
}

var _ = Describe("Control Plane Cache", func() {

	It("returns sane values for NewStatusInfo", func() {
		node := &envoy_api_v2_core.Node{Id: "test"}
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
			Node: &envoy_api_v2_core.Node{
				Id: "test",
			},
		})
		keys = cache.GetStatusKeys()
		Expect(keys).To(Equal([]string{"test"}))
	})
})
