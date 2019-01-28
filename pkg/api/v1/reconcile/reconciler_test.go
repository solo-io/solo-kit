package reconcile_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	. "github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("Reconciler", func() {
	var (
		namespace          = helpers.RandString(5)
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
			Expect(mockList[i]).To(Equal(desiredMockResources[i]))
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
			Expect(mockList[i]).To(Equal(desiredMockResources[i]))
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
			Expect(mockList[i]).To(Equal(desiredMockResources[i]))
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
})
