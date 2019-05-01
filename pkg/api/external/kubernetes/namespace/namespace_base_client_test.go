package namespace

import (
	"context"
	"log"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/test/helpers"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("Namespace base client", func() {

	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace    string
		client       *namespaceResourceClient
		kube         kubernetes.Interface
		kubeCache    cache.KubeCoreCache
		namespaceObj *kubev1.Namespace
	)

	BeforeEach(func() {
		var err error
		namespace = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client = newResourceClient(kube, kubeCache)
		Expect(err).NotTo(HaveOccurred())
		namespaceObj, err = kube.CoreV1().Namespaces().Create(&kubev1.Namespace{
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
	It("converts a kubernetes namespace to solo-kit resource", func() {

		Eventually(func() bool {
			namespaces, err := client.List("", clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
			foundNamespace := false
			for _, v := range namespaces {
				if v.GetMetadata().Name == namespaceObj.Name {
					foundNamespace = true
				}
			}
			return foundNamespace
		}, time.Minute, time.Second*15).Should(BeTrue())

	})
})
