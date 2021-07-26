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
		Expect(os.Unsetenv("POD_NAMESPACE")).NotTo(HaveOccurred())
	})

	Context("Set and Get ReporterStatus", func() {
		It("Should return the same ReporterStatus that was Set", func() {
			mockRes := v1.MockResource{}
			reporterStatus := &ReporterStatus{
				Statuses: map[string]*Status{
					"test-ns1": {
						State:      Status_Accepted,
						ReportedBy: "gloo",
					},
					"test-ns2": {
						State:      Status_Pending,
						ReportedBy: "gloo",
					},
				},
			}
			mockRes.SetReporterStatus(reporterStatus)

			Expect(mockRes.GetReporterStatus()).To(BeEquivalentTo(reporterStatus))
		})
	})

	Context("HasReporterStatus", func() {
		It("Should be false when there is no reporter status", func() {
			mockRes := &v1.MockResource{}
			Expect(mockRes.HasReporterStatus()).To(BeFalse())
		})

		It("Should be true when there is a reporter status", func() {
			mockRes := &v1.MockResource{}
			mockRes.SetReporterStatus(&ReporterStatus{Statuses: map[string]*Status{
				"gloo-system": {
					State:      Status_Accepted,
					ReportedBy: "gloo",
				},
			}})
			Expect(mockRes.HasReporterStatus()).To(BeTrue())
		})
	})

	Context("GetNamespacedStatus", func() {
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
					"test-ns1": &ns1Status,
					"test-ns2": &ns2Status,
				},
			}
			mockRes.SetReporterStatus(reporterStatus)

			SimulateInPodNamespace("test-ns1", func() {
				Expect(mockRes.GetNamespacedStatus()).To(BeEquivalentTo(&ns1Status))
			})
			SimulateInPodNamespace("test-ns2", func() {
				Expect(mockRes.GetNamespacedStatus()).To(BeEquivalentTo(&ns2Status))
			})
		})
	})

	Context("UpsertReporterStatus", func() {
		It("Should use POD_NAMESPACE environment variable for map keys", func() {
			mockRes := v1.MockResource{}
			mockRes.SetReporterStatus(&ReporterStatus{})
			SimulateInPodNamespace("test-ns", func() {
				mockRes.UpsertReporterStatus(&Status{
					State:      Status_Accepted,
					ReportedBy: "gloo",
				})
				for key := range mockRes.GetReporterStatus().GetStatuses() {
					Expect(key).To(BeEquivalentTo("test-ns"))
				}
			})
		})

		It("Should replace an existing status by the same reporter", func() {
			mockRes := v1.MockResource{}
			mockRes.SetReporterStatus(&ReporterStatus{})
			SimulateInPodNamespace("test-ns", func() {
				initStatus := Status{
					State:      Status_Pending,
					ReportedBy: "gloo",
				}
				changedStatus := Status{
					State:      Status_Accepted,
					ReportedBy: "gloo",
				}
				mockRes.UpsertReporterStatus(&initStatus)
				for _, status := range mockRes.GetReporterStatus().GetStatuses() {
					Expect(status).To(BeEquivalentTo(&initStatus))
				}
				mockRes.UpsertReporterStatus(&changedStatus)
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
	Expect(os.Unsetenv("POD_NAMESPACE")).NotTo(HaveOccurred())
}
