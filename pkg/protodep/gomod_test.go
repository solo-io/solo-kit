package protodep

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/protodep/api"
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
				MatchOptions: []*api.GoModImport{
					GogoProtoMatcher,
					ExtProtoMatcher,
					ValidateProtoMatcher,
				},
				LocalMatchers: []string{"api/**/*.proto"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(modules).To(HaveLen(4))
			Expect(modules[0].ImportPath).To(Equal(ValidateProtoMatcher.Package))
			Expect(modules[1].ImportPath).To(Equal(GogoProtoMatcher.Package))
			Expect(modules[2].ImportPath).To(Equal(ExtProtoMatcher.Package))
			Expect(modules[3].ImportPath).To(Equal("github.com/solo-io/solo-kit"))
			Expect(mgr.copy(modules)).NotTo(HaveOccurred())
		})
	})
})
