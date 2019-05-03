package reconcile_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	. "github.com/solo-io/solo-kit/pkg/multicluster/reconcile"
	"github.com/solo-io/solo-kit/test/util"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("MulticlusterReconciler", func() {
	namespace := "namespace"
	var base1, base2 *memory.ResourceClient
	var reconciler reconcile.Reconciler
	clusterName1 := "clustr1"
	clusterName2 := "clustr2"
	BeforeEach(func() {
		base1 = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		base2 = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		cluster1 := wrapper.NewClusterClient(base1, clusterName1)
		cluster2 := wrapper.NewClusterClient(base2, clusterName2)
		reconciler = NewMultiClusterReconciler(map[string]clients.ResourceClient{
			clusterName1: cluster1,
			clusterName2: cluster2,
		})
	})
	withCluster1 := func(resource resources.Resource) resources.Resource {
		resources.UpdateMetadata(resource, func(meta *core.Metadata) {
			meta.Cluster = clusterName1
		})
		return resource
	}
	withCluster2 := func(resource resources.Resource) resources.Resource {
		resources.UpdateMetadata(resource, func(meta *core.Metadata) {
			meta.Cluster = clusterName2
		})
		return resource
	}
	It("performs reconciles across clusters", func() {
		c1res1 := withCluster1(v1.NewMockResource(namespace, "c1res1"))
		c1res2 := withCluster1(v1.NewMockResource(namespace, "c1res2"))
		c2res1 := withCluster2(v1.NewMockResource(namespace, "c2res1"))
		c2res2 := withCluster2(v1.NewMockResource(namespace, "c2res2"))
		desiredResources := resources.ResourceList{
			c1res1,
			c1res2,
			c2res1,
			c2res2,
		}

		err := reconciler.Reconcile(namespace, desiredResources, nil, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())

		list1, err := base1.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())
		list1.Each(util.ZeroResourceVersion)
		Expect(list1).To(Equal(resources.ResourceList{
			c1res1,
			c1res2,
		}))
		list2, err := base2.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())
		list2.Each(util.ZeroResourceVersion)
		Expect(list2).To(Equal(resources.ResourceList{
			c2res1,
			c2res2,
		}))
	})
})
