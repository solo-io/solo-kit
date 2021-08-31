package reconcile_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/utils/envutils"
)

var namespace string

var _ = BeforeSuite(func() {
	namespace = "reconcile-suite-test-ns"

	err := os.Setenv(envutils.PodNamespaceEnvName, namespace)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.Unsetenv(envutils.PodNamespaceEnvName)
	Expect(err).NotTo(HaveOccurred())
})

func TestReconcile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reconcile Suite")
}
