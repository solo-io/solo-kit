package wrapper_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
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
		generic.TestCrudClient("test", cluster, time.Minute, generic.Callback{
			PostWriteFunc: func(res resources.Resource) {
				Expect(res.GetMetadata().Cluster).To(Equal(clusterName))
			},
			PostReadFunc: func(res resources.Resource) {
				Expect(res.GetMetadata().Cluster).To(Equal(clusterName))
			},
		})
	})
})
