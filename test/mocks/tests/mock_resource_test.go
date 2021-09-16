package tests

import (
	"os"

	"github.com/solo-io/solo-kit/pkg/utils/statusutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

const (
	blueNamespace  = "ns-blue"
	greenNamespace = "ns-green"
)

var _ = Describe("MockResource", func() {

	AfterEach(func() {
		Expect(os.Unsetenv(statusutils.PodNamespaceEnvName)).NotTo(HaveOccurred())
	})

	Context("Set and Get NamespacedStatuses", func() {
		It("Should return the same NamespacedStatuses that was Set", func() {
			mockRes := v1.MockResource{}
			namespacedStatuses := &NamespacedStatuses{
				Statuses: map[string]*Status{
					blueNamespace: {
						State:      Status_Accepted,
						ReportedBy: "gloo",
					},
					greenNamespace: {
						State:      Status_Pending,
						ReportedBy: "gloo",
					},
				},
			}
			mockRes.SetNamespacedStatuses(namespacedStatuses)

			Expect(mockRes.GetNamespacedStatuses()).To(BeEquivalentTo(namespacedStatuses))
		})
	})

	Context("GetStatusForNamespace", func() {
		It("Should return the correct status with respect to the namespace", func() {
			mockRes := v1.MockResource{}
			blueNamespaceStatus := Status{
				State:      Status_Accepted,
				ReportedBy: "gloo",
			}
			greenNamespaceStatus := Status{
				State:      Status_Pending,
				ReportedBy: "gloo",
			}
			namespacedStatuses := &NamespacedStatuses{
				Statuses: map[string]*Status{
					blueNamespace:  &blueNamespaceStatus,
					greenNamespace: &greenNamespaceStatus,
				},
			}
			mockRes.SetNamespacedStatuses(namespacedStatuses)

			SimulateInPodNamespace(blueNamespace, func() {
				status := mockRes.GetStatusForNamespace(blueNamespace)
				Expect(status).To(BeEquivalentTo(&blueNamespaceStatus))
			})
			SimulateInPodNamespace(greenNamespace, func() {
				status := mockRes.GetStatusForNamespace(greenNamespace)
				Expect(status).To(BeEquivalentTo(&greenNamespaceStatus))
			})
		})
	})

	Context("SetStatusForNamespace", func() {
		It("Should use namespace variable for map keys", func() {
			mockRes := v1.MockResource{}
			mockRes.SetNamespacedStatuses(&NamespacedStatuses{})
			SimulateInPodNamespace(blueNamespace, func() {
				mockRes.SetStatusForNamespace(blueNamespace, &Status{
					State:      Status_Accepted,
					ReportedBy: "gloo",
				})
				for key := range mockRes.GetNamespacedStatuses().GetStatuses() {
					Expect(key).To(BeEquivalentTo(blueNamespace))
				}
			})
		})

		It("Should replace an existing status by the same reporter", func() {
			mockRes := v1.MockResource{}
			mockRes.SetNamespacedStatuses(&NamespacedStatuses{})
			SimulateInPodNamespace(blueNamespace, func() {
				initStatus := Status{
					State:      Status_Pending,
					ReportedBy: "gloo",
				}
				changedStatus := Status{
					State:      Status_Accepted,
					ReportedBy: "gloo",
				}
				mockRes.SetStatusForNamespace(blueNamespace, &initStatus)
				for _, status := range mockRes.GetNamespacedStatuses().GetStatuses() {
					Expect(status).To(BeEquivalentTo(&initStatus))
				}
				mockRes.SetStatusForNamespace(blueNamespace, &changedStatus)
				Expect(mockRes.GetNamespacedStatuses().GetStatuses()).To(HaveLen(1))
				for _, status := range mockRes.GetNamespacedStatuses().GetStatuses() {
					Expect(status).To(BeEquivalentTo(&changedStatus))
				}
			})
		})
	})
})

func SimulateInPodNamespace(namespace string, body func()) {
	podNamespaceEnvName := statusutils.PodNamespaceEnvName
	originalPodNamespace := os.Getenv(podNamespaceEnvName)

	err := os.Setenv(podNamespaceEnvName, namespace)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	defer func() {
		err := os.Setenv(podNamespaceEnvName, originalPodNamespace)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
	}()

	body()
}
