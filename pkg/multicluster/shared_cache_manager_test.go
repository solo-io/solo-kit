package multicluster_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/multicluster"
	"k8s.io/client-go/rest"
)

var _ = FDescribe("Shared Cache Manager", func() {
	var manager multicluster.KubeSharedCacheManager

	BeforeEach(func() {
		manager = multicluster.NewKubeSharedCacheManager(context.Background())
	})

	It("assigns a unique cache to each cluster", func() {
		cluster1, cluster2 := "one", "two"
		cache1 := manager.GetCache(cluster1)
		Expect(cache1).NotTo(BeNil())
		cache2 := manager.GetCache(cluster2)
		Expect(cache2).NotTo(BeNil())
		Expect(cache1).NotTo(BeIdenticalTo(cache2))
	})

	It("assigns one and only one cache to a given cluster", func() {
		cluster := "one"
		cache := manager.GetCache(cluster)
		Expect(cache).NotTo(BeNil())
		sameCache := manager.GetCache(cluster)
		Expect(sameCache).NotTo(BeNil())
		Expect(cache).To(BeIdenticalTo(sameCache))
	})

	It("creates a new cache on a subsequent call to GetCache after a cluster is removed", func() {
		cluster := "one"
		firstCache := manager.GetCache(cluster)
		Expect(firstCache).NotTo(BeNil())
		manager.ClusterRemoved(cluster, &rest.Config{})
		secondCache := manager.GetCache(cluster)
		Expect(secondCache).NotTo(BeNil())
		Expect(firstCache).NotTo(BeIdenticalTo(secondCache))
	})
})
