package multicluster

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	mocks2 "github.com/solo-io/solo-kit/pkg/api/v1/clients/mocks"
	mock_factory "github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster/factory/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster/mocks"
	"k8s.io/client-go/rest"
)

var _ = Describe("ClusterClientGetter", func() {
	var (
		ctx                context.Context
		subject            *clusterClientManager
		mockCtrl           *gomock.Controller
		factory            *mock_factory.MockClusterClientFactory
		handler            *mocks.MockClientForClusterHandler
		client1, client2   *mocks2.MockResourceClient
		cluster1, cluster2 = "one", "two"
		cfg1, cfg2         = &rest.Config{Host: "foo"}, &rest.Config{Host: "foo"}
		testErr            = eris.New("test error")
	)

	BeforeEach(func() {
		mockCtrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		factory = mock_factory.NewMockClusterClientFactory(mockCtrl)
		handler = mocks.NewMockClientForClusterHandler(mockCtrl)
		client1 = mocks2.NewMockResourceClient(mockCtrl)
		client2 = mocks2.NewMockResourceClient(mockCtrl)
		subject = NewClusterClientManager(context.Background(), factory, handler)
	})

	expectClusterAdded := func(client *mocks2.MockResourceClient, cluster string, cfg *rest.Config) {
		factory.EXPECT().GetClient(ctx, cluster, cfg).Return(client, nil)
		client.EXPECT().Register().Return(nil)
		handler.EXPECT().HandleNewClusterClient(cluster, client)
		subject.ClusterAdded(cluster, cfg)
	}

	Describe("ClusterAdded", func() {
		It("works when a client can be created and registered", func() {
			expectClusterAdded(client1, cluster1, cfg1)
			newClient, found := subject.ClientForCluster(cluster1)
			Expect(found).To(BeTrue())
			Expect(newClient).To(Equal(client1))
		})

		It("does nothing when a client cannot be created", func() {
			factory.EXPECT().GetClient(ctx, cluster1, cfg1).Return(nil, testErr)
			subject.ClusterAdded(cluster1, cfg1)
			newClient, found := subject.ClientForCluster(cluster1)
			Expect(found).To(BeFalse())
			Expect(newClient).To(BeNil())
		})

		It("does nothing when a client cannot be registered", func() {
			factory.EXPECT().GetClient(ctx, cluster1, cfg1).Return(client1, nil)
			client1.EXPECT().Register().Return(testErr)
			client1.EXPECT().Kind() // Called in error log
			subject.ClusterAdded(cluster1, cfg1)
			newClient, found := subject.ClientForCluster(cluster1)
			Expect(found).To(BeFalse())
			Expect(newClient).To(BeNil())
		})
	})

	Describe("ClusterRemoved", func() {
		It("works when client exists for cluster", func() {
			expectClusterAdded(client1, cluster1, cfg1)
			handler.EXPECT().HandleRemovedClusterClient(cluster1, client1)

			subject.ClusterRemoved(cluster1, cfg1)
			removedClient, found := subject.ClientForCluster(cluster1)
			Expect(found).To(BeFalse())
			Expect(removedClient).To(BeNil())
		})

		It("does nothing when client does not exist for cluster", func() {
			// mock handler is not called
			subject.ClusterRemoved(cluster1, cfg1)
		})
	})

	Describe("ClientForCluster", func() {
		BeforeEach(func() {
			expectClusterAdded(client1, cluster1, cfg1)
			expectClusterAdded(client2, cluster2, cfg2)
		})

		It("returns unique clients for each cluster", func() {
			actual1, found := subject.ClientForCluster(cluster1)
			Expect(found).To(BeTrue())
			Expect(actual1).To(BeIdenticalTo(client1))

			actual2, found := subject.ClientForCluster(cluster2)
			Expect(found).To(BeTrue())
			Expect(actual2).To(BeIdenticalTo(client2))

			Expect(actual1).NotTo(BeIdenticalTo(actual2))
		})

		It("returns nil, false when a client cannot be found", func() {
			actual, found := subject.ClientForCluster("cluster-does-not-exist")
			Expect(actual).To(BeNil())
			Expect(found).To(BeFalse())
		})
	})

})
