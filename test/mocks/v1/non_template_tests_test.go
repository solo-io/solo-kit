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
)

var _ FakeResourceClient = new(slowFakeResourceClient)

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

		fakeResource1a *FakeResource
		fakeResource1b *FakeResource
	)

	BeforeEach(func() {
		err := os.Setenv(statusutils.PodNamespaceEnvName, "default")
		Expect(err).NotTo(HaveOccurred())

		ctx = context.Background()
		namespace1 = helpers.RandString(8)
		slowWatchNamespace = "slow-watch-namespace"

		kube = helpers.MustKubeClient()
		err = kubeutils.CreateNamespacesInParallel(ctx, kube, namespace1, slowWatchNamespace)
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

		// We create a resource client which has built in latency
		originalFakeResourceClient, err := NewFakeResourceClient(ctx, fakeResourceClientFactory)
		Expect(err).NotTo(HaveOccurred())
		fakeResourceClient := NewSlowFakeResourceClient(originalFakeResourceClient, map[string]time.Duration{
			slowWatchNamespace: time.Second * 5,
		})

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

		// create `FakeResource`s in "namespace1" and "slowWatchNamespace"
		fakeResource1a, err = fakeResourceClient.Write(NewFakeResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		fakeResource1b, err = fakeResourceClient.Write(NewFakeResource(slowWatchNamespace, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespace1, slowWatchNamespace)

		// clean up created resources
		err := fakeResourceClient.Delete(fakeResource1a.GetMetadata().Namespace, fakeResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = fakeResourceClient.Delete(fakeResource1b.GetMetadata().Namespace, fakeResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should not overwrite initial listed resources for non-returned namespace watchers", func() {
		/*
			This test makes use of the below-defined `NewSlowFakeResourceClient`, which itself is a wrapper around `FakeResourceClient`.
			The idea is that we have a designated namespace (slowWatchNamespace), that has a 5s delay when calling .Watch.  With this
			in place, we can simulate a race condition between a normal, unimpeded namespace (namespace1) and a slower namespace (slowWatchNamespace).
		*/
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		// create a snapshot observer on "namespace1" and "slowWatchNamespace"
		snapshots, _, err := emitter.Snapshots([]string{namespace1, slowWatchNamespace}, clients.WatchOpts{
			Ctx:         ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())

		timeout := time.After(time.Second * 10)
	outer:
		for {
			select {
			case <-ctx.Done():
				break outer
			case <-timeout:
				break outer
			case snap := <-snapshots:
				// 2 Fakes is optimal (everything is found)
				// 1 Fakes represents the case of a watcher overwriting the initially populated 2 Fakes from .List calls
				// 0 Fakes represents the observer not yet having produced a snapshot
				Expect(len(snap.Fakes)).NotTo(Equal(1))
			}
		}
	})
})

type slowFakeResourceClient struct {
	FakeResourceClient
	latencyPerNamespace map[string]time.Duration
}

func NewSlowFakeResourceClient(client FakeResourceClient, latencyPerNs map[string]time.Duration) *slowFakeResourceClient {
	return &slowFakeResourceClient{
		FakeResourceClient:  client,
		latencyPerNamespace: latencyPerNs,
	}
}

func (s *slowFakeResourceClient) BaseClient() clients.ResourceClient {
	return s.FakeResourceClient.BaseClient()
}

func (s *slowFakeResourceClient) Register() error {
	return s.FakeResourceClient.Register()
}

func (s *slowFakeResourceClient) Read(namespace, name string, opts clients.ReadOpts) (*FakeResource, error) {
	return s.FakeResourceClient.Read(namespace, name, opts)
}

func (s *slowFakeResourceClient) Write(resource *FakeResource, opts clients.WriteOpts) (*FakeResource, error) {
	return s.FakeResourceClient.Write(resource, opts)
}

func (s *slowFakeResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	return s.FakeResourceClient.Delete(namespace, name, opts)
}

func (s *slowFakeResourceClient) List(namespace string, opts clients.ListOpts) (FakeResourceList, error) {
	nsLatency, ok := s.latencyPerNamespace[namespace]
	if ok {
		time.Sleep(nsLatency)
	}
	return s.FakeResourceClient.List(namespace, opts)
}

func (s *slowFakeResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan FakeResourceList, <-chan error, error) {
	nsLatency, ok := s.latencyPerNamespace[namespace]
	if ok {
		time.Sleep(nsLatency)
	}
	return s.FakeResourceClient.Watch(namespace, opts)
}
