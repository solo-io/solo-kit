package cache_test

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/resource"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/types"
	"go.opencensus.io/stats/view"
)

// A withheld response is otherwise invisible. With retention off the watch is
// silently discarded; with it on the watch survives but still goes unanswered
// until a compatible snapshot arrives. Neither case logs above debug or moves
// any other counter, so this measure is the only thing that says how often the
// condition fires -- which is what makes the retention default decidable from
// field data rather than from argument.
var _ = Describe("Suppressed response metric", func() {

	const suppressedNode = "suppressed-response-node"

	// suppressedCount totals the rows of xds/suppressed_responses. The view is
	// process-wide and other specs record into it, so a test reads the delta
	// across the action rather than an absolute.
	suppressedCount := func() int64 {
		rows, err := view.RetrieveData(cache.SuppressedResponsesView.Name)
		Expect(err).NotTo(HaveOccurred())
		var total int64
		for _, row := range rows {
			if data, ok := row.Data.(*view.CountData); ok {
				total += data.Value
			}
		}
		return total
	}

	// withholdOneResponse asks for "a" while the snapshot also holds "b", which
	// is the ADS name mismatch that makes the cache withhold.
	withholdOneResponse := func(snapshotCache cache.SnapshotCache) func() {
		snapshotCache.SetSnapshot(suppressedNode, &TestSnapshot{
			Endpoints: cache.NewResources("v1", []cache.Resource{
				resource.NewEnvoyResource(makeEndpoint("a")),
				resource.NewEnvoyResource(makeEndpoint("b")),
			}),
		})
		_, cancel := snapshotCache.CreateWatch(cache.Request{
			TypeUrl:       types.EndpointTypeV3,
			ResourceNames: []string{"a"},
			VersionInfo:   "v0",
			Node:          &envoy_config_core_v3.Node{Id: suppressedNode},
		})
		return cancel
	}

	DescribeTable("counts a withheld response whether or not watches are retained",
		func(retention string) {
			GinkgoT().Setenv(cache.RetainUnansweredXDSWatchesEnv, retention)
			snapshotCache := cache.NewSnapshotCache(cache.CacheSettings{Ads: true, Hash: TestIDHash{}})

			before := suppressedCount()
			cancel := withholdOneResponse(snapshotCache)
			defer cancel()

			Expect(suppressedCount()).To(BeNumerically(">", before),
				"a withheld response must be counted regardless of the retention setting")
		},
		Entry("retention on", "on"),
		Entry("retention off", "off"),
	)

	It("does not count a response that was sent", func() {
		GinkgoT().Setenv(cache.RetainUnansweredXDSWatchesEnv, "on")
		snapshotCache := cache.NewSnapshotCache(cache.CacheSettings{Ads: true, Hash: TestIDHash{}})
		snapshotCache.SetSnapshot(suppressedNode, &TestSnapshot{
			Endpoints: cache.NewResources("v1", []cache.Resource{
				resource.NewEnvoyResource(makeEndpoint("a")),
			}),
		})

		before := suppressedCount()
		// The request now names every resource in the snapshot, so it is answered.
		_, cancel := snapshotCache.CreateWatch(cache.Request{
			TypeUrl:       types.EndpointTypeV3,
			ResourceNames: []string{"a"},
			VersionInfo:   "v0",
			Node:          &envoy_config_core_v3.Node{Id: suppressedNode},
		})
		defer cancel()

		Expect(suppressedCount()).To(Equal(before))
	})

	It("tags the reason so the cause is distinguishable", func() {
		GinkgoT().Setenv(cache.RetainUnansweredXDSWatchesEnv, "on")
		snapshotCache := cache.NewSnapshotCache(cache.CacheSettings{Ads: true, Hash: TestIDHash{}})
		cancel := withholdOneResponse(snapshotCache)
		defer cancel()

		rows, err := view.RetrieveData(cache.SuppressedResponsesView.Name)
		Expect(err).NotTo(HaveOccurred())

		var reasons, typeURLs []string
		for _, row := range rows {
			for _, t := range row.Tags {
				switch t.Key.Name() {
				case "reason":
					reasons = append(reasons, t.Value)
				case "type":
					typeURLs = append(typeURLs, t.Value)
				}
			}
		}
		Expect(reasons).To(ContainElement("ads_name_mismatch"))
		Expect(typeURLs).To(ContainElement(types.EndpointTypeV3))
	})
})
