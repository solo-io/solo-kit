package customresourcedefinition

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/log"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		ctx       context.Context
		client    *customResourceDefinitionResourceClient
		apiExts   apiexts.Interface
		kubeCache KubeCustomResourceDefinitionCache
		crdName   = "testcrds.integrationtests.solokit.solo.io"
	)

	BeforeEach(func() {
		ctx = context.Background()
		apiExts = helpers.MustApiExtsClient()
		var err error
		kubeCache, err = NewKubeCustomResourceDefinitionCache(context.TODO(), apiExts)
		Expect(err).NotTo(HaveOccurred())
		client = newResourceClient(apiExts, kubeCache)
	})
	AfterEach(func() {
		_ = apiExts.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, crdName, metav1.DeleteOptions{})
	})
	It("converts a kubernetes deployment to solo-kit resource", func() {
		originalKubeCrd, err := apiExts.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, &v1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: crdName,
			},
			Spec: v1.CustomResourceDefinitionSpec{
				Group: "integrationtests.solokit.solo.io",
				Scope: v1.ClusterScoped,
				Names: v1.CustomResourceDefinitionNames{
					Plural:     "testcrds",
					Kind:       "TestCrd",
					ShortNames: []string{"tc"},
				},
				Versions: []v1.CustomResourceDefinitionVersion{
					{
						Name: "v1",
						Schema: &v1.CustomResourceValidation{
							OpenAPIV3Schema: &v1.JSONSchemaProps{
								Type: "object",
							},
						},
						Storage: true,
						Served:  true,
					},
				},
			},
		}, metav1.CreateOptions{})
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
