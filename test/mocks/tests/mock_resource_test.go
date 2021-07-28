package tests

import (
	"os"

	"github.com/solo-io/solo-kit/pkg/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("MockResource", func() {

	AfterEach(func() {
		Expect(os.Unsetenv("POD_NAMESPACE")).NotTo(HaveOccurred())
	})

	Context("Set and Get NamespacedStatuses", func() {
		It("Should return the same NamespacedStatuses that was Set", func() {
			mockRes := v1.MockResource{}
			namespacedStatuses := &NamespacedStatuses{
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
			mockRes.SetNamespacedStatuses(namespacedStatuses)

			Expect(mockRes.GetNamespacedStatuses()).To(BeEquivalentTo(namespacedStatuses))
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
			namespacedStatuses := &NamespacedStatuses{
				Statuses: map[string]*Status{
					"test-ns1": &ns1Status,
					"test-ns2": &ns2Status,
				},
			}
			mockRes.SetNamespacedStatuses(namespacedStatuses)

			SimulateInPodNamespace("test-ns1", func() {
				status, err := mockRes.GetNamespacedStatus()
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(BeEquivalentTo(&ns1Status))
			})
			SimulateInPodNamespace("test-ns2", func() {
				status, err := mockRes.GetNamespacedStatus()
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(BeEquivalentTo(&ns2Status))
			})
		})

		It("Should return a podNamespaceErr if POD_NAMESPACE is not set.", func() {
			mockRes := v1.MockResource{}
			status := Status{
				State:      Status_Accepted,
				ReportedBy: "gloo",
			}
			namespacedStatuses := &NamespacedStatuses{
				Statuses: map[string]*Status{
					"test-ns1": &status,
				},
			}
			mockRes.SetNamespacedStatuses(namespacedStatuses)

			_, err := mockRes.GetNamespacedStatus()
			Expect(err).To(HaveOccurred())
			Expect(errors.IsPodNamespace(err)).To(BeTrue())
		})
	})

	Context("UpsertNamespacedStatus", func() {
		It("Should use POD_NAMESPACE environment variable for map keys", func() {
			mockRes := v1.MockResource{}
			mockRes.SetNamespacedStatuses(&NamespacedStatuses{})
			SimulateInPodNamespace("test-ns", func() {
				Expect(mockRes.UpsertNamespacedStatus(&Status{
					State:      Status_Accepted,
					ReportedBy: "gloo",
				})).NotTo(HaveOccurred())
				for key := range mockRes.GetNamespacedStatuses().GetStatuses() {
					Expect(key).To(BeEquivalentTo("test-ns"))
				}
			})
		})

		It("Should replace an existing status by the same reporter", func() {
			mockRes := v1.MockResource{}
			mockRes.SetNamespacedStatuses(&NamespacedStatuses{})
			SimulateInPodNamespace("test-ns", func() {
				initStatus := Status{
					State:      Status_Pending,
					ReportedBy: "gloo",
				}
				changedStatus := Status{
					State:      Status_Accepted,
					ReportedBy: "gloo",
				}
				Expect(mockRes.UpsertNamespacedStatus(&initStatus)).NotTo(HaveOccurred())
				for _, status := range mockRes.GetNamespacedStatuses().GetStatuses() {
					Expect(status).To(BeEquivalentTo(&initStatus))
				}
				Expect(mockRes.UpsertNamespacedStatus(&changedStatus)).NotTo(HaveOccurred())
				Expect(mockRes.GetNamespacedStatuses().GetStatuses()).To(HaveLen(1))
				for _, status := range mockRes.GetNamespacedStatuses().GetStatuses() {
					Expect(status).To(BeEquivalentTo(&changedStatus))
				}
			})
		})

		It("Should return a podNamespaceErr if POD_NAMESPACE is not set.", func() {
			mockRes := v1.MockResource{}
			status := Status{
				State:      Status_Pending,
				ReportedBy: "gloo",
			}

			err := mockRes.UpsertNamespacedStatus(&status)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsPodNamespace(err)).To(BeTrue())
		})
	})
})

func SimulateInPodNamespace(namespace string, body func()) {
	Expect(os.Setenv("POD_NAMESPACE", namespace)).NotTo(HaveOccurred())
	body()
	Expect(os.Unsetenv("POD_NAMESPACE")).NotTo(HaveOccurred())
}
