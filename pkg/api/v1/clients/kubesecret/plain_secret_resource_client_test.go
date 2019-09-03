package kubesecret_test

import (
	"context"
	"os"
	"time"

	kubehelpers "github.com/solo-io/go-utils/testutils/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/solo-io/solo-kit/test/mocks/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/log"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/generic"
	"k8s.io/client-go/kubernetes"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("Kube Secret Client Plain=True", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		ns1, ns2 string
		client   *ResourceClient
		kube     kubernetes.Interface
	)
	BeforeEach(func() {
		ns1 = helpers.RandString(8)
		ns2 = helpers.RandString(8)
		kube = helpers.MustKubeClient()

		err := kubeutils.CreateNamespacesInParallel(kube, ns1, ns2)
		Expect(err).NotTo(HaveOccurred())
		kcache, err := cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client, err = NewResourceClient(kube, &v1.MockResource{}, true, kcache)
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kube, ns1, ns2)
		Expect(err).NotTo(HaveOccurred())

		kubehelpers.WaitForNamespaceTeardown(ns1)
		kubehelpers.WaitForNamespaceTeardown(ns2)
	})
	It("CRUDs resources", func() {
		selectors := map[string]string{
			helpers.TestLabel: helpers.RandString(8),
		}
		generic.TestCrudClient(ns1, ns2, client, clients.WatchOpts{
			RefreshRate: time.Minute,
			Selector:    selectors,
			Ctx:         context.TODO(),
		})
	})
	It("does not escape string fields", func() {
		foo := "test-data-keys"
		input := v1.NewMockResource(ns1, foo)
		data := "hello: goodbye"
		input.Data = data
		labels := map[string]string{"pick": "me"}
		input.Metadata.Labels = labels

		err := client.Register()
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Write(input, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		cm, err := kube.CoreV1().Secrets(input.Metadata.Namespace).Get(input.Metadata.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(cm.Data).To(HaveKey("data.json"))
		Expect(string(cm.Data["data.json"])).To(Equal("hello: goodbye"))
	})
	It("emits empty fields", func() {
		foo := "test-data-keys"
		input := v1.NewMockResource(ns1, foo)
		data := ""
		input.Data = data
		labels := map[string]string{"pick": "me"}
		input.Metadata.Labels = labels

		err := client.Register()
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Write(input, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		cm, err := kube.CoreV1().Secrets(input.Metadata.Namespace).Get(input.Metadata.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(cm.Data).To(HaveKey("data.json"))
	})
	// no string escape
})
