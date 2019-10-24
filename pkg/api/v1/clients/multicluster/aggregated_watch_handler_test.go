package multicluster

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	mocks2 "github.com/solo-io/solo-kit/pkg/api/v1/clients/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper/mocks"
)

var _ = Describe("Aggregated Watch Cluster Client Handler", func() {
	var (
		subject   ClientForClusterHandler
		mockCtrl  *gomock.Controller
		mockWatch *mocks.MockWatchAggregator
		client    *mocks2.MockResourceClient
		cluster   = "my-cluster"
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockWatch = mocks.NewMockWatchAggregator(mockCtrl)
		client = mocks2.NewMockResourceClient(mockCtrl)
		subject = NewAggregatedWatchClusterClientHandler(mockWatch)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("HandleNewClusterClient", func() {
		It("adds a watch to the aggregator", func() {
			mockWatch.EXPECT().AddWatch(client)
			subject.HandleNewClusterClient(cluster, client)
		})
	})

	Describe("HandleRemovedClusterClient", func() {
		It("removes a watch from the aggregator", func() {
			mockWatch.EXPECT().RemoveWatch(client)
			subject.HandleRemovedClusterClient(cluster, client)
		})
	})
})
