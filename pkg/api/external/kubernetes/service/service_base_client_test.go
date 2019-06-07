package service

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

var _ = Describe("ServiceBaseClient", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace string
		client    *serviceResourceClient
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
	It("converts a kubernetes service to solo-kit resource", func() {

		service, err := kube.CoreV1().Services(namespace).Create(&kubev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "happy",
				Namespace: namespace,
			},
			Spec: kubev1.ServiceSpec{
				Ports: []kubev1.ServicePort{
					{
						Name: "grpc",
						Port: 8080,
					},
				},
				Selector: map[string]string{"foo": "bar"},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		var services resources.ResourceList
		Eventually(func() (resources.ResourceList, error) {
			services, err = client.List(namespace, clients.ListOpts{})
			return services, err
		}).Should(HaveLen(1))
		Expect(err).NotTo(HaveOccurred())
		Expect(services).To(HaveLen(1))
		Expect(services[0].GetMetadata().Name).To(Equal(service.Name))
		Expect(services[0].GetMetadata().Namespace).To(Equal(service.Namespace))
		kubeService, err := ToKubeService(services[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(kubeService.Spec.Ports).To(Equal(service.Spec.Ports))
		Expect(kubeService.Spec.Selector).To(Equal(service.Spec.Selector))
	})
})
