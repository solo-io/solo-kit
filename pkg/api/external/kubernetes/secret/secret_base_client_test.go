package secret

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	kubernetes2 "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/test/helpers"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("secret base client", func() {

	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace string
		client    kubernetes2.SecretClient
		kube      kubernetes.Interface
		kubeCache cache.KubeCoreCache
		secretObj *kubev1.Secret
	)

	BeforeEach(func() {
		namespace = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(kube, namespace)
		kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client = NewSecretClient(kube, kubeCache)
		Expect(err).NotTo(HaveOccurred())
		secretObj, err = kube.CoreV1().Secrets(namespace).Create(&kubev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      namespace,
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	It("converts a kubernetes pod to solo-kit resource", func() {

		Eventually(func() bool {
			secrets, err := client.List("", clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
			foundSecret := false
			for _, v := range secrets {
				if v.GetMetadata().Name == secretObj.Name {
					foundSecret = true
				}
			}
			return foundSecret
		}, time.Minute, time.Second*15).Should(BeTrue())

	})
})
