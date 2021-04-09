package v1beta1_test

import (
	"fmt"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

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

	Context("WriteCRDSpecToFile", func() {

		It("can write YAML", func() {
			// Read the source crd
			sourceCrd, err := v1beta1.GetCRDFromFile(getPathToSourceCrdFile("cc.yaml"))
			Expect(err).NotTo(HaveOccurred())

			// Write the crd to a file
			err = v1beta1.WriteCRDSpecToFile(sourceCrd, getPathToGeneratedCrdFile("cc.gen.yaml"))
			Expect(err).NotTo(HaveOccurred())

			generatedCrd, err := v1beta1.GetCRDFromFile(getPathToGeneratedCrdFile("cc.gen.yaml"))
			Expect(err).NotTo(HaveOccurred())

			// The generated crd should match the original
			Expect(generatedCrd).To(Equal(sourceCrd))
		})
	})

	Context("WriteCRDListToFile", func() {

		It("can write YAML", func() {
			crd, err := v1beta1.GetCRDFromFile(getPathToSourceCrdFile("cc.yaml"))
			Expect(err).NotTo(HaveOccurred())

			err = v1beta1.WriteCRDListToFile(
				[]apiextv1beta1.CustomResourceDefinition{crd, crd},
				getPathToGeneratedCrdFile("manifest-list.gen.yaml"))
			Expect(err).NotTo(HaveOccurred())
		})

	})

})

func getPathToSourceCrdFile(fileName string) string {
	return fmt.Sprintf("test_utils/source/%s", fileName)
}

func getPathToGeneratedCrdFile(fileName string) string {
	return fmt.Sprintf("test_utils/generated/%s", fileName)
}
