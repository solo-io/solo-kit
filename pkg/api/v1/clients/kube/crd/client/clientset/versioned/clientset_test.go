package versioned_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/go-utils/kubeutils"
	crdv1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/test/helpers"
	mocksv1 "github.com/solo-io/solo-kit/test/mocks/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("Clientset", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace string
		cfg       *rest.Config
		kube      kubernetes.Interface
	)
	BeforeEach(func() {
		namespace = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(kube, namespace)
		Expect(err).NotTo(HaveOccurred())
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	It("registers, creates, deletes resource implementations", func() {
		apiextsClient, err := apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		err = mocksv1.MockResourceCrd.Register(apiextsClient)
		Expect(err).NotTo(HaveOccurred())

		c, err := apiextsClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(c.Items)).To(BeNumerically(">=", 1))
		var found bool
		for _, i := range c.Items {
			if i.Name == mocksv1.MockResourceCrd.FullName() {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())

		mockCrdClient, err := NewForConfig(cfg, mocksv1.MockResourceCrd)
		Expect(err).NotTo(HaveOccurred())
		name := "foo"
		input := mocksv1.NewMockResource(namespace, name)
		input.Data = name
		inputCrd := mocksv1.MockResourceCrd.KubeResource(input)
		created, err := mockCrdClient.ResourcesV1().Resources(namespace).Create(inputCrd)
		Expect(err).NotTo(HaveOccurred())
		Expect(created).NotTo(BeNil())
		Expect(created.Spec).NotTo(BeNil())
		Expect(created.Spec).To(Equal(&crdv1.Spec{
			"data":     "foo",
			"metadata": map[string]interface{}{"name": "foo", "namespace": namespace},
			"status":   map[string]interface{}{},
		}))
	})
})
