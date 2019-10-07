package foo

// DO NOT MERGE -- test package to trial the ergonomics of the multicluster changes

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func bar() {
	ctx := context.Background()

	// Get setup to watch config
	kcw := multicluster.NewKubeConfigWatcher()

	cfg, _ := kubeutils.GetConfig("", "")
	kubeclient, _ := kubernetes.NewForConfig(cfg)
	kubeCache, _ := cache.NewKubeCoreCache(ctx, kubeclient)

	cacheManager := multicluster.NewKubeSharedCacheManager(ctx)
	watchAggregator := wrapper.NewWatchAggregator()
	mockClientSet := multicluster.NewMockResourceClientWithWatchAggregator(cacheManager, watchAggregator)
	restConfigHandler := multicluster.NewRestConfigHandler(kcw, cacheManager, mockClientSet)

	errs, err := restConfigHandler.Run(ctx, cfg, kubeclient, kubeCache)

	rw := wrapper.ResourceWatch(watchAggregator, "foo", nil)
	emitter := v1.NewTestingSimpleEmitter(rw)
	errs, err = v1.NewTestingSimpleEventLoop(emitter, testSyncer{}).Run(ctx)

	mockResource, err := mockClientSet.Read("foo", "bar", clients.ReadOpts{Ctx: ctx, Cluster: "cluster"})
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatal(zap.Error(err))
	}
	mockResource.GetMetadata()

	if err != nil {
		contextutils.LoggerFrom(ctx).Fatal(zap.Error(err))
	}
	for err := range errs {
		contextutils.LoggerFrom(ctx).Fatal(zap.Error(err))
	}
}

type testSyncer struct{}

func (t testSyncer) Sync(ctx context.Context, s *v1.TestingSnapshot) error {
	contextutils.LoggerFrom(ctx).Info(zap.Any("snap", s))
	return nil
}
