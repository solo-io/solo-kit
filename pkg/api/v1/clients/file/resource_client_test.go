package file_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/test/helpers"

	"io/ioutil"
	"os"

	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/file"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/tests/generic"
)

var _ = Describe("Base", func() {
	var (
		client *ResourceClient
		tmpDir string
	)
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "base_test")
		Expect(err).NotTo(HaveOccurred())
		client = NewResourceClient(tmpDir, &v1.MockResource{})
	})
	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})
	It("CRUDs resources", func() {
		selector := map[string]string{
			helpers.TestLabel: helpers.RandString(8),
		}
		generic.TestCrudClient("ns1", "ns2", client, clients.WatchOpts{
			Selector:    selector,
			Ctx:         context.TODO(),
			RefreshRate: time.Millisecond,
		})
	})
})
