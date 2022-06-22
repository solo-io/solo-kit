package statusutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"
	mocksv1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

const soloKitNamespace = "solo-kit"

var _ = Describe("Marshal", func() {

	var (
		statusUnmarshaler resources.StatusUnmarshaler
		inputResource     resources.InputResource
	)

	Context("NamespacedStatusUnmarshaler", func() {

		BeforeEach(func() {
			statusUnmarshaler = statusutils.NewNamespacedStatusesUnmarshaler(protoutils.UnmarshalMapToProto)
			inputResource = &mocksv1.MockResource{}
		})

		It("can unmarshal map of statuses", func() {
			resourceStatus := map[string]interface{}{
				"statuses": map[string]*core.Status{
					soloKitNamespace: {
						State:  core.Status_Accepted,
						Reason: "Test reason",
					},
				},
			}

			statusUnmarshaler.UnmarshalStatus(resourceStatus, inputResource)

			statuses := inputResource.GetNamespacedStatuses().GetStatuses()
			Expect(statuses).NotTo(BeEmpty())
			Expect(statuses[soloKitNamespace].GetState()).To(Equal(core.Status_Accepted))
			Expect(statuses[soloKitNamespace].GetReason()).To(Equal("Test reason"))
		})

		It("can unmarshal single status into an empty map", func() {
			resourceStatus := map[string]interface{}{
				soloKitNamespace: &core.Status{
					State:  core.Status_Accepted,
					Reason: "Test reason",
				},
			}
			statusUnmarshaler.UnmarshalStatus(resourceStatus, inputResource)

			Expect(inputResource.GetNamespacedStatuses().GetStatuses()).To(BeEmpty())
		})

		It("can unmarshal an unknown type into an empty map", func() {
			resourceStatus := map[string]interface{}{
				"some-unexpected-key": true,
			}
			statusUnmarshaler.UnmarshalStatus(resourceStatus, inputResource)

			Expect(inputResource.GetNamespacedStatuses().GetStatuses()).To(BeEmpty())
		})
	})

})
