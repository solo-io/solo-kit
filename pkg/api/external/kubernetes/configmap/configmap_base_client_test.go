package kubernetes

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	kubernetes2 "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/test/helpers"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("Configmap base client", func() {

	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		ctx       context.Context
		namespace string
		client    kubernetes2.ConfigMapClient
		kube      kubernetes.Interface
		kubeCache cache.KubeCoreCache
		cmObj     *kubev1.ConfigMap
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespace = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(ctx, kube, namespace)
		kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client = NewConfigMapClient(kube, kubeCache)
		Expect(err).NotTo(HaveOccurred())
		cmObj, err = kube.CoreV1().ConfigMaps(namespace).Create(ctx, &kubev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      namespace,
			},
		}, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	It("converts a kubernetes pod to solo-kit resource", func() {

		Eventually(func() bool {
			configmaps, err := client.List("", clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
			foundCm := false
			for _, v := range configmaps {
				if v.GetMetadata().Name == cmObj.Name {
					foundCm = true
				}
			}
			return foundCm
		}, time.Minute, time.Second*15).Should(BeTrue())

	})
})
