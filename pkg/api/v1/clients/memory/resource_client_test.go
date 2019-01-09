package memory_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/tests/generic"
)

var _ = Describe("Base", func() {
	var (
		client *ResourceClient
	)
	BeforeEach(func() {
		client = NewResourceClient(NewInMemoryResourceCache(), &v1.MockResource{})
	})
	AfterEach(func() {
	})
	It("CRUDs resources", func() {
		generic.TestCrudClient("", client, time.Minute)
	})
})
