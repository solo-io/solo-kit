package v1

import (
	"context"
	"os"
	"time"

	"github.com/solo-io/solo-kit/pkg/utils/statusutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/test/helpers"

	// Needed to run tests in GKE
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"

	// From https://github.com/kubernetes/client-go/blob/53c7adfd0294caa142d961e1f780f74081d5b15f/examples/out-of-cluster-client-configuration/main.go#L31
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("V1Emitter", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		ctx                       context.Context
		namespace1                string
		slowWatchNamespace        string
		name1                     = "angela" + helpers.RandString(3)
		cfg                       *rest.Config
		clientset                 *apiext.Clientset
		kube                      kubernetes.Interface
		emitter                   TestingEmitter
		simpleMockResourceClient  SimpleMockResourceClient
		mockResourceClient        MockResourceClient
		fakeResourceClient        FakeResourceClient
		anotherMockResourceClient AnotherMockResourceClient
		clusterResourceClient     ClusterResourceClient
		mockCustomTypeClient      MockCustomTypeClient
		podClient                 github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodClient
	)

	BeforeEach(func() {
		err := os.Setenv(statusutils.PodNamespaceEnvName, "default")
		Expect(err).NotTo(HaveOccurred())

		ctx = context.Background()
		namespace1 = helpers.RandString(8)
		slowWatchNamespace = "slow-watch-namespace"

		kube = helpers.MustKubeClient()
		err = kubeutils.CreateNamespacesInParallel(ctx, kube, namespace1)
		Expect(err).NotTo(HaveOccurred())
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())

		clientset, err = apiext.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		// SimpleMockResource Constructor
		simpleMockResourceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}

		simpleMockResourceClient, err = NewSimpleMockResourceClient(ctx, simpleMockResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())
		// MockResource Constructor
		mockResourceClientFactory := &factory.KubeResourceClientFactory{
			Crd:         MockResourceCrd,
			Cfg:         cfg,
			SharedCache: kuberc.NewKubeCache(context.TODO()),
		}

		err = helpers.AddAndRegisterCrd(ctx, MockResourceCrd, clientset)
		Expect(err).NotTo(HaveOccurred())

		mockResourceClient, err = NewMockResourceClient(ctx, mockResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())
		// FakeResource Constructor
		fakeResourceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}

		fakeResourceClient, err = NewFakeResourceClient(ctx, fakeResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())
		// AnotherMockResource Constructor
		anotherMockResourceClientFactory := &factory.KubeResourceClientFactory{
			Crd:         AnotherMockResourceCrd,
			Cfg:         cfg,
			SharedCache: kuberc.NewKubeCache(context.TODO()),
		}

		err = helpers.AddAndRegisterCrd(ctx, AnotherMockResourceCrd, clientset)
		Expect(err).NotTo(HaveOccurred())

		anotherMockResourceClient, err = NewAnotherMockResourceClient(ctx, anotherMockResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())
		// ClusterResource Constructor
		clusterResourceClientFactory := &factory.KubeResourceClientFactory{
			Crd:         ClusterResourceCrd,
			Cfg:         cfg,
			SharedCache: kuberc.NewKubeCache(context.TODO()),
		}

		err = helpers.AddAndRegisterCrd(ctx, ClusterResourceCrd, clientset)
		Expect(err).NotTo(HaveOccurred())

		clusterResourceClient, err = NewClusterResourceClient(ctx, clusterResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())
		// MockCustomType Constructor
		mockCustomTypeClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}

		mockCustomTypeClient, err = NewMockCustomTypeClient(ctx, mockCustomTypeClientFactory)
		Expect(err).NotTo(HaveOccurred())
		// Pod Constructor
		podClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}

		podClient, err = github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPodClient(ctx, podClientFactory)
		Expect(err).NotTo(HaveOccurred())
		emitter = NewTestingEmitter(simpleMockResourceClient, mockResourceClient, fakeResourceClient, anotherMockResourceClient, clusterResourceClient, mockCustomTypeClient, podClient)
	})
	AfterEach(func() {
		kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespace1, slowWatchNamespace)
	})

	It("Should not overwrite initial listed resources for non-returned namespace watchers", func() {
		/*
			This test relies on some hardcoded logic in `resource_client_template.go`.  In particular:
			`FakeResourceClient`, when used in conjunction with the `slowWatchNamespace` has an artificial
			5s delay when getting back the results from .Watch.  The idea behind the test is that
			we have a normal, unimpeded watch running in parallel with a slow watch to simulate a race condition.
		*/
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		// start by creating the test-specific namespace

		err = kubeutils.CreateNamespacesInParallel(ctx, kube, slowWatchNamespace)
		Expect(err).NotTo(HaveOccurred())

		// then, write `FakeResource`s to "namespace1" and "slowWatchNamespace"
		fakeResource1a, err := fakeResourceClient.Write(NewFakeResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		fakeResource1b, err := fakeResourceClient.Write(NewFakeResource(slowWatchNamespace, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		// create a snapshot observer on "namespace1" and "slowWatchNamespace"
		snapshots, _, err := emitter.Snapshots([]string{namespace1, slowWatchNamespace}, clients.WatchOpts{
			Ctx:         ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())

		timeout := time.After(time.Second * 10)
		ticker := time.Tick(time.Second / 10)
	outer:
		for {
			select {
			case <-ctx.Done():
				break outer
			case <-timeout:
				// clean up created resources
				err = fakeResourceClient.Delete(fakeResource1a.GetMetadata().Namespace, fakeResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				err = fakeResourceClient.Delete(fakeResource1b.GetMetadata().Namespace, fakeResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				err = kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, slowWatchNamespace)
				Expect(err).NotTo(HaveOccurred())
				break outer
			case <-ticker:
				snap := <-snapshots
				// 2 Fakes is optimal (everything is found)
				// 1 Fakes represents the case of a watcher overwriting the initially populated 2 Fakes from .List calls
				// 0 Fakes represents the observer not yet having produced a snapshot
				Expect(len(snap.Fakes)).NotTo(Equal(1))
			}
		}
	})
})
