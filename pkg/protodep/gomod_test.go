package protodep

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/utils/modutils"
)

var _ = Describe("protodep", func() {
	var (
		modPathString string
		mgr           *goModFactory
	)
	BeforeEach(func() {
		modBytes, err := modutils.GetCurrentModPackageFile()
		modFileString := strings.TrimSpace(modBytes)
		Expect(err).NotTo(HaveOccurred())
		modPathString = filepath.Dir(modFileString)
		mgr, err = NewGoModFactory(modPathString)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("vendor protos", func() {
		It("can vendor protos", func() {
			modules, err := mgr.gather(goModOptions{
				MatchOptions: []*GoModImport{
					GoogleProtoMatcher,
					GogoProtoMatcher,
					ExtProtoMatcher,
					EnvoyValidateProtoMatcher,
				},
				LocalMatchers: []string{"api/**/*.proto"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(modules).To(HaveLen(5))
			Expect(modules[0].ImportPath).To(Equal(EnvoyValidateProtoMatcher.Package))
			Expect(modules[1].ImportPath).To(Equal(GoogleProtoMatcher.Package))
			Expect(modules[2].ImportPath).To(Equal(GogoProtoMatcher.Package))
			Expect(modules[3].ImportPath).To(Equal(ExtProtoMatcher.Package))
			Expect(modules[4].ImportPath).To(Equal("github.com/solo-io/solo-kit"))
			Expect(mgr.copy(modules)).NotTo(HaveOccurred())
		})
	})
})
