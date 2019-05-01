package vault_test

import (
	"context"
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
		rootKey = "/secret/" + helpers.RandString(4)
		cfg := api.DefaultConfig()
		cfg.Address = "http://127.0.0.1:8200"
		c, err := api.NewClient(cfg)
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
		generic.TestCrudClient("", secrets, clients.WatchOpts{
			Selector:    selector,
			Ctx:         context.TODO(),
			RefreshRate: time.Second / 8,
		})
	})
})
