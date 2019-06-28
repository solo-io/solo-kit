package tests

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
)

var _ = Describe("properly creates crds with multiple versions in kube", func() {
	var ()
	BeforeEach(func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			Fail("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		}
	})
	It("properly creates crds with multiple versions in kube", func() {
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())

	})
})
