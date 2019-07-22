package consul_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/test/setup"
)

func TestConsul(t *testing.T) {
	if os.Getenv("RUN_CONSUL_TESTS") != "1" {
		log.Printf("This test downloads and runs consul and is disabled by default. To enable, set RUN_CONSUL_TESTS=1 in your env.")
		return
	}
	RegisterFailHandler(Fail)
	log.DefaultOut = GinkgoWriter
	RunSpecs(t, "Consul Suite")
}

var (
	consulFactory  *setup.ConsulFactory
	consulInstance *setup.ConsulInstance
	err            error
)

var _ = BeforeSuite(func() {
	consulFactory, err = setup.NewConsulFactory()
	Expect(err).NotTo(HaveOccurred())
	consulInstance, err = consulFactory.NewConsulInstance()
	Expect(err).NotTo(HaveOccurred())
	err = consulInstance.Run()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if consulInstance != nil {
		consulInstance.Clean()
	}
	if consulFactory != nil {
		consulFactory.Clean()
	}
})
