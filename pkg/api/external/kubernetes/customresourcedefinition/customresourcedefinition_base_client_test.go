package customresourcedefinition

import (
	"context"
	"os"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("DeploymentBaseClient", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		client    *customResourceDefinitionResourceClient
		apiExts   apiexts.Interface
		kubeCache KubeCustomResourceDefinitionCache
		crdName   = "testcrds.integrationtests.solokit.solo.io"
	)

	BeforeEach(func() {
		apiExts = helpers.MustApiExtsClient()
		var err error
		kubeCache, err = NewKubeCustomResourceDefinitionCache(context.TODO(), apiExts)
		Expect(err).NotTo(HaveOccurred())
		client = newResourceClient(apiExts, kubeCache)
	})
	AfterEach(func() {
		_ = apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crdName, nil)
	})
	It("converts a kubernetes deployment to solo-kit resource", func() {
		originalKubeCrd, err := apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&v1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: crdName,
			},
			Spec: v1beta1.CustomResourceDefinitionSpec{
				Group:   "integrationtests.solokit.solo.io",
				Version: "v1",
				Scope:   v1beta1.ClusterScoped,
				Names: v1beta1.CustomResourceDefinitionNames{
					Plural:     "testcrds",
					Kind:       "TestCrd",
					ShortNames: []string{"tc"},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		var kubeCrd resources.Resource
		Eventually(func() (resources.Resource, error) {
			kubeCrd, err = client.Read("", crdName, clients.ReadOpts{})
			return kubeCrd, err
		}).Should(Not(BeNil()))
		apiCrd, err := ToKubeCustomResourceDefinition(kubeCrd)
		Expect(err).NotTo(HaveOccurred())
		Expect(apiCrd.Spec.Group).To(Equal(originalKubeCrd.Spec.Group))
	})
})
