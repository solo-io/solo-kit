package vault_test

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/vault"
	"github.com/solo-io/solo-kit/test/helpers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/tests/generic"
)

var _ = Describe("Base", func() {
	var (
		vault   *api.Client
		rootKey string
		secrets clients.ResourceClient
	)
	BeforeEach(func() {
		rootKey = "test-prefix"
		cfg := api.DefaultConfig()
		cfg.Address = fmt.Sprintf("http://127.0.0.1:%v", vaultInstance.Port)
		c, err := api.NewClient(cfg)
		Expect(err).NotTo(HaveOccurred())
		c.SetToken(vaultInstance.Token())
		Expect(err).NotTo(HaveOccurred())
		vault = c
		secrets = NewResourceClient(vault, rootKey, &v1.MockResource{})
	})
	AfterEach(func() {
		vault.Logical().Delete(rootKey)
	})
	It("CRUDs secrets", func() {
		selector := map[string]string{
			helpers.TestLabel: helpers.RandString(8),
		}
		generic.TestCrudClient("ns1", "ns2", secrets, clients.WatchOpts{
			Selector:    selector,
			Ctx:         context.TODO(),
			RefreshRate: time.Second / 8,
		})
	})
})
