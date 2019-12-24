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
		mgr, err = NewManager(modPathString)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("vendor protos", func() {
		It("can vendor protos", func() {
			modules, err := mgr.Gather([]MatchOptions{
				GogoProtoMatcher,
				ExtProtoMatcher,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(modules).To(HaveLen(2))
			Expect(mgr.Copy(modules)).NotTo(HaveOccurred())
		})
	})
})
