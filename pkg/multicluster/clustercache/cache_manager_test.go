package clustercache_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	. "github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/rest"
)

var _ = Describe("Cache Manager", func() {
	var (
		manager    CacheManager
		restConfig *rest.Config
		err        error
	)

	BeforeEach(func() {
		manager, err = NewCacheManager(context.Background(), kube.FromConfig)
		Expect(err).NotTo(HaveOccurred())
	})

	It("assigns a unique cache to each cluster", func() {
		cluster1, cluster2 := "one", "two"
		cache1 := manager.GetCache(cluster1, restConfig)
		Expect(cache1).NotTo(BeNil())
		cache2 := manager.GetCache(cluster2, restConfig)
		Expect(cache2).NotTo(BeNil())
		Expect(cache1).NotTo(BeIdenticalTo(cache2))
	})

	It("assigns one and only one cache to a given cluster", func() {
		cluster := "one"
		cache := manager.GetCache(cluster, restConfig)
		Expect(cache).NotTo(BeNil())
		sameCache := manager.GetCache(cluster, restConfig)
		Expect(sameCache).NotTo(BeNil())
		Expect(cache).To(BeIdenticalTo(sameCache))
	})

	It("creates a new cache on a subsequent call to GetCache after a cluster is removed", func() {
		cluster := "one"
		firstCache := manager.GetCache(cluster, restConfig)
		Expect(firstCache).NotTo(BeNil())
		manager.ClusterRemoved(cluster, restConfig)
		secondCache := manager.GetCache(cluster, restConfig)
		Expect(secondCache).NotTo(BeNil())
		Expect(firstCache).NotTo(BeIdenticalTo(secondCache))
	})
})
