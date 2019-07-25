package vault_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/test/setup"
)

// TODO: fix tests
func TestVault(t *testing.T) {
	if os.Getenv("RUN_VAULT_TESTS") != "1" {
		log.Printf("This test downloads and runs consul and is disabled by default. To enable, set RUN_VAULT_TESTS=1 in your env.")
		return
	}
	RegisterFailHandler(Fail)
	log.DefaultOut = GinkgoWriter
	RunSpecs(t, "Vault Suite")
}

var (
	vaultFactory  *setup.VaultFactory
	vaultInstance *setup.VaultInstance
	err           error
)

var _ = BeforeSuite(func() {
	vaultFactory, err = setup.NewVaultFactory()
	Expect(err).NotTo(HaveOccurred())
	vaultInstance, err = vaultFactory.NewVaultInstance()
	Expect(err).NotTo(HaveOccurred())
	err = vaultInstance.Run()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	vaultInstance.Clean()
	vaultFactory.Clean()
})
