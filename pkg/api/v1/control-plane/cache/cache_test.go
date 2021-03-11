package cache_test

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/resource"
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

const (
	clusterName = "test"
	routeName   = "test-route"
	version     = "x"
	version2    = "y"
)

var _ = Describe("Control Plane Cache", func() {

	It("returns sane values for NewStatusInfo", func() {
		node := &envoy_config_core_v3.Node{Id: "test"}
		info := cache.NewStatusInfo(node)

		Expect(info.GetNode()).To(Equal(node))

		Expect(info.GetNumWatches()).To(Equal(0))

		Expect(info.GetLastWatchRequestTime().IsZero()).To(BeTrue())
	})

	It("returns sane values for GetStatusKeys", func() {
		c := cache.NewSnapshotCache(false, TestIDHash{}, nil)

		keys := c.GetStatusKeys()
		Expect(len(keys)).To(Equal(0))

		c.SetSnapshot("test", nil)
		c.CreateWatch(cache.Request{
			VersionInfo: "",
			Node: &envoy_config_core_v3.Node{
				Id: "test",
			},
		})
		keys = c.GetStatusKeys()
		Expect(keys).To(Equal([]string{"test"}))
	})

	It("Setting snapshot correctly updates the version", func() {
		names := map[string][]string{
			resource.EndpointTypeV3: {clusterName},
			resource.ClusterTypeV3:  nil,
			resource.RouteTypeV3:    {routeName},
			resource.ListenerTypeV3: nil,
		}

		testTypes := []string{
			resource.EndpointTypeV3,
			resource.ClusterTypeV3,
			resource.RouteTypeV3,
			resource.ListenerTypeV3,
		}

		c := cache.NewSnapshotCache(true, TestIDHash{}, nil)
		key := "test"

		_, err := c.GetSnapshot(key)
		Expect(err).To(MatchError("no snapshot found for node test"))

		watches := make(map[string]chan cache.Response)
		for _, typ := range testTypes {
			watches[typ], _ = c.CreateWatch(cache.Request{
				TypeUrl:       typ,
				ResourceNames: names[typ],
				VersionInfo:   version,
				Node: &envoy_config_core_v3.Node{
					Id: "test",
				},
			})
		}

		snapshot := &TestSnapshot{
			Clusters: cache.NewResources(version, []cache.Resource{
				resource.NewEnvoyResource(makeCluster(clusterName)),
			}),
			Listeners: cache.NewResources(version, []cache.Resource{
				resource.NewEnvoyResource(makeHTTPListener(routeName)),
			}),
			Routes: cache.NewResources(version, []cache.Resource{
				resource.NewEnvoyResource(makeRoute(routeName, clusterName)),
			}),
		}

		err = snapshot.Consistent()
		Expect(err).ToNot(HaveOccurred())
		err = c.SetSnapshot(key, snapshot)
		Expect(err).ToNot(HaveOccurred())

		snap, err := c.GetSnapshot(key)
		Expect(err).ToNot(HaveOccurred())
		// check versions for resources
		Expect(snap.GetResources(resource.ListenerTypeV3).Version).To(Equal(version))
		Expect(snap.GetResources(resource.ClusterTypeV3).Version).To(Equal(version))
		Expect(snap.GetResources(resource.RouteTypeV3).Version).To(Equal(version))
		// endpoint resource was not set in snapshot
		Expect(snap.GetResources(resource.EndpointTypeV3).Version).To(Equal(""))

		newName := "test2"
		snapshot2 := &TestSnapshot{
			Endpoints: cache.NewResources(version2, []cache.Resource{
				resource.NewEnvoyResource(makeEndpoint(newName)),
			}),
			Clusters: cache.NewResources(version2, []cache.Resource{
				resource.NewEnvoyResource(makeEDSCluster(newName)),
			}),
		}

		err = snapshot2.Consistent()
		Expect(err).ToNot(HaveOccurred())
		err = c.SetSnapshot(key, snapshot2)
		Expect(err).ToNot(HaveOccurred())

		snap2, err := c.GetSnapshot(key)
		Expect(err).ToNot(HaveOccurred())
		// update to version y
		Expect(snap2.GetResources(resource.EndpointTypeV3).Version).To(Equal(version2))
		Expect(snap2.GetResources(resource.ClusterTypeV3).Version).To(Equal(version2))
		// the cache will reset to empty version for missing resources
		Expect(snap2.GetResources(resource.ListenerTypeV3).Version).To(Equal(""))
		Expect(snap2.GetResources(resource.RouteTypeV3).Version).To(Equal(""))
	})

})
