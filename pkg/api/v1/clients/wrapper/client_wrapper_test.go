package wrapper_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/test/helpers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/tests/generic"
)

var _ = Describe("ClientWrapper", func() {
	var cluster *Client
	clusterName := "clustr"
	BeforeEach(func() {
		base := memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
		cluster = NewClusterClient(base, clusterName)
	})
	It("applies the process func to resources upon CRUD", func() {
		opts := clients.WatchOpts{
			RefreshRate: time.Minute,
			Ctx:         context.TODO(),
			Selector: map[string]string{
				helpers.TestLabel: helpers.RandString(8),
			},
		}
		generic.TestCrudClient("test", cluster, opts, generic.Callback{
			PostWriteFunc: func(res resources.Resource) {
				Expect(res.GetMetadata().Cluster).To(Equal(clusterName))
			},
			PostReadFunc: func(res resources.Resource) {
				Expect(res.GetMetadata().Cluster).To(Equal(clusterName))
			},
		})
	})
})
