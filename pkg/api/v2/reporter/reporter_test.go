package reporter_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/solo-kit/pkg/utils/statusutils"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	rep "github.com/solo-io/solo-kit/pkg/api/v2/reporter"
	"github.com/solo-io/solo-kit/pkg/errors"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("Reporter", func() {

	var (
		reporter                               rep.Reporter
		mockResourceClient, fakeResourceClient clients.ResourceClient

		statusClient = statusutils.NewNamespacedStatusesClient(namespace)
	)

	BeforeEach(func() {
		mockResourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		fakeResourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.FakeResource{})
		reporter = rep.NewReporter("test", statusClient, mockResourceClient, fakeResourceClient)
	})

	It("reports errors for resources", func() {
		r1, err := mockResourceClient.Write(v1.NewMockResource("", "mocky"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		r2, err := mockResourceClient.Write(v1.NewMockResource("", "fakey"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		r3, err := mockResourceClient.Write(v1.NewMockResource("", "blimpy"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		r4, err := mockResourceClient.Write(v1.NewMockResource("", "phony"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		resourceErrs := rep.ResourceReports{
			r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("everyone makes mistakes")},
			r2.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("try your best")},
			r3.(*v1.MockResource): rep.Report{Warnings: []string{"didn't somebody ever tell ya", "it's not gonna be easy?"}},
			r4.(*v1.MockResource): rep.Report{Messages: []string{"I'm just a message"}},
		}
		err = reporter.WriteReports(context.TODO(), resourceErrs, nil)
		Expect(err).NotTo(HaveOccurred())

		r1, err = mockResourceClient.Read(r1.GetMetadata().Namespace, r1.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		r2, err = mockResourceClient.Read(r2.GetMetadata().Namespace, r2.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		r3, err = mockResourceClient.Read(r3.GetMetadata().Namespace, r3.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		r4, err = mockResourceClient.Read(r4.GetMetadata().Namespace, r4.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		status := statusClient.GetStatus(r1.(*v1.MockResource))
		Expect(status).To(Equal(&core.Status{
			State:      2,
			Reason:     "everyone makes mistakes",
			ReportedBy: "test",
			Messages:   nil,
		}))

		status = statusClient.GetStatus(r2.(*v1.MockResource))
		Expect(status).To(Equal(&core.Status{
			State:      2,
			Reason:     "try your best",
			ReportedBy: "test",
			Messages:   nil,
		}))

		status = statusClient.GetStatus(r3.(*v1.MockResource))
		Expect(status).To(Equal(&core.Status{
			State:      core.Status_Warning,
			Reason:     "warning: \n  didn't somebody ever tell ya\nit's not gonna be easy?",
			ReportedBy: "test",
			Messages:   nil,
		}))

		status = statusClient.GetStatus(r4.(*v1.MockResource))
		Expect(status).To(Equal(&core.Status{
			State:      core.Status_Accepted,
			Reason:     "",
			ReportedBy: "test",
			Messages:   []string{"I'm just a message"},
		}))
	})

	It("truncates large errors", func() {
		r1, err := mockResourceClient.Write(v1.NewMockResource("", "mocky"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		var sb strings.Builder
		for i := 0; i < rep.MaxStatusBytes+1; i++ {
			sb.WriteString("a")
		}

		// an error larger than our max (1kb) that should be truncated
		veryLargeError := sb.String()

		trimmedErr := veryLargeError[:rep.MaxStatusBytes]                   // we expect to trim this to 1kb in parent. 1/2 more for each nested subresource
		recursivelyTrimmedErr := veryLargeError[:rep.MaxStatusBytes/2]      // we expect to trim this to 1kb/2
		childRecursivelyTrimmedErr := veryLargeError[:rep.MaxStatusBytes/4] // we expect to trim this to 1kb/4

		childSubresourceStatuses := map[string]*core.Status{}
		for i := 0; i < rep.MaxStatusKeys+1; i++ { // we have numerous keys, and expect to trim to num(parentkeys)/2 (i.e. rep.MaxStatusKeys/2)
			var sb strings.Builder
			for j := 0; j <= i; j++ {
				sb.WriteString("a")
			}
			childSubresourceStatuses[fmt.Sprintf("child-subresource-%s", sb.String())] = &core.Status{
				State:               core.Status_Warning,
				Reason:              veryLargeError,
				ReportedBy:          "test",
				SubresourceStatuses: nil, // intentionally nil; only test recursive call once
			}
		}

		subresourceStatuses := map[string]*core.Status{}
		for i := 0; i < rep.MaxStatusKeys+1; i++ { // we have numerous keys, and expect to trim to 100 keys (rep.MaxStatusKeys)
			var sb strings.Builder
			for j := 0; j <= i; j++ {
				sb.WriteString("a")
			}
			subresourceStatuses[fmt.Sprintf("parent-subresource-%s", sb.String())] = &core.Status{
				State:               core.Status_Warning,
				Reason:              veryLargeError,
				ReportedBy:          "test",
				SubresourceStatuses: childSubresourceStatuses,
			}
		}

		trimmedChildSubresourceStatuses := map[string]*core.Status{}
		for i := 0; i < rep.MaxStatusKeys/2; i++ { // we expect to trim to num(parentkeys)/2 (i.e. rep.MaxStatusKeys/2)
			var sb strings.Builder
			for j := 0; j <= i; j++ {
				sb.WriteString("a")
			}
			trimmedChildSubresourceStatuses[fmt.Sprintf("child-subresource-%s", sb.String())] = &core.Status{
				State:               core.Status_Warning,
				Reason:              childRecursivelyTrimmedErr,
				ReportedBy:          "test",
				SubresourceStatuses: nil, // intentionally nil; only test recursive call once
			}
		}

		trimmedSubresourceStatuses := map[string]*core.Status{}
		for i := 0; i < rep.MaxStatusKeys; i++ { // we expect to trim to 100 keys (rep.MaxStatusKeys)
			var sb strings.Builder
			for j := 0; j <= i; j++ {
				sb.WriteString("a")
			}
			trimmedSubresourceStatuses[fmt.Sprintf("parent-subresource-%s", sb.String())] = &core.Status{
				State:               core.Status_Warning,
				Reason:              recursivelyTrimmedErr,
				ReportedBy:          "test",
				SubresourceStatuses: trimmedChildSubresourceStatuses,
			}
		}

		resourceErrs := rep.ResourceReports{
			r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf(veryLargeError)},
		}
		err = reporter.WriteReports(context.TODO(), resourceErrs, subresourceStatuses)
		Expect(err).NotTo(HaveOccurred())

		r1, err = mockResourceClient.Read(r1.GetMetadata().Namespace, r1.GetMetadata().Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		status := statusClient.GetStatus(r1.(*v1.MockResource))
		Expect(status).To(Equal(&core.Status{
			State:               2,
			Reason:              trimmedErr,
			ReportedBy:          "test",
			SubresourceStatuses: trimmedSubresourceStatuses,
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

	Context("merge functionality", func() {

		getResources := func() (resources.Resource, resources.Resource, resources.Resource) {
			r1, err := mockResourceClient.Write(v1.NewMockResource("test-ns", "testres1"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			r2, err := mockResourceClient.Write(v1.NewMockResource("test-ns", "testres2"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			r3, err := mockResourceClient.Write(v1.NewMockResource("test-ns", "testres3"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(r1).NotTo(BeNil())
			Expect(r2).NotTo(BeNil())
			Expect(r3).NotTo(BeNil())
			return r1, r2, r3
		}

		It("should handle a basic merge - no overlapping resources", func() {
			r1, r2, _ := getResources()
			reports1 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r1err1"), Warnings: []string{"r1warn1"}},
			}
			reports2 := rep.ResourceReports{
				r2.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r2err1"), Warnings: []string{"r2warn1"}},
			}

			reports1.Merge(reports2)

			expectedReports := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r1err1"), Warnings: []string{"r1warn1"}},
				r2.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r2err1"), Warnings: []string{"r2warn1"}},
			}

			Expect(expectedReports).To(Equal(reports1))
		})

		It("should merge a resource with no error report with one containing an error report", func() {
			r1, _, _ := getResources()
			reports1 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r1err1")},
			}
			reports2 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{},
			}

			reports1.Merge(reports2)

			expectedReports := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r1err1")},
			}

			Expect(expectedReports).To(Equal(reports1))
		})

		It("should merge two reports with the same error on the same resource", func() {
			r1, r2, r3 := getResources()
			reports1 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1")}}, Warnings: []string{"r1warn1"}},
				r2.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r2err1")}}},
				r3.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r3err1")}}},
			}
			reports2 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1")}}, Warnings: []string{"r1warn1"}},
			}

			reports1.Merge(reports2)

			expectedReports := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1")}}, Warnings: []string{"r1warn1"}},
				r2.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r2err1")}}},
				r3.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r3err1")}}},
			}

			Expect(expectedReports).To(Equal(reports1))
		})

		It("should merge two reports with different errors on the same resource", func() {
			r1, r2, r3 := getResources()
			reports1 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1")}}, Warnings: []string{"r1warn1"}},
				r2.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r2err1")}}},
				r3.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r3err1")}}},
			}
			reports2 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err2")}}, Warnings: []string{"r1warn2"}},
			}

			reports1.Merge(reports2)

			expectedReports := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1"), fmt.Errorf("r1err2")}}, Warnings: []string{"r1warn1", "r1warn2"}},
				r2.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r2err1")}}},
				r3.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r3err1")}}},
			}

			Expect(expectedReports).To(Equal(reports1))
		})

		It("should merge two reports with warnings on both but no errors on the second", func() {
			r1, r2, r3 := getResources()
			reports1 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1")}}, Warnings: []string{"r1warn1"}},
				r2.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r2err1")}}},
				r3.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r3err1")}}},
			}
			reports2 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Warnings: []string{"r1warn2"}},
			}

			reports1.Merge(reports2)

			expectedReports := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1")}}, Warnings: []string{"r1warn1", "r1warn2"}},
				r2.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r2err1")}}},
				r3.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r3err1")}}},
			}

			Expect(expectedReports).To(Equal(reports1))
		})

		It("should merge two reports 1st with multi err 2nd with regular err", func() {
			r1, _, _ := getResources()
			reports1 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1")}}, Warnings: []string{"r1warn1"}},
			}
			reports2 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r1err2"), Warnings: []string{"r1warn2"}},
			}

			reports1.Merge(reports2)

			expectedReports := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1"), fmt.Errorf("r1err2")}}, Warnings: []string{"r1warn1", "r1warn2"}},
			}

			Expect(expectedReports).To(Equal(reports1))
		})

		It("should merge two reports 1st with regular err 2nd with multi err", func() {
			r1, _, _ := getResources()
			reports1 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r1err1"), Warnings: []string{"r1warn1"}},
			}
			reports2 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err2")}}, Warnings: []string{"r1warn2"}},
			}

			reports1.Merge(reports2)

			expectedReports := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1"), fmt.Errorf("r1err2")}}, Warnings: []string{"r1warn1", "r1warn2"}},
			}

			Expect(expectedReports).To(Equal(reports1))
		})

		It("should merge two reports both with non-multi err", func() {
			r1, _, _ := getResources()
			reports1 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r1err1"), Warnings: []string{"r1warn1"}},
			}
			reports2 := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: fmt.Errorf("r1err2"), Warnings: []string{"r1warn2"}},
			}

			reports1.Merge(reports2)

			expectedReports := rep.ResourceReports{
				r1.(*v1.MockResource): rep.Report{Errors: &multierror.Error{Errors: []error{fmt.Errorf("r1err1"), fmt.Errorf("r1err2")}}, Warnings: []string{"r1warn1", "r1warn2"}},
			}

			Expect(expectedReports).To(Equal(reports1))
		})
	})

	Context("completely mocked resource client", func() {

		var (
			ctx, reporterCtx     context.Context
			mockCtrl             *gomock.Controller
			mockedResourceClient *mocks.MockResourceClient
		)

		BeforeEach(func() {
			ctx = context.Background()
			reporterCtx = contextutils.WithLogger(ctx, "reporter")

			mockCtrl = gomock.NewController(GinkgoT())
			mockedResourceClient = mocks.NewMockResourceClient(mockCtrl)
			mockedResourceClient.EXPECT().Kind().Return("*v1.MockResource")
			reporter = rep.NewReporter("test", statusClient, mockedResourceClient)
		})

		It("handles multiple conflict", func() {
			res := v1.NewMockResource("", "mocky")
			resourceErrs := rep.ResourceReports{
				res: rep.Report{Errors: fmt.Errorf("everyone makes mistakes")},
			}

			applyOpts := clients.ApplyStatusOpts{
				Ctx: reporterCtx,
			}

			// first write fails due to resource version
			mockedResourceClient.EXPECT().ApplyStatus(gomock.Any(), gomock.Any(), applyOpts).Return(nil, errors.NewResourceVersionErr("ns", "name", "given", "expected"))

			// we retry, and fail again on resource version error
			mockedResourceClient.EXPECT().ApplyStatus(gomock.Any(), gomock.Any(), applyOpts).Return(nil, errors.NewResourceVersionErr("ns", "name", "given", "expected"))

			// this time we succeed to write the status
			mockedResourceClient.EXPECT().ApplyStatus(gomock.Any(), gomock.Any(), applyOpts).Return(res, nil)

			err := reporter.WriteReports(ctx, resourceErrs, nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
