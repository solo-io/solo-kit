package file_test

import (
	"context"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/test/helpers"

	"io/ioutil"

	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/file"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/tests/generic"
)

var _ = Describe("Base", func() {

	var (
		client     *ResourceClient
		tmpDir     string
		namespace1 = "ns1"
		namespace2 = "ns2"
	)

	JustBeforeEach(func() {
		var err error

		tmpDir, err = ioutil.TempDir("", "base_test")
		Expect(err).NotTo(HaveOccurred())

		// Create a directory per namespace inside the tmpDir.
		// These will be cleaned up when the tmpDir is cleaned up
		// The underlying TestCrudClient relies on these folders existing
		err = os.Mkdir(filepath.Join(tmpDir, namespace1), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
		err = os.Mkdir(filepath.Join(tmpDir, namespace2), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		client = NewResourceClient(tmpDir, &v1.MockResource{})
	})

	JustAfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	It("CRUDs resources", func() {
		selector := map[string]string{
			helpers.TestLabel: helpers.RandString(8),
		}
		generic.TestCrudClient(namespace1, namespace2, client, clients.WatchOpts{
			Selector:    selector,
			Ctx:         context.TODO(),
			RefreshRate: time.Millisecond,
		})
	})
})
