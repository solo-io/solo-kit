package reconcile_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/mocks"
	. "github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/test/matchers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("Reconciler", func() {

	var (
		mockReconciler     Reconciler
		mockResourceClient clients.ResourceClient
	)

	BeforeEach(func() {
		mockResourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		mockReconciler = NewReconciler(mockResourceClient)
	})

	It("does the crudding for you so you can sip a nice coconut", func() {
		desiredMockResources := resources.ResourceList{
			v1.NewMockResource(namespace, "a1-barry"),
			v1.NewMockResource(namespace, "b2-dave"),
		}

		// creates when doesn't exist
		err := mockReconciler.Reconcile(namespace, desiredMockResources, nil, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())

		mockList, err := mockResourceClient.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())

		Expect(mockList).To(HaveLen(2))
		for i := range mockList {
			resources.UpdateMetadata(mockList[i], func(meta *core.Metadata) {
				meta.ResourceVersion = ""
			})
			Expect(mockList[i]).To(matchers.MatchProto(desiredMockResources[i].(resources.ProtoResource)))
		}

		// updates
		desiredMockResources[0].(*v1.MockResource).Data = "foo"
		desiredMockResources[1].(*v1.MockResource).Data = "bar"
		err = mockReconciler.Reconcile(namespace, desiredMockResources, nil, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())

		mockList, err = mockResourceClient.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())

		Expect(mockList).To(HaveLen(2))
		for i := range mockList {
			resources.UpdateMetadata(mockList[i], func(meta *core.Metadata) {
				meta.ResourceVersion = ""
			})
			resources.UpdateMetadata(desiredMockResources[i], func(meta *core.Metadata) {
				meta.ResourceVersion = ""
			})
			Expect(mockList[i]).To(matchers.MatchProto(desiredMockResources[i].(resources.ProtoResource)))
		}

		// updates with transition function
		tznFnc := func(original, desired resources.Resource) (bool, error) {
			originalMock, desiredMock := original.(*v1.MockResource), desired.(*v1.MockResource)
			desiredMock.Data = "some_" + originalMock.Data
			return true, nil
		}
		mockReconciler = NewReconciler(mockResourceClient)
		err = mockReconciler.Reconcile(namespace, desiredMockResources, tznFnc, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())

		mockList, err = mockResourceClient.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())

		Expect(mockList).To(HaveLen(2))
		for i := range mockList {
			resources.UpdateMetadata(mockList[i], func(meta *core.Metadata) {
				meta.ResourceVersion = ""
			})
			resources.UpdateMetadata(desiredMockResources[i], func(meta *core.Metadata) {
				meta.ResourceVersion = ""
			})
			Expect(mockList[i]).To(matchers.MatchProto(desiredMockResources[i].(resources.ProtoResource)))
			Expect(mockList[i].(*v1.MockResource).Data).To(ContainSubstring("some_"))
		}

		// clean it all up now
		desiredMockResources = resources.ResourceList{}
		err = mockReconciler.Reconcile(namespace, desiredMockResources, nil, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())

		mockList, err = mockResourceClient.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())

		Expect(mockList).To(HaveLen(0))
	})

	Context("completely mocked resource client", func() {

		var (
			mockCtrl              *gomock.Controller
			mockedResourceClient  *mocks.MockResourceClient
			mockResource          *v1.MockResource
			originalMockResources resources.ResourceList
			desiredMockResources  resources.ResourceList
		)

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			mockedResourceClient = mocks.NewMockResourceClient(mockCtrl)
			mockReconciler = NewReconciler(mockedResourceClient)

			// original state of the world
			mockResource = v1.NewMockResource(namespace, "name")
			mockResource.Metadata.Labels = map[string]string{"ver": "v1"}
			originalMockResources = resources.ResourceList{
				mockResource,
			}

			// desired state of the world; must be different than original state so we try to write updated resources
			desiredMockResource := v1.NewMockResource(namespace, "name")
			desiredMockResource.Metadata.Labels = map[string]string{"ver": "v2"}
			desiredMockResources = resources.ResourceList{
				desiredMockResource,
			}
		})

		It("handles multiple conflict", func() {
			mockedResourceClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(originalMockResources, nil)

			// first write fails due to resource version
			mockedResourceClient.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil, errors.NewResourceVersionErr("ns", "name", "given", "expected"))
			mockedResourceClient.EXPECT().Read(mockResource.Metadata.Namespace, mockResource.Metadata.Name, gomock.Any()).Return(mockResource, nil)

			// we retry, and fail again on resource version error
			mockedResourceClient.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil, errors.NewResourceVersionErr("ns", "name", "given", "expected"))
			mockedResourceClient.EXPECT().Read(mockResource.Metadata.Namespace, mockResource.Metadata.Name, gomock.Any()).Return(mockResource, nil)

			// this time we succeed to write the status
			mockedResourceClient.EXPECT().Write(gomock.Any(), gomock.Any()).Return(mockResource, nil)

			err := mockReconciler.Reconcile(namespace, desiredMockResources, nil, clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("doesn't infinite retry on resource version write error and read errors (e.g., no read RBAC)", func() {
			resVerErr := errors.NewResourceVersionErr("ns", "name", "given", "expected")

			mockedResourceClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(originalMockResources, nil)

			// first write fails due to resource version
			mockedResourceClient.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil, resVerErr)
			mockedResourceClient.EXPECT().Read(mockResource.Metadata.Namespace, mockResource.Metadata.Name, gomock.Any()).Return(nil, errors.Errorf("no read RBAC"))

			err := mockReconciler.Reconcile(namespace, desiredMockResources, nil, clients.ListOpts{})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring(resVerErr.Error())))
		})
	})

	Context("nil transition function passed", func() {
		It("does not re-write resources which are identical", func() {

			res := v1.NewMockResource(namespace, "a1-barry")
			mockResourceClient.Write(res, clients.WriteOpts{})

			// use test client to check that Write is not called
			mockReconciler = NewReconciler(&testResourceClient{errorOnWrite: true, base: mockResourceClient})
			err := mockReconciler.Reconcile(namespace, resources.ResourceList{res}, nil, clients.ListOpts{})
			// error will occur if Write was called
			Expect(err).NotTo(HaveOccurred())

		})
	})
	Context("transition function passed", func() {
		It("can update for resources which are identical", func() {
			res := v1.NewMockResource(namespace, "a1-barry")
			mockResourceClient.Write(res, clients.WriteOpts{})

			// use test client to check that Write is not called
			mockReconciler = NewReconciler(&testResourceClient{errorOnRead: true, errorOnWrite: true, base: mockResourceClient})
			err := mockReconciler.Reconcile(namespace, resources.ResourceList{res}, func(original, desired resources.Resource) (b bool, e error) {
				// always return true
				return true, nil
			}, clients.ListOpts{})
			// error will occur if Write was called
			Expect(err).To(HaveOccurred())

		})
	})
})

type testResourceClient struct {
	errorOnRead  bool
	errorOnWrite bool
	base         clients.ResourceClient
}

func (c *testResourceClient) Kind() string {
	panic("implement me")
}

func (c *testResourceClient) NewResource() resources.Resource {
	panic("implement me")
}

func (c *testResourceClient) Register() error {
	panic("implement me")
}

func (c *testResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if c.errorOnRead {
		return nil, errors.Errorf("read should not have been called")
	}
	return nil, nil
}

func (c *testResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	if c.errorOnWrite {
		return nil, errors.Errorf("write should not have been called")
	}
	return nil, nil
}

func (c *testResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	panic("implement me")
}

func (c *testResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	return c.base.List(namespace, opts)
}

func (c *testResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	panic("implement me")
}
