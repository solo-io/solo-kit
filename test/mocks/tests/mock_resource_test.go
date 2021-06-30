package tests

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("MockResource", func() {

	AfterEach(func() {
		Expect(os.Setenv("POD_NAMESPACE", "")).NotTo(HaveOccurred())
	})

	Context("Set and Get ReporterStatus", func() {
		It("Should return the same ReporterStatus that was Set", func() {
			mockRes := v1.MockResource{}
			reporterStatus := &ReporterStatus{
				Statuses: map[string]*Status{
					"test-ns1:gloo": {
						State:      Status_Accepted,
						ReportedBy: "gloo",
					},
					"test-ns2:gloo": {
						State:      Status_Pending,
						ReportedBy: "gloo",
					},
				},
			}
			mockRes.SetReporterStatus(reporterStatus)

			Expect(mockRes.GetReporterStatus()).To(BeEquivalentTo(reporterStatus))
		})
	})

	Context("GetStatusForReporter", func() {
		It("Should return the correct status with respect to the POD_NAMESPACE", func() {
			mockRes := v1.MockResource{}
			ns1Status := Status{
				State:      Status_Accepted,
				ReportedBy: "gloo",
			}
			ns2Status := Status{
				State:      Status_Pending,
				ReportedBy: "gloo",
			}
			reporterStatus := &ReporterStatus{
				Statuses: map[string]*Status{
					"test-ns1:gloo": &ns1Status,
					"test-ns2:gloo": &ns2Status,
				},
			}
			mockRes.SetReporterStatus(reporterStatus)

			SimulateInPodNamespace("test-ns1", func() {
				Expect(mockRes.GetStatusForReporter("gloo")).To(BeEquivalentTo(&ns1Status))
			})
			SimulateInPodNamespace("test-ns2", func() {
				Expect(mockRes.GetStatusForReporter("gloo")).To(BeEquivalentTo(&ns2Status))
			})
		})

		It("Should return the correct status with respect to the Status.ReportedBy", func() {
			mockRes := v1.MockResource{}
			glooStatus := Status{
				State:      Status_Accepted,
				ReportedBy: "gloo",
			}
			gatewayStatus := Status{
				State:      Status_Pending,
				ReportedBy: "gateway",
			}
			mockRes.SetReporterStatus(&ReporterStatus{
				Statuses: map[string]*Status{
					"test-ns:gloo":    &glooStatus,
					"test-ns:gateway": &gatewayStatus,
				},
			})

			SimulateInPodNamespace("test-ns", func() {
				Expect(mockRes.GetStatusForReporter("gloo")).To(BeEquivalentTo(&glooStatus))
				Expect(mockRes.GetStatusForReporter("gateway")).To(BeEquivalentTo(&gatewayStatus))
			})
		})
	})

	Context("AddToReporterStatus", func() {
		It("Should format map keys correctly", func() {
			mockRes := v1.MockResource{ReporterStatus: &ReporterStatus{}}
			SimulateInPodNamespace("test-ns", func() {
				mockRes.AddToReporterStatus(&Status{
					State:      Status_Accepted,
					ReportedBy: "gloo",
				})
				for key := range mockRes.GetReporterStatus().GetStatuses() {
					Expect(key).To(BeEquivalentTo("test-ns:gloo"))
				}
			})
		})

		It("Should replace an existing status by the same reporter", func() {
			mockRes := v1.MockResource{ReporterStatus: &ReporterStatus{}}
			SimulateInPodNamespace("test-ns", func() {
				initStatus := Status{
					State:      Status_Pending,
					ReportedBy: "gloo",
				}
				changedStatus := Status{
					State:      Status_Accepted,
					ReportedBy: "gloo",
				}
				mockRes.AddToReporterStatus(&initStatus)
				for _, status := range mockRes.GetReporterStatus().GetStatuses() {
					Expect(status).To(BeEquivalentTo(&initStatus))
				}
				mockRes.AddToReporterStatus(&changedStatus)
				Expect(mockRes.GetReporterStatus().GetStatuses()).To(HaveLen(1))
				for _, status := range mockRes.GetReporterStatus().GetStatuses() {
					Expect(status).To(BeEquivalentTo(&changedStatus))
				}
			})
		})
	})
})

func SimulateInPodNamespace(namespace string, body func()) {
	Expect(os.Setenv("POD_NAMESPACE", namespace)).NotTo(HaveOccurred())
	body()
	Expect(os.Setenv("POD_NAMESPACE", "")).NotTo(HaveOccurred())
}
