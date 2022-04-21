package consul_test

import (
	"context"
	"fmt"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"time"

	"github.com/hashicorp/consul/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/consul"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/generic"
)

var _ = Describe("Base", func() {
	var (
		consul  *api.Client
		client  *ResourceClient
		rootKey string
	)
	BeforeEach(func() {
		rootKey = "my-root-key"

		cfg := api.DefaultConfig()
		cfg.Address = fmt.Sprintf("127.0.0.1:%v", consulInstance.Ports.HttpPort)

		c, err := api.NewClient(cfg)
		Expect(err).NotTo(HaveOccurred())
		consul = c

	})
	AfterEach(func() {
		consul.KV().DeleteTree(rootKey, nil)
	})
	It("CRUDs resources", func() {
		queryOptions := &api.QueryOptions{AllowStale: false, RequireConsistent: true}
		client = NewResourceClient(consul, rootKey, queryOptions, &v1.MockResource{})

		selector := map[string]string{
			helpers.TestLabel: helpers.RandString(8),
		}
		generic.TestCrudClient("ns1", "ns2", client, clients.WatchOpts{
			Selector:    selector,
			Ctx:         context.TODO(),
			RefreshRate: time.Minute,
		})
	})
	It("CRUDs resources when no query options are passed", func() {
		client = NewResourceClient(consul, rootKey, nil, &v1.MockResource{})

		selector := map[string]string{
			helpers.TestLabel: helpers.RandString(8),
		}
		generic.TestCrudClient("ns1", "ns2", client, clients.WatchOpts{
			Selector:    selector,
			Ctx:         context.TODO(),
			RefreshRate: time.Minute,
		})
	})
})
