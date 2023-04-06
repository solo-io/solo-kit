package sk_anyvendor

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("import abstraction", func() {
	It("can properly translate Imports to anyvendor.Config", func() {
		imports := CreateDefaultMatchOptions(DefaultMatchPatterns)
		Expect(imports.Local).To(Equal(DefaultMatchPatterns))
		Expect(imports.External).To(Equal(DefaultExternalMatchOptions))
	})
})
