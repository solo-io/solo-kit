package cache_test

import (
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/resource"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/types"
)

// In ADS mode a response is withheld when the request's resource names do not
// cover the snapshot's resources of that type, which happens whenever the
// snapshot knows about a resource the client has not subscribed to yet (a
// cluster it has not applied, or one it rejected). The request must stay an
// open watch through that window: the client is waiting for a response and
// only issues a new request once it gets one, so a dropped watch leaves it on
// its current configuration until it reconnects.
var _ = Describe("Snapshot cache watch retention", func() {
	const retentionNode = "watch-retention-node"

	var snapshotCache cache.SnapshotCache

	BeforeEach(func() {
		snapshotCache = cache.NewSnapshotCache(cache.CacheSettings{Ads: true, Hash: TestIDHash{}})
	})

	// endpointSnapshot builds a snapshot whose EDS resources are assignments
	// for the named clusters.
	endpointSnapshot := func(version string, clusters ...string) cache.Snapshot {
		resources := make([]cache.Resource, 0, len(clusters))
		for _, name := range clusters {
			resources = append(resources, resource.NewEnvoyResource(makeEndpoint(name)))
		}
		return cache.NewEasyGenericSnapshot(version, resources)
	}

	endpointRequest := func(version string, names ...string) cache.Request {
		return cache.Request{
			Node:          &envoy_config_core_v3.Node{Id: retentionNode},
			TypeUrl:       types.EndpointTypeV3,
			ResourceNames: names,
			VersionInfo:   version,
		}
	}

	openWatches := func() int {
		return snapshotCache.GetStatusInfo(retentionNode).GetNumWatches()
	}

	It("keeps an open watch whose response a new snapshot cannot carry, and answers it once one can", func() {
		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("1", "cluster-a"))

		initial, _ := snapshotCache.CreateWatch(endpointRequest("", "cluster-a"))
		Eventually(initial, time.Second).Should(Receive())

		// The client acks and re-requests: the watch is parked.
		parked, cancel := snapshotCache.CreateWatch(endpointRequest("1", "cluster-a"))
		defer cancel()
		Expect(openWatches()).To(Equal(1))

		// A snapshot that also carries an assignment for a cluster this client
		// has not subscribed to cannot be sent to that watch.
		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("2", "cluster-a", "cluster-b"))
		Consistently(parked, 200*time.Millisecond, 20*time.Millisecond).ShouldNot(Receive())
		Expect(openWatches()).To(Equal(1), "the watch must survive a snapshot it cannot be sent")

		// Once the snapshot is one this request can accept, the retained watch
		// receives it.
		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("3", "cluster-a"))
		var response cache.Response
		Eventually(parked, time.Second).Should(Receive(&response))
		Expect(response.Version).To(Equal("3"))
		Expect(openWatches()).To(Equal(0), "an answered watch is discarded")
	})

	It("opens a watch for a request the current snapshot cannot answer, and answers it once one can", func() {
		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("2", "cluster-a", "cluster-b"))

		// The request asks for a newer version than it holds, so the cache
		// tries to answer it immediately -- and cannot, because the snapshot
		// carries cluster-b as well.
		pending, cancel := snapshotCache.CreateWatch(endpointRequest("1", "cluster-a"))
		defer cancel()
		Consistently(pending, 200*time.Millisecond, 20*time.Millisecond).ShouldNot(Receive())
		Expect(openWatches()).To(Equal(1), "a request that could not be answered must become an open watch")

		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("3", "cluster-a"))
		var response cache.Response
		Eventually(pending, time.Second).Should(Receive(&response))
		Expect(response.Version).To(Equal("3"))
	})

	It("still answers a request the snapshot can carry immediately", func() {
		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("2", "cluster-a"))

		answered, _ := snapshotCache.CreateWatch(endpointRequest("1", "cluster-a"))
		var response cache.Response
		Eventually(answered, time.Second).Should(Receive(&response))
		Expect(response.Version).To(Equal("2"))
		Expect(openWatches()).To(Equal(0), "an immediately answered request leaves no watch behind")
	})

	It("still parks a request that is already up to date", func() {
		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("1", "cluster-a"))

		parked, cancel := snapshotCache.CreateWatch(endpointRequest("1", "cluster-a"))
		defer cancel()
		Consistently(parked, 200*time.Millisecond, 20*time.Millisecond).ShouldNot(Receive())
		Expect(openWatches()).To(Equal(1))

		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("2", "cluster-a"))
		Eventually(parked, time.Second).Should(Receive())
	})

	It("is unaffected for wildcard requests, which are never withheld", func() {
		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("1", "cluster-a"))

		initial, _ := snapshotCache.CreateWatch(endpointRequest(""))
		Eventually(initial, time.Second).Should(Receive())

		parked, cancel := snapshotCache.CreateWatch(endpointRequest("1"))
		defer cancel()
		Expect(openWatches()).To(Equal(1))

		snapshotCache.SetSnapshot(retentionNode, endpointSnapshot("2", "cluster-a", "cluster-b"))
		Eventually(parked, time.Second).Should(Receive())
		Expect(openWatches()).To(Equal(0))
	})
})
