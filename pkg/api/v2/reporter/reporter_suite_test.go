package reporter_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"
)

var namespace string

var _ = BeforeSuite(func() {
	namespace = "v2-reporter-suite-test-ns"

	err := os.Setenv(statusutils.PodNamespaceEnvName, namespace)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.Unsetenv(statusutils.PodNamespaceEnvName)
	Expect(err).NotTo(HaveOccurred())
})

func TestReporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reporter Suite")
}
