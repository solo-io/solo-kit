package cache_test

import (
	"encoding/json"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
)

const extensionType = "type.googleapis.com/test.ExtensionConfig"

// extensionResource stands in for the resources the extension services publish
// through a GenericSnapshot -- rate limit and ext auth, among others. Its name
// comes from a field the wrapper knows about rather than from the proto's own
// identity, and the proto it wraps is not itself a cache.Resource. Copying one
// by unwrapping it to that proto loses the name and, before this was fixed,
// panicked outright.
type extensionResource struct {
	domain string
	inner  *envoy_config_endpoint_v3.ClusterLoadAssignment
}

func (r *extensionResource) Self() cache.XdsResourceReference {
	return cache.XdsResourceReference{Name: r.domain, Type: extensionType}
}
func (r *extensionResource) ResourceProto() cache.ResourceProto { return r.inner }
func (r *extensionResource) References() []cache.XdsResourceReference {
	return []cache.XdsResourceReference{{Name: "referenced", Type: "type.googleapis.com/test.Other"}}
}

var _ = Describe("GenericSnapshot", func() {

	var snapshot *cache.GenericSnapshot

	BeforeEach(func() {
		snapshot = cache.NewEasyGenericSnapshot(version, []cache.Resource{
			&extensionResource{domain: "crd", inner: &envoy_config_endpoint_v3.ClusterLoadAssignment{ClusterName: "crd"}},
			&extensionResource{domain: "custom", inner: &envoy_config_endpoint_v3.ClusterLoadAssignment{ClusterName: "custom"}},
		})
	})

	Describe("Clone", func() {

		It("does not panic", func() {
			// SnapshotCache.GetSnapshot clones unconditionally, so a snapshot
			// that cannot be cloned cannot be read at all.
			Expect(func() { snapshot.Clone() }).NotTo(Panic())
		})

		It("keeps every resource, with its version", func() {
			resources := snapshot.Clone().GetResources(extensionType)

			Expect(resources.Version).To(Equal(version))
			Expect(resources.Items).To(HaveLen(2))
			Expect(resources.Items).To(HaveKey("crd"))
			Expect(resources.Items).To(HaveKey("custom"))
		})

		It("keeps each resource's identity", func() {
			original := snapshot.GetResources(extensionType).Items["crd"]
			cloned := snapshot.Clone().GetResources(extensionType).Items["crd"]

			Expect(cloned.Self()).To(Equal(original.Self()))
			Expect(cloned.References()).To(Equal(original.References()))
			Expect(cloned.ResourceProto().(*envoy_config_endpoint_v3.ClusterLoadAssignment).GetClusterName()).
				To(Equal("crd"))
		})

		It("copies the proto rather than sharing it", func() {
			cloned := snapshot.Clone()

			snapshot.GetResources(extensionType).Items["crd"].
				ResourceProto().(*envoy_config_endpoint_v3.ClusterLoadAssignment).ClusterName = "mutated"

			Expect(cloned.GetResources(extensionType).Items["crd"].
				ResourceProto().(*envoy_config_endpoint_v3.ClusterLoadAssignment).GetClusterName()).
				To(Equal("crd"), "a clone must not be changed by writes to the original")
		})

		It("survives a round trip through the cache", func() {
			// The path that used to panic: a client subscribes, a snapshot is
			// set, and something reads the snapshot back out.
			snapshotCache := cache.NewSnapshotCache(cache.CacheSettings{Ads: true, Hash: TestIDHash{}})
			_, cancel := snapshotCache.CreateWatch(cache.Request{
				Node:    &envoy_config_core_v3.Node{Id: "ratelimit"},
				TypeUrl: extensionType,
			})
			defer cancel()
			snapshotCache.SetSnapshot("ratelimit", snapshot)

			readBack, err := snapshotCache.GetSnapshot("ratelimit")

			Expect(err).NotTo(HaveOccurred())
			Expect(readBack.GetResources(extensionType).Items).To(HaveLen(2))
		})

		It("handles an explicitly empty type", func() {
			empty := cache.NewGenericSnapshot(cache.TypedResources{
				extensionType: {Version: "empty", Items: map[string]cache.Resource{}},
			})

			Expect(func() { empty.Clone() }).NotTo(Panic())
			Expect(empty.Clone().GetResources(extensionType).Version).To(Equal("empty"))
			Expect(empty.Clone().GetResources(extensionType).Items).To(BeEmpty())
		})
	})

	Describe("MarshalJSON", func() {

		It("renders the resources rather than an empty object", func() {
			// The resources live in an unexported field, so a marshaller cannot
			// reach them on its own.
			rendered, err := json.Marshal(snapshot)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(rendered)).To(ContainSubstring(extensionType))
			Expect(string(rendered)).To(ContainSubstring("crd"))
		})

		It("renders a clone the same way", func() {
			rendered, err := json.Marshal(snapshot.Clone())

			Expect(err).NotTo(HaveOccurred())
			Expect(string(rendered)).To(ContainSubstring("crd"))
		})
	})
})
