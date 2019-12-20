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
		mgr           *manager
	)
	BeforeEach(func() {
		modBytes, err := modutils.GetCurrentModPackageFile()
		modFileString := strings.TrimSpace(modBytes)
		Expect(err).NotTo(HaveOccurred())
		modPathString = filepath.Dir(modFileString)
		mgr = NewManager()
	})

	Context("vendor protos", func() {
		It("can vendor protos", func() {
			modules, err := mgr.Gather(Options{
				WorkingDirectory: modPathString,
				MatchPattern:     "",
				IncludePackages: []string{
					"github.com/solo-io/protoc-gen-ext",
					"github.com/envoyproxy/protoc-gen-validate",
				},
			})
			Expect(modules).To(HaveLen(2))
			Expect(err).NotTo(HaveOccurred())
			Expect(mgr.Copy(modules)).NotTo(HaveOccurred())
		})
	})
})
