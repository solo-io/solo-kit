package cache

import (
	"testing"

	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

// TestIDHash uses ID field as the node hash.
type TestIDHash struct{}

// ID uses the node ID field
func (TestIDHash) ID(node *core.Node) string {
	if node == nil {
		return ""
	}
	return node.Id
}

func TestGetStatusKeys(t *testing.T) {
	cache := NewSnapshotCache(false, TestIDHash{}, nil)

	keys := cache.GetStatusKeys()
	if len(keys) != 0 {
		t.Errorf("GetStatusKeys() => got %v, wanted empty list", keys)
	}

	cache.SetSnapshot("test", nil)
	cache.CreateWatch(Request{
		VersionInfo: "",
		Node: &core.Node{
			Id: "test",
		},
	})
	keys = cache.GetStatusKeys()
	if len(keys) != 1 {
		t.Errorf("GetStatusKeys() => got %v with length %d, wanted [test]", keys, len(keys))
	}
}
