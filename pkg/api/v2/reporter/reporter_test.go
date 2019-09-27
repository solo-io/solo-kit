package reporter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	rep "github.com/solo-io/solo-kit/pkg/api/v2/reporter"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("Reporter", func() {
	var (
		reporter                               rep.Reporter
		mockResourceClient, fakeResourceClient clients.ResourceClient
	)
	BeforeEach(func() {
		mockResourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		fakeResourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.FakeResource{})
		reporter = rep.NewReporter("test", mockResourceClient, fakeResourceClient)
	})
	It("reports errors for resources", func() {
		r1, err := mockResourceClient.Write(v1.NewMockResource("", "mocky"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		r2, err := mockResourceClient.Write(v1.NewMockResource("", "fakey"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		r3, err := mockResourceClient.Write(v1.NewMockResource("", "blimpy"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		resourceErrs := rep.ResourceReports{
			r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("everyone makes mistakes")},
			r2.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("try your best")},
			r3.(*v1.MockResource): rep.Report{Warnings: []string{"didn't somebody ever tell ya", "it's not gonna be easy?"}},
		}
		err = reporter.WriteReports(context.TODO(), resourceErrs, nil)
		Expect(err).NotTo(HaveOccurred())

		r1, err = mockResourceClient.Read(r1.GetMetadata().Namespace, r1.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		r2, err = mockResourceClient.Read(r2.GetMetadata().Namespace, r2.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		r3, err = mockResourceClient.Read(r3.GetMetadata().Namespace, r3.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(r1.(*v1.MockResource).GetStatus()).To(Equal(core.Status{
			State:      2,
			Reason:     "everyone makes mistakes",
			ReportedBy: "test",
		}))
		Expect(r2.(*v1.MockResource).GetStatus()).To(Equal(core.Status{
			State:      2,
			Reason:     "try your best",
			ReportedBy: "test",
		}))
		Expect(r3.(*v1.MockResource).GetStatus()).To(Equal(core.Status{
			State:      core.Status_Warning,
			Reason:     "warning: \n  didn't somebody ever tell ya\nit's not gonna be easy?",
			ReportedBy: "test",
		}))
	})

	It("handles conflict", func() {
		r1, err := mockResourceClient.Write(v1.NewMockResource("", "mocky"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		resourceErrs := rep.ResourceReports{
			r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("everyone makes mistakes")},
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
