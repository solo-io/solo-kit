package modutils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
)

var _ = Describe("modutils", func() {
	It("will fail to get the package name if given the incorrect file", func() {
		_, err := GetCurrentModPackageName("fake_file")
		Expect(err).To(HaveOccurred())
		Expect(eris.Is(err, ModPackageFileError)).To(BeTrue())
	})
	It("will function correctly in conjuction with get mod file", func() {
		name, err := GetCurrentModPackageFile()
		Expect(err).NotTo(HaveOccurred())
		pacakgeName, err := GetCurrentModPackageName(name)
		Expect(err).NotTo(HaveOccurred())
		Expect(pacakgeName).To(Equal("github.com/solo-io/solo-kit"))
	})
	It("can list the packages used by this module", func() {
		_, err := GetCurrentPackageList()
		Expect(err).NotTo(HaveOccurred())
	})
})
