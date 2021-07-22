package reporter_test

import (
	"context"
	"fmt"
	"os"

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
	)
	BeforeEach(func() {
		mockResourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		fakeResourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.FakeResource{})
		reporter = rep.NewReporter("test", mockResourceClient, fakeResourceClient)

		Expect(os.Setenv("POD_NAMESPACE", "test-ns")).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.Unsetenv("POD_NAMESPACE")).NotTo(HaveOccurred())
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
		Expect(r1.(*v1.MockResource).GetStatusForReporter("test")).To(Equal(&core.Status{
			State:      2,
			Reason:     "everyone makes mistakes",
			ReportedBy: "test",
		}))
		Expect(r2.(*v1.MockResource).GetStatusForReporter("test")).To(Equal(&core.Status{
			State:      2,
			Reason:     "try your best",
			ReportedBy: "test",
		}))
		Expect(r3.(*v1.MockResource).GetStatusForReporter("test")).To(Equal(&core.Status{
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
			mockCtrl             *gomock.Controller
			mockedResourceClient *mocks.MockResourceClient
		)

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			mockedResourceClient = mocks.NewMockResourceClient(mockCtrl)
			mockedResourceClient.EXPECT().Kind().Return("*v1.MockResource")
			reporter = rep.NewReporter("test", mockedResourceClient)
		})

		It("checks to make sure a resource exists before writing to it", func() {
			res := v1.NewMockResource("", "mocky")
			resourceErrs := rep.ResourceReports{
				res: rep.Report{Errors: fmt.Errorf("pocky")},
			}

			mockedResourceClient.EXPECT().Read(res.Metadata.Namespace, res.Metadata.Name, gomock.Any()).Return(nil, errors.NewNotExistErr("", "mocky"))
			// Since the resource doesn't exist, we shouldn't write to it.
			mockedResourceClient.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)

			err := reporter.WriteReports(context.TODO(), resourceErrs, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resourceErrs)).To(Equal(0))
		})

		It("handles multiple conflict", func() {
			res := v1.NewMockResource("", "mocky")
			resourceErrs := rep.ResourceReports{
				res: rep.Report{Errors: fmt.Errorf("everyone makes mistakes")},
			}

			// first write fails due to resource version
			mockedResourceClient.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil, errors.NewResourceVersionErr("ns", "name", "given", "expected"))
			mockedResourceClient.EXPECT().Read(res.Metadata.Namespace, res.Metadata.Name, gomock.Any()).Return(res, nil).Times(2)

			// we retry, and fail again on resource version error
			mockedResourceClient.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil, errors.NewResourceVersionErr("ns", "name", "given", "expected"))
			mockedResourceClient.EXPECT().Read(res.Metadata.Namespace, res.Metadata.Name, gomock.Any()).Return(res, nil).Times(2)

			// this time we succeed to write the status
			mockedResourceClient.EXPECT().Write(gomock.Any(), gomock.Any()).Return(res, nil)
			mockedResourceClient.EXPECT().Read(res.Metadata.Namespace, res.Metadata.Name, gomock.Any()).Return(res, nil)

			err := reporter.WriteReports(context.TODO(), resourceErrs, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("doesn't infinite retry on resource version write error and read errors (e.g., no read RBAC)", func() {
			res := v1.NewMockResource("", "mocky")
			resourceErrs := rep.ResourceReports{
				res: rep.Report{Errors: fmt.Errorf("everyone makes mistakes")},
			}

			resVerErr := errors.NewResourceVersionErr("ns", "name", "given", "expected")

			// first write fails due to resource version
			mockedResourceClient.EXPECT().Read(res.Metadata.Namespace, res.Metadata.Name, gomock.Any()).Return(nil, nil) // resource exists
			mockedResourceClient.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil, resVerErr)
			mockedResourceClient.EXPECT().Read(res.Metadata.Namespace, res.Metadata.Name, gomock.Any()).Return(nil, errors.Errorf("no read RBAC")).Times(2)

			err := reporter.WriteReports(context.TODO(), resourceErrs, nil)
			Expect(err).To(HaveOccurred())
			flattenedErrs := err.(*multierror.Error).WrappedErrors()
			Expect(flattenedErrs).To(HaveLen(1))
			Expect(flattenedErrs[0]).To(MatchError(ContainSubstring(resVerErr.Error())))
		})
	})
})
