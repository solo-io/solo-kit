package multicluster

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	mock_clients "github.com/solo-io/solo-kit/pkg/api/v1/clients/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster/factory/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	"k8s.io/client-go/rest"
)

var _ = Describe("MultiClusterResourceClient", func() {
	var (
		mockCtrl            *gomock.Controller
		mockClientFactory   *mocks.MockClusterClientFactory
		mockResourceClient1 *mock_clients.MockResourceClient
		mockResourceClient2 *mock_clients.MockResourceClient
		clientSet           *clusterClientManager
		resType             resources.Resource
		subject             clients.ResourceClient
		cluster1, cluster2  = "c-one", "c-two"
		config1, config2    = &rest.Config{}, &rest.Config{}
		namespace           = "test-ns"
		testErr             = errors.New("test error")
	)

	BeforeEach(func() {
		resType = &v2alpha1.MockResource{}
		mockCtrl = gomock.NewController(GinkgoT())
		mockClientFactory = mocks.NewMockClusterClientFactory(mockCtrl)
		mockResourceClient1 = mock_clients.NewMockResourceClient(mockCtrl)
		mockResourceClient2 = mock_clients.NewMockResourceClient(mockCtrl)
		clientSet = NewClusterClientManager(context.Background(), mockClientFactory)
		subject = NewMultiClusterResourceClient(resType, clientSet)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Kind", func() {
		It("works", func() {
			Expect(subject.Kind()).To(Equal("*v2alpha1.MockResource"))
		})
	})

	Describe("NewResource", func() {
		It("works", func() {
			Expect(subject.NewResource()).To(Equal(resType))
		})
	})

	Describe("Register", func() {
		It("returns nil", func() {
			Expect(subject.Register()).To(BeNil())
		})
	})

	Context("CRUD", func() {
		var (
			resource1    = &v2alpha1.MockResource{Metadata: core.Metadata{Namespace: namespace, Name: "n-one", Cluster: cluster1}}
			resource2    = &v2alpha1.MockResource{Metadata: core.Metadata{Namespace: namespace, Name: "n-two", Cluster: cluster2}}
			fakeResource = &v2alpha1.MockResource{Metadata: core.Metadata{Namespace: "fake-ns", Name: "fake", Cluster: "fake-cluster"}}
			list1        = resources.ResourceList{resource1}
			list2        = resources.ResourceList{resource2}
		)

		BeforeEach(func() {
			mockClientFactory.EXPECT().GetClient(cluster1, config1).Return(mockResourceClient1, nil)
			mockResourceClient1.EXPECT().Register().Return(nil)
			clientSet.ClusterAdded(cluster1, config1)

			mockClientFactory.EXPECT().GetClient(cluster2, config2).Return(mockResourceClient2, nil)
			mockResourceClient2.EXPECT().Register().Return(nil)
			clientSet.ClusterAdded(cluster2, config2)
		})

		Describe("Read", func() {
			It("delegates to the correct subclient", func() {
				mockResourceClient1.
					EXPECT().
					Read(resource1.Metadata.Namespace, resource1.Metadata.Name, clients.ReadOpts{Cluster: cluster1}).
					Return(resource1, nil)

				actual, err := subject.Read(resource1.Metadata.Namespace, resource1.Metadata.Name, clients.ReadOpts{Cluster: cluster1})
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(resource1))

				mockResourceClient2.
					EXPECT().
					Read(resource2.Metadata.Namespace, resource2.Metadata.Name, clients.ReadOpts{Cluster: cluster2}).
					Return(resource2, nil)

				actual, err = subject.Read(resource2.Metadata.Namespace, resource2.Metadata.Name, clients.ReadOpts{Cluster: cluster2})
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(resource2))

				mockResourceClient2.
					EXPECT().
					Read("invalid", "invalid", clients.ReadOpts{Cluster: cluster2}).
					Return(nil, testErr)

				_, err = subject.Read("invalid", "invalid", clients.ReadOpts{Cluster: cluster2})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(testErr.Error()))
			})

			It("errors when a client cannot be found for the given cluster", func() {
				_, err := subject.Read("any", "any", clients.ReadOpts{Cluster: "fake-cluster"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(NoClientForClusterError(subject.Kind(), "fake-cluster").Error()))
			})
		})

		Describe("Write", func() {
			It("delegates to the correct subclient", func() {
				mockResourceClient1.
					EXPECT().
					Write(resource1, clients.WriteOpts{}).
					Return(resource1, nil)

				actual, err := subject.Write(resource1, clients.WriteOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(resource1))

				mockResourceClient2.
					EXPECT().
					Write(resource2, clients.WriteOpts{}).
					Return(resource2, nil)

				actual, err = subject.Write(resource2, clients.WriteOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(resource2))

				mockResourceClient2.
					EXPECT().
					Write(resource2, clients.WriteOpts{}).
					Return(nil, testErr)

				_, err = subject.Write(resource2, clients.WriteOpts{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(testErr.Error()))
			})

			It("errors when a client cannot be found for the given cluster", func() {
				_, err := subject.Write(fakeResource, clients.WriteOpts{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(NoClientForClusterError(subject.Kind(), "fake-cluster").Error()))
			})
		})

		Describe("Delete", func() {
			It("delegates to the correct subclient", func() {
				mockResourceClient1.
					EXPECT().
					Delete(resource1.Metadata.Namespace, resource1.Metadata.Name, clients.DeleteOpts{Cluster: cluster1}).
					Return(nil)

				err := subject.Delete(resource1.Metadata.Namespace, resource1.Metadata.Name, clients.DeleteOpts{Cluster: cluster1})
				Expect(err).NotTo(HaveOccurred())

				mockResourceClient2.
					EXPECT().
					Delete(resource2.Metadata.Namespace, resource2.Metadata.Name, clients.DeleteOpts{Cluster: cluster2}).
					Return(nil)

				err = subject.Delete(resource2.Metadata.Namespace, resource2.Metadata.Name, clients.DeleteOpts{Cluster: cluster2})
				Expect(err).NotTo(HaveOccurred())

				mockResourceClient2.
					EXPECT().
					Delete("invalid", "invalid", clients.DeleteOpts{Cluster: cluster2}).
					Return(testErr)

				err = subject.Delete("invalid", "invalid", clients.DeleteOpts{Cluster: cluster2})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(testErr.Error()))
			})

			It("errors when a client cannot be found for the given cluster", func() {
				err := subject.Delete("any", "any", clients.DeleteOpts{Cluster: "fake-cluster"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(NoClientForClusterError(subject.Kind(), "fake-cluster").Error()))
			})
		})

		Describe("List", func() {
			It("delegates to the correct subclient", func() {
				mockResourceClient1.
					EXPECT().
					List(namespace, clients.ListOpts{Cluster: cluster1}).
					Return(list1, nil)

				actual, err := subject.List(namespace, clients.ListOpts{Cluster: cluster1})
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(list1))

				mockResourceClient2.
					EXPECT().
					List(namespace, clients.ListOpts{Cluster: cluster2}).
					Return(list2, nil)

				actual, err = subject.List(namespace, clients.ListOpts{Cluster: cluster2})
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(list2))

				mockResourceClient2.
					EXPECT().
					List("invalid", clients.ListOpts{Cluster: cluster2}).
					Return(nil, testErr)

				_, err = subject.List("invalid", clients.ListOpts{Cluster: cluster2})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(testErr.Error()))
			})

			It("errors when a client cannot be found for the given cluster", func() {
				_, err := subject.List("any", clients.ListOpts{Cluster: "fake-cluster"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(NoClientForClusterError(subject.Kind(), "fake-cluster").Error()))
			})
		})

		Describe("Watch", func() {
			It("delegates to the correct subclient", func() {
				ch1, ch2 := make(chan resources.ResourceList, 1), make(chan resources.ResourceList, 1)
				ch1 <- list1
				ch2 <- list2
				errChan1, errChan2 := make(chan error, 1), make(chan error, 1)
				errChan1 <- nil
				errChan2 <- nil

				mockResourceClient1.
					EXPECT().
					Watch(resource1.Metadata.Namespace, clients.WatchOpts{Cluster: cluster1}).
					Return(ch1, errChan1, nil)

				actual, errs, err := subject.Watch(resource1.Metadata.Namespace, clients.WatchOpts{Cluster: cluster1})
				Expect(err).NotTo(HaveOccurred())
				Expect(<-actual).To(Equal(list1))
				Expect(<-errs).To(BeNil())

				mockResourceClient2.
					EXPECT().
					Watch(resource2.Metadata.Namespace, clients.WatchOpts{Cluster: cluster2}).
					Return(ch2, errChan2, nil)

				actual, errs, err = subject.Watch(resource2.Metadata.Namespace, clients.WatchOpts{Cluster: cluster2})
				Expect(err).NotTo(HaveOccurred())
				Expect(<-actual).To(Equal(list2))
				Expect(<-errs).To(BeNil())

				mockResourceClient2.
					EXPECT().
					Watch("invalid", clients.WatchOpts{Cluster: cluster2}).
					Return(nil, nil, testErr)

				_, _, err = subject.Watch("invalid", clients.WatchOpts{Cluster: cluster2})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(testErr.Error()))
			})

			It("errors when a client cannot be found for the given cluster", func() {
				_, _, err := subject.Watch("any", clients.WatchOpts{Cluster: "fake-cluster"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(NoClientForClusterError(subject.Kind(), "fake-cluster").Error()))
			})
		})
	})
})
