package reconcile_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var namespace string

var _ = BeforeSuite(func() {
	namespace = "reconcile-suite-test-ns"

	err := os.Setenv("POD_NAMESPACE", namespace)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.Unsetenv("POD_NAMESPACE")
	Expect(err).NotTo(HaveOccurred())
})

func TestReconcile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reconcile Suite")
}
