package pod

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/test/helpers"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("PodBaseClient", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace string
		client    *podResourceClient
		kube      kubernetes.Interface
		kubeCache cache.KubeCoreCache
	)

	BeforeEach(func() {
		namespace = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(kube, namespace)
		kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client = newResourceClient(kube, kubeCache)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	It("converts a kubernetes pod to solo-kit resource", func() {

		pod, err := kube.CoreV1().Pods(namespace).Create(&kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "happy",
				Namespace: namespace,
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					{
						Name:  "nginx",
						Image: "nginx:latest",
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		var pods resources.ResourceList
		Eventually(func() (resources.ResourceList, error) {
			pods, err = client.List(namespace, clients.ListOpts{})
			return pods, err
		}).Should(HaveLen(1))
		Expect(err).NotTo(HaveOccurred())
		Expect(pods).To(HaveLen(1))
		Expect(pods[0].GetMetadata().Name).To(Equal(pod.Name))
		Expect(pods[0].GetMetadata().Namespace).To(Equal(pod.Namespace))
		kubePod, err := ToKubePod(pods[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(kubePod.Spec.Containers).To(Equal(pod.Spec.Containers))
	})
})
