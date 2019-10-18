package multicluster_test

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
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/namespace"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/service"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/clientgetter"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/testutils"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("ClientGetter", func() {
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

		Describe("namespace client getter", func() {
			It("works", func() {
				testClientGetter(namespace.NewNamespaceResourceClientGetter(cacheGetter))
				testClientGetterWithWrongCache(namespace.NewNamespaceResourceClientGetter(awfulCacheGetter))
			})
		})

		Describe("configmap client getter", func() {
			It("works", func() {
				testClientGetter(kubernetes.NewConfigmapResourceClientGetter(cacheGetter, &v1.MockResource{}))
				testClientGetterWithWrongCache(kubernetes.NewConfigmapResourceClientGetter(awfulCacheGetter, &v1.MockResource{}))
			})
		})

		Describe("pod client getter", func() {
			It("works", func() {
				testClientGetter(pod.NewPodResourceClientGetter(cacheGetter))
				testClientGetterWithWrongCache(pod.NewPodResourceClientGetter(awfulCacheGetter))
			})
		})

		Describe("service client getter", func() {
			It("works", func() {
				testClientGetter(service.NewServiceResourceClientGetter(cacheGetter))
				testClientGetterWithWrongCache(service.NewServiceResourceClientGetter(awfulCacheGetter))
			})
		})
	})

	Describe("deployment client getter", func() {
		It("works", func() {
			cacheGetter, err = clustercache.NewCacheManager(context.Background(), cache.NewDeploymentCacheFromConfig)
			Expect(err).NotTo(HaveOccurred())
			testClientGetter(deployment.NewDeploymentResourceClientGetter(cacheGetter))
			testClientGetterWithWrongCache(deployment.NewDeploymentResourceClientGetter(awfulCacheGetter))
		})
	})

	Describe("customresourcedefinition client getter", func() {
		It("works", func() {
			cacheGetter, err = clustercache.NewCacheManager(context.Background(), customresourcedefinition.NewCrdCacheForConfig)
			Expect(err).NotTo(HaveOccurred())
			testClientGetter(customresourcedefinition.NewCrdResourceClientGetter(cacheGetter))
			testClientGetterWithWrongCache(customresourcedefinition.NewCrdResourceClientGetter(awfulCacheGetter))
		})
	})

	Describe("crd client getter", func() {
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
			testClientGetter(
				clientgetter.NewKubeResourceClientGetter(
					cacheGetter,
					v1.MockResourceCrd,
					false,
					nil,
					0,
					factory.NewResourceClientParams{
						ResourceType: &v1.MockResource{},
					},
				),
			)
			testClientGetterWithWrongCache(
				clientgetter.NewKubeResourceClientGetter(
					awfulCacheGetter,
					v1.MockResourceCrd,
					true,
					nil,
					0,
					factory.NewResourceClientParams{
						ResourceType: &v1.MockResource{},
					},
				),
			)
		})
	})
})

func testClientGetter(getter multicluster.ClientGetter) {
	cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	Expect(err).NotTo(HaveOccurred())
	client, err := getter.GetClient("", cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(client).NotTo(BeNil())
}

func testClientGetterWithWrongCache(getter multicluster.ClientGetter) {
	cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	Expect(err).NotTo(HaveOccurred())
	client, err := getter.GetClient("", cfg)
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
