package cli_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/solo-kit/pkg/cli"
)

var _ = Describe("Compile", func() {
	Describe("CreateRequest(dir)", func() {
		It("invokes protoc to return the code generation request", func() {
			dir := "/home/ilackarms/go/src/" +
				"github.com/solo-io/solo-projects/projects/gloo/api/v1"
			req, err := CreateRequest(dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(req).NotTo(BeNil())
		})
	})
})
