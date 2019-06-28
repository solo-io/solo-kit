package factory_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/solo-io/go-utils/log"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"context"

	"github.com/solo-io/go-utils/kubeutils"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	// import k8s client pugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type BeErrTypeMatcher struct {
	ExpectedErrType string
	IsErrType       func(error) bool
}

func (matcher BeErrTypeMatcher) Match(actual interface{}) (bool, error) {
	err, ok := actual.(error)
	if !ok {
		return false, fmt.Errorf("Expected a boolean.  Got:\n%s", format.Object(actual, 1))
	}

	return matcher.IsErrType(err), nil
}

func (matcher BeErrTypeMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be %v", matcher.ExpectedErrType)
}

func (matcher BeErrTypeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to be %v", matcher.ExpectedErrType)
}

var _ = Describe("ResourceClientFactory", func() {

	Describe("CrdClient when the CRD has not been registered", func() {

		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
			return
		}
		var (
			cfg *rest.Config
		)
		BeforeEach(func() {
			var err error
			cfg, err = kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())

			// Create the CRD in the cluster
			apiExts, err := apiext.NewForConfig(cfg)
			Expect(err).NotTo(HaveOccurred())

			// ensure the crd is not registered
			err = apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(v1.MockResourceCrd.FullName(), nil)
			if err != nil {
				Expect(err).To(BeErrTypeMatcher{
					ExpectedErrType: "not found",
					IsErrType:       errors.IsNotFound,
				})
			}
			Eventually(func() bool {
				_, err := apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Get(v1.MockResourceCrd.FullName(), v12.GetOptions{})
				return err != nil && errors.IsNotFound(err)
			}, time.Minute, time.Second*5).Should(BeTrue())

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
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("list check failed: the server could not find the requested resource (get %s)", v1.MockResourceCrd.FullName())))
			})
		})
		Context("and SkipCrdCreation=false", func() {
			It("does not error", func() {
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
