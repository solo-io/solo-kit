package tests

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("MockResource", func() {

	AfterEach(func() {
		err := os.Setenv("POD_NAMESPACE", "")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("SetReporterStatus", func() {

		// Keys should be formatted like $POD_NAMESPACE:$REPORTER_NAME
		It("Should format ReporterStatus.statuses keys correctly", func() {
			err := os.Setenv("POD_NAMESPACE", "pod-namespace")
			Expect(err).NotTo(HaveOccurred())

			mockRes := v1.MockResource{}
			status := core.Status{
				ReportedBy: "gloo",
				State:      core.Status_Accepted,
			}
			mockRes.SetReporterStatus(&status)

			reporterStatus := mockRes.GetReporterStatus()
			Expect(reporterStatus).NotTo(BeNil())
			Expect(reporterStatus.GetStatuses()).NotTo(BeNil())
			Expect(len(reporterStatus.GetStatuses())).To(Equal(1))
			for k, v := range reporterStatus.Statuses {
				Expect(k).To(Equal("pod-namespace:gloo"))
				Expect(v).To(Equal(&status))
			}
		})

		It("Should set ReporterStatus when POD_NAMESPACE is set", func() {
			err := os.Setenv("POD_NAMESPACE", "pod-namespace")
			Expect(err).NotTo(HaveOccurred())

			mockRes := v1.MockResource{}
			mockRes.SetReporterStatus(&core.Status{
				ReportedBy: "gloo",
				State:      core.Status_Accepted,
			})

			Expect(mockRes.GetReporterStatus()).NotTo(BeNil())
		})

		It("Should not set the reporter status if POD_NAMESPACE is not set", func() {
			Expect(os.Getenv("POD_NAMESPACE")).To(Equal(""))

			mockRes := v1.MockResource{}
			mockRes.SetReporterStatus(&core.Status{
				ReportedBy: "gloo",
				State:      core.Status_Accepted,
			})

			Expect(mockRes.GetReporterStatus()).To(BeNil())
		})
	})

})
