package factory_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"context"

	"github.com/solo-io/solo-kit/test/kubeutils"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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
			fmt.Print("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
			return
		}
		var (
			ctx context.Context
			cfg *rest.Config
		)
		BeforeEach(func() {
			ctx = context.Background()
			var err error
			cfg, err = kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())

			// Create the CRD in the cluster
			apiExts, err := apiext.NewForConfig(cfg)
			Expect(err).NotTo(HaveOccurred())

			// ensure the crd is not registered
			err = apiExts.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, v1.MockResourceCrd.FullName(), v12.DeleteOptions{})
			if err != nil {
				Expect(err).To(BeErrTypeMatcher{
					ExpectedErrType: "not found",
					IsErrType:       errors.IsNotFound,
				})
			}
			Eventually(func() bool {
				_, err := apiExts.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, v1.MockResourceCrd.FullName(), v12.GetOptions{})
				return err != nil && errors.IsNotFound(err)
			}, time.Minute, time.Second*5).Should(BeTrue())

		})
	})
})
