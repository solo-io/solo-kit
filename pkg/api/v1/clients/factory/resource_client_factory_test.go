package factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	"k8s.io/client-go/rest"
	"os"

	"context"
	"github.com/solo-io/go-utils/kubeutils"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"log"
	// import k8s client pugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var _ = Describe("ResourceClientFactory", func() {

	Describe("CrdClient when the CRD has not been registered", func() {

		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
			return
		}
		var (
			namespace string
			cfg       *rest.Config
		)
		BeforeEach(func() {
			namespace = helpers.RandString(8)
			err := setup.SetupKubeForTest(namespace)
			Expect(err).NotTo(HaveOccurred())

			cfg, err = kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())

			// Create the CRD in the cluster
			apiExts, err := apiext.NewForConfig(cfg)
			Expect(err).NotTo(HaveOccurred())

			// ensure the crd is not registered
			apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(v1.MockResourceCrd.FullName(), nil)

		})

		Context("and SkipCrdCreation=true", func() {

			It("returns a CrdNotRegistered error", func() {
				factory := KubeResourceClientFactory{
					Crd:             v1.MockResourceCrd,
					Cfg:             cfg,
					SharedCache:     kube.NewKubeCache(context.TODO()),
					SkipCrdCreation: true,
				}
				_, err := factory.NewResourceClient(NewResourceClientParams{
					ResourceType: &v1.MockResource{},
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("list check failed: the server could not find the requested resource (get mocks.crds.testing.solo.io)"))
			})
		})
		Context("and SkipCrdCreation=false", func() {
			It("returns a CrdNotRegistered error", func() {
				factory := KubeResourceClientFactory{
					Crd:         v1.MockResourceCrd,
					Cfg:         cfg,
					SharedCache: kube.NewKubeCache(context.TODO()),
				}
				_, err := factory.NewResourceClient(NewResourceClientParams{
					ResourceType: &v1.MockResource{},
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
