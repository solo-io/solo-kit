package v1beta1_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen/v1beta1"
)

var _ = Describe("Parser", func() {

	Context("GetCRDFromFile", func() {

		It("can parse YAML", func() {
			crd, err := v1beta1.GetCRDFromFile(getPathToSourceCrdFile("cc.yaml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(crd.Name).To(Equal("crd.test.gloo.solo.io"))
			Expect(crd.Spec.Names.Singular).To(Equal("customconfig"))
		})

		It("errors if CRD.ApiVersion does not match v1beta1", func() {
			_, err := v1beta1.GetCRDFromFile(getPathToSourceCrdFile("v1.yaml"))
			Expect(err).To(MatchError(v1beta1.ApiVersionMismatch("apiextensions.k8s.io/v1beta1", "apiextensions.k8s.io/v1")))
		})

	})

})

func getPathToSourceCrdFile(fileName string) string {
	return fmt.Sprintf("fixtures/source/%s", fileName)
}
