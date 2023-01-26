package shared_test

import (
	"context"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/shared"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("Status", func() {

	var (
		ctx            context.Context
		statusWriterNs string
	)

	BeforeEach(func() {
		ctx = context.Background()

		statusWriterNs = "my-namespace"
		err := os.Setenv(statusutils.PodNamespaceEnvName, statusWriterNs)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Unsetenv(statusutils.PodNamespaceEnvName)
	})

	Context("GetJsonPatchData", func() {

		var inputResource resources.InputResource

		BeforeEach(func() {
			inputResource = &v1.MockResource{
				Metadata: &core.Metadata{
					Name:      "my-resource",
					Namespace: "ns",
				},
				NamespacedStatuses: &core.NamespacedStatuses{
					Statuses: map[string]*core.Status{
						statusWriterNs: {
							State:      2,
							Reason:     "test",
							ReportedBy: "me",
						},
					},
				},
			}
		})

		It("can successfully return status patch", func() {
			data, err := shared.GetJsonPatchData(ctx, inputResource)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).NotTo((BeNil()))
		})

		It("returns error if resource has no statuses", func() {
			inputResource.SetNamespacedStatuses(nil)

			_, err := shared.GetJsonPatchData(ctx, inputResource)
			Expect(err.Error()).To(Equal(shared.NoNamespacedStatusesError(inputResource).Error()))
		})

		It("returns error if POD_NAMESPACE is not set", func() {
			os.Unsetenv(statusutils.PodNamespaceEnvName)

			_, err := shared.GetJsonPatchData(ctx, inputResource)
			Expect(err.Error()).To(Equal(shared.StatusReporterNamespaceError(errors.NewPodNamespaceErr()).Error()))
		})

		It("returns error if no status entry is found for namespace", func() {
			inputResource.SetNamespacedStatuses(&core.NamespacedStatuses{
				Statuses: map[string]*core.Status{
					"another-namespace": {
						State:      2,
						Reason:     "test",
						ReportedBy: "me",
					},
				},
			})

			_, err := shared.GetJsonPatchData(ctx, inputResource)
			Expect(err.Error()).To(Equal(shared.NamespacedStatusNotFoundError(inputResource, statusWriterNs).Error()))
		})

		It("returns error if status is too large", func() {
			var sb strings.Builder
			for i := 0; i < shared.MaxStatusBytes+1; i++ {
				sb.WriteString("a")
			}
			tooLargeReason := sb.String()
			inputResource.GetNamespacedStatuses().GetStatuses()[statusWriterNs].Reason = tooLargeReason

			_, err := shared.GetJsonPatchData(ctx, inputResource)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("patch is too large"))
		})

		It("does not return error if status is too large and max size is disabled", func() {
			shared.DisableMaxStatusSize = true
			var sb strings.Builder
			for i := 0; i < shared.MaxStatusBytes+1; i++ {
				sb.WriteString("a")
			}
			tooLargeReason := sb.String()
			inputResource.GetNamespacedStatuses().GetStatuses()[statusWriterNs].Reason = tooLargeReason

			data, err := shared.GetJsonPatchData(ctx, inputResource)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).NotTo((BeNil()))
		})
	})

})
