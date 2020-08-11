package deployment

import (
	"context"
	"os"

	v1 "k8s.io/api/apps/v1"

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

var _ = Describe("DeploymentBaseClient", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		ctx       context.Context
		namespace string
		client    *deploymentResourceClient
		kube      kubernetes.Interface
		kubeCache cache.KubeDeploymentCache
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespace = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(ctx, kube, namespace)
		kubeCache, err = cache.NewKubeDeploymentCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client = newResourceClient(kube, kubeCache)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	It("converts a kubernetes deployment to solo-kit resource", func() {

		labs := map[string]string{"a": "b"}
		deployment, err := kube.AppsV1().Deployments(namespace).Create(ctx, &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "happy",
				Namespace: namespace,
			},
			Spec: v1.DeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: labs},
				Template: kubev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: labs},
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:latest",
							},
						},
					},
				},
			},
		}, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		var deployments resources.ResourceList
		Eventually(func() (resources.ResourceList, error) {
			deployments, err = client.List(namespace, clients.ListOpts{})
			return deployments, err
		}).Should(HaveLen(1))
		Expect(err).NotTo(HaveOccurred())
		Expect(deployments).To(HaveLen(1))
		Expect(deployments[0].GetMetadata().Name).To(Equal(deployment.Name))
		Expect(deployments[0].GetMetadata().Namespace).To(Equal(deployment.Namespace))
		kubeDeployment, err := ToKubeDeployment(deployments[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(kubeDeployment.Spec.Template.Spec.Containers).To(Equal(deployment.Spec.Template.Spec.Containers))
	})
})
