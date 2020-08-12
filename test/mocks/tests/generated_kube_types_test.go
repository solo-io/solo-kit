package tests_test

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	skv1alpha2 "github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1/kube/apis/testing.solo.io/v2alpha1"
	v2alpha1client "github.com/solo-io/solo-kit/test/mocks/v2alpha1/kube/client/clientset/versioned/typed/testing.solo.io/v2alpha1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate bash ../v2alpha1/kube/hack/update-codegen.sh

var _ = Describe("Generated Kube Code", func() {
	var (
		ctx        context.Context
		apiExts    apiext.Interface
		testClient v2alpha1client.TestingV2alpha1Interface
		skClient   skv1alpha2.MockResourceClient
	)

	BeforeEach(func() {
		ctx = context.Background()
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())

		// Create the CRD in the cluster
		apiExts, err = apiext.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())

		err = skv1alpha2.MockResourceCrd.Register(ctx, apiExts)
		Expect(err).NotTo(HaveOccurred())

		testClient, err = v2alpha1client.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())

		skClient, err = skv1alpha2.NewMockResourceClient(ctx, &factory.KubeResourceClientFactory{
			Crd:             skv1alpha2.MockResourceCrd,
			Cfg:             cfg,
			SharedCache:     kube.NewKubeCache(context.TODO()),
			SkipCrdCreation: true,
		})

	})
	AfterEach(func() {
		_ = apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(ctx, skv1alpha2.MockResourceCrd.FullName(), v1.DeleteOptions{})
	})

	It("can read and write a solo kit resource as a typed kube object", func() {
		res := &v2alpha1.MockResource{
			ObjectMeta: v1.ObjectMeta{Name: "foo", Namespace: "default"},
			Spec: skv1alpha2.MockResource{
				WeStuckItInAOneof: &skv1alpha2.MockResource_SomeDumbField{
					SomeDumbField: "we did it",
				},
			},
		}

		out, err := testClient.MockResources(res.Namespace).Create(ctx, res, v1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
		out.Spec.Metadata = core.Metadata{}
		Expect(out.Spec).To(Equal(res.Spec))

		skOut, err := skClient.Read(res.Namespace, res.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		Expect(skOut.WeStuckItInAOneof).To(Equal(out.Spec.WeStuckItInAOneof))
	})
})
