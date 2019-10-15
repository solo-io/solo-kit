package multicluster_test

import (
	"context"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/clientgetter"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	multicluster2 "github.com/solo-io/solo-kit/pkg/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("MultiClusterResourceClient integration tests", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}

	var (
		client multicluster.MultiClusterResourceClient
	)

	BeforeEach(func() {
		ctx := context.Background()
		sharedCacheManager := clustercache.NewKubeSharedCacheManager(ctx)
		clientGetter := clientgetter.NewKubeResourceClientGetter(
			sharedCacheManager,
			v2alpha1.MockResourceCrd,
			false,
			nil,
			0,
			factory.NewResourceClientParams{
				ResourceType: &v2alpha1.MockResource{},
			},
		)
		watchAggregator := wrapper.NewWatchAggregator()

		client = multicluster.NewMultiClusterResourceClient(&v2alpha1.MockResource{}, clientGetter, watchAggregator)
		coreCacheManager := clustercache.NewKubeCoreCacheManager(ctx)

		cfgWatcher := multicluster2.NewKubeConfigWatcher()
		cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
		Expect(err).NotTo(HaveOccurred())
		kube, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		handler := multicluster2.NewRestConfigHandler(cfgWatcher, sharedCacheManager, coreCacheManager, client)
		errs, err := handler.Run(ctx, cfg, kube, coreCacheManager.GetCache("", cfg))
		Expect(err).NotTo(HaveOccurred())
		// TODO
		Expect(errs).ShouldNot(Receive())
	})
})
