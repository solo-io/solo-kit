package factory_test

import (
	"context"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/external/kubernetes/configmap"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/customresourcedefinition"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/deployment"
	kubenamespace "github.com/solo-io/solo-kit/pkg/api/external/kubernetes/namespace"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/service"
	client_factory "github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	kubefactory "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/clientfactory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster/factory"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/testutils"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("ClusterClientFactory", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}

	var (
		cacheGetter, awfulCacheGetter clustercache.CacheGetter
		err                           error
	)

	BeforeEach(func() {
		awfulCacheGetter, err = clustercache.NewCacheManager(context.Background(), NewAlwaysWrongCacheForConfig)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("core cache clients", func() {
		BeforeEach(func() {
			cacheGetter, err = clustercache.NewCacheManager(context.Background(), cache.NewCoreCacheForConfig)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("namespace client factory", func() {
			It("works", func() {
				testClientFactory(kubenamespace.NewNamespaceResourceClientFactory(cacheGetter))
				testClientFactoryWithWrongCache(kubenamespace.NewNamespaceResourceClientFactory(awfulCacheGetter))
			})
		})

		Describe("configmap client factory", func() {
			It("works", func() {
				testClientFactory(kubernetes.NewConfigmapResourceClientFactory(cacheGetter, &v1.MockResource{}))
				testClientFactoryWithWrongCache(kubernetes.NewConfigmapResourceClientFactory(awfulCacheGetter, &v1.MockResource{}))
			})
		})

		Describe("pod client factory", func() {
			It("works", func() {
				testClientFactory(pod.NewPodResourceClientFactory(cacheGetter))
				testClientFactoryWithWrongCache(pod.NewPodResourceClientFactory(awfulCacheGetter))
			})
		})

		Describe("service client factory", func() {
			It("works", func() {
				testClientFactory(service.NewServiceResourceClientFactory(cacheGetter))
				testClientFactoryWithWrongCache(service.NewServiceResourceClientFactory(awfulCacheGetter))
			})
		})
	})

	Describe("deployment client factory", func() {
		It("works", func() {
			cacheGetter, err = clustercache.NewCacheManager(context.Background(), cache.NewDeploymentCacheFromConfig)
			Expect(err).NotTo(HaveOccurred())
			testClientFactory(deployment.NewDeploymentResourceClientFactory(cacheGetter))
			testClientFactoryWithWrongCache(deployment.NewDeploymentResourceClientFactory(awfulCacheGetter))
		})
	})

	Describe("customresourcedefinition client factory", func() {
		It("works", func() {
			cacheGetter, err = clustercache.NewCacheManager(context.Background(), customresourcedefinition.NewCrdCacheForConfig)
			Expect(err).NotTo(HaveOccurred())
			testClientFactory(customresourcedefinition.NewCrdResourceClientFactory(cacheGetter))
			testClientFactoryWithWrongCache(customresourcedefinition.NewCrdResourceClientFactory(awfulCacheGetter))
		})
	})

	Describe("crd client factory", func() {
		AfterEach(func() {
			cfg, err := kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())
			clientset, err := apiexts.NewForConfig(cfg)
			Expect(err).NotTo(HaveOccurred())
			err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("mocks.testing.solo.io", &metav1.DeleteOptions{})
			testutils.ErrorNotOccuredOrNotFound(err)
		})

		It("works", func() {
			cacheGetter, err = clustercache.NewCacheManager(context.Background(), kube.NewKubeSharedCacheForConfig)
			Expect(err).NotTo(HaveOccurred())
			testClientFactory(
				kubefactory.NewKubeResourceClientFactory(
					cacheGetter,
					v1.MockResourceCrd,
					false,
					nil,
					0,
					client_factory.NewResourceClientParams{
						ResourceType: &v1.MockResource{},
					},
				),
			)
			testClientFactoryWithWrongCache(
				kubefactory.NewKubeResourceClientFactory(
					awfulCacheGetter,
					v1.MockResourceCrd,
					true,
					nil,
					0,
					client_factory.NewResourceClientParams{
						ResourceType: &v1.MockResource{},
					},
				),
			)
		})
	})
})

func testClientFactory(f factory.ClusterClientFactory) {
	cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	Expect(err).NotTo(HaveOccurred())
	client, err := f.GetClient("", cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(client).NotTo(BeNil())
}

func testClientFactoryWithWrongCache(f factory.ClusterClientFactory) {
	cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	Expect(err).NotTo(HaveOccurred())
	client, err := f.GetClient("", cfg)
	Expect(err).To(HaveOccurred())
	Expect(client).To(BeNil())
}

type alwaysWrongCache struct{}

var _ clustercache.ClusterCache = alwaysWrongCache{}

func (a alwaysWrongCache) IsClusterCache() {}

func NewAlwaysWrongCacheForConfig(ctx context.Context, cluster string, restConfig *rest.Config) clustercache.ClusterCache {
	return alwaysWrongCache{}
}

var _ clustercache.NewClusterCacheForConfig = NewAlwaysWrongCacheForConfig
