package kubesecret_test

import (
	"context"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"os"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/solo-io/solo-kit/test/mocks/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/solo-kit/test/tests/generic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("Kube Secret Client Plain=True", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace string
		cfg       *rest.Config
		client    *ResourceClient
		kube      kubernetes.Interface
	)
	BeforeEach(func() {
		namespace = helpers.RandString(8)
		err := setup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		kube, err = kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		client, err = NewResourceClient(kube, &v1.MockResource{}, true, cache.NewKubeCoreCache(context.TODO(), kube))
	})
	AfterEach(func() {
		setup.TeardownKube(namespace)
	})
	It("CRUDs resources", func() {
		generic.TestCrudClient(namespace, client, time.Minute)
	})
	It("does not escape string fields", func() {
		foo := "test-data-keys"
		input := v1.NewMockResource(namespace, foo)
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
		input := v1.NewMockResource(namespace, foo)
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
