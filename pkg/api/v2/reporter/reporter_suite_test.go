package reporter_test

import (
	"os"
	"testing"

	"github.com/solo-io/solo-kit/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var namespace string

var _ = BeforeSuite(func() {
	namespace = helpers.RandString(5)

	err := os.Setenv("POD_NAMESPACE", namespace)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.Unsetenv("POD_NAMESPACE")
	Expect(err).NotTo(HaveOccurred())
})

func TestReporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reporter Suite")
}
