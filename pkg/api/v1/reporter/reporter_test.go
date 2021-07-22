package reporter_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	rep "github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("Reporter", func() {
	var (
		reporter                               rep.Reporter
		mockResourceClient, fakeResourceClient clients.ResourceClient
	)

	BeforeEach(func() {
		Expect(os.Setenv("POD_NAMESPACE", "gloo-system")).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.Unsetenv("POD_NAMESPACE")).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		mockResourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		fakeResourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.FakeResource{})
		reporter = rep.NewReporter("test", mockResourceClient, fakeResourceClient)
	})
	JustAfterEach(func() {
	})
	It("reports errors for resources", func() {
		r1, err := mockResourceClient.Write(v1.NewMockResource("", "mocky"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		r2, err := mockResourceClient.Write(v1.NewMockResource("", "fakey"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		resourceErrs := rep.ResourceErrors{
			r1.(*v1.MockResource): fmt.Errorf("everyone makes mistakes"),
			r2.(*v1.MockResource): fmt.Errorf("try your best"),
		}
		err = reporter.WriteReports(context.TODO(), resourceErrs, nil)
		Expect(err).NotTo(HaveOccurred())

		r1, err = mockResourceClient.Read(r1.GetMetadata().Namespace, r1.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		r2, err = mockResourceClient.Read(r2.GetMetadata().Namespace, r2.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		fmt.Println(r1.(*v1.MockResource).GetReporterStatus().String())

		r1Status := r1.(*v1.MockResource).GetStatusForReporter("test")
		Expect(r1Status.GetState()).To(Equal(core.Status_Rejected))
		Expect(r1Status.GetReason()).To(Equal("everyone makes mistakes"))
		Expect(r1Status.GetReportedBy()).To(Equal("test"))

		r2Status := r2.(*v1.MockResource).GetStatusForReporter("test")
		Expect(r2Status).To(Equal(&core.Status{
			State:      2,
			Reason:     "try your best",
			ReportedBy: "test",
		}))
	})

	It("handles conflict", func() {
		r1, err := mockResourceClient.Write(v1.NewMockResource("", "mocky"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		resourceErrs := rep.ResourceErrors{
			r1.(*v1.MockResource): fmt.Errorf("everyone makes mistakes"),
		}

		// write again to update resource version
		newR1 := v1.NewMockResource("", "mocky")
		newR1.Metadata.ResourceVersion = r1.GetMetadata().ResourceVersion
		r1updated, err := mockResourceClient.Write(newR1, clients.WriteOpts{OverwriteExisting: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(r1.GetMetadata().ResourceVersion).NotTo(Equal(r1updated.GetMetadata().ResourceVersion))

		err = reporter.WriteReports(context.TODO(), resourceErrs, nil)
		Expect(err).NotTo(HaveOccurred())
	})
})
