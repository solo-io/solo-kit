package job

import (
	"context"
	"log"
	"os"

	batchv1 "k8s.io/api/batch/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/test/helpers"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("JobBaseClient", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		ctx       context.Context
		namespace string
		client    *jobResourceClient
		kube      kubernetes.Interface
		kubeCache cache.KubeJobCache
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespace = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(ctx, kube, namespace)
		kubeCache, err = cache.NewKubeJobCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client = newResourceClient(kube, kubeCache)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	It("converts a kubernetes job to solo-kit resource", func() {

		labs := map[string]string{"a": "b"}
		job, err := kube.BatchV1().Jobs(namespace).Create(ctx, &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "happy",
				Namespace: namespace,
			},
			Spec: batchv1.JobSpec{
				Template: kubev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: labs},
					Spec: kubev1.PodSpec{
						RestartPolicy: kubev1.RestartPolicyOnFailure,
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

		var jobs resources.ResourceList
		Eventually(func() (resources.ResourceList, error) {
			jobs, err = client.List(namespace, clients.ListOpts{})
			return jobs, err
		}).Should(HaveLen(1))
		Expect(err).NotTo(HaveOccurred())
		Expect(jobs).To(HaveLen(1))
		Expect(jobs[0].GetMetadata().Name).To(Equal(job.Name))
		Expect(jobs[0].GetMetadata().Namespace).To(Equal(job.Namespace))
		kubeJob, err := ToKubeJob(jobs[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(kubeJob.Spec.Template.Spec.Containers).To(Equal(job.Spec.Template.Spec.Containers))
		Expect(kubeJob.Spec.Template.Spec.RestartPolicy).To(Equal(job.Spec.Template.Spec.RestartPolicy))
	})
})
