//go:build solokit
// +build solokit

package v1

import (
	"context"
	"os"
	"time"

	github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/test/helpers"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

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
		namespace2                string
		name1, name2              = "angela" + helpers.RandString(3), "bob" + helpers.RandString(3)
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
		namespace2 = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err = kubeutils.CreateNamespacesInParallel(ctx, kube, namespace1, namespace2)
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

		mockCustomSpecHashTypeClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		mockCustomSpecHashTypeClient, err := NewMockCustomSpecHashTypeClient(ctx, mockCustomSpecHashTypeClientFactory)
		Expect(err).NotTo(HaveOccurred())

		// Pod Constructor
		podClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}

		podClient, err = github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPodClient(ctx, podClientFactory)
		Expect(err).NotTo(HaveOccurred())
		emitter = NewTestingEmitter(simpleMockResourceClient, mockResourceClient, fakeResourceClient, anotherMockResourceClient, clusterResourceClient, mockCustomTypeClient, mockCustomSpecHashTypeClient, podClient)
	})
	AfterEach(func() {
		err := os.Unsetenv(statusutils.PodNamespaceEnvName)
		Expect(err).NotTo(HaveOccurred())

		err = kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespace1, namespace2)
		Expect(err).NotTo(HaveOccurred())
		clusterResourceClient.Delete(name1, clients.DeleteOpts{})
		clusterResourceClient.Delete(name2, clients.DeleteOpts{})
	})

	It("tracks snapshots on changes to any resource", func() {
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		snapshots, errs, err := emitter.Snapshots([]string{namespace1, namespace2}, clients.WatchOpts{
			Ctx:         ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())

		var snap *TestingSnapshot

		/*
			SimpleMockResource
		*/

		assertSnapshotSimplemocks := func(expectSimplemocks SimpleMockResourceList, unexpectSimplemocks SimpleMockResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectSimplemocks {
						if _, err := snap.Simplemocks.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectSimplemocks {
						if _, err := snap.Simplemocks.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := simpleMockResourceClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := simpleMockResourceClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		simpleMockResource1a, err := simpleMockResourceClient.Write(NewSimpleMockResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		simpleMockResource1b, err := simpleMockResourceClient.Write(NewSimpleMockResource(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotSimplemocks(SimpleMockResourceList{simpleMockResource1a, simpleMockResource1b}, nil)
		simpleMockResource2a, err := simpleMockResourceClient.Write(NewSimpleMockResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		simpleMockResource2b, err := simpleMockResourceClient.Write(NewSimpleMockResource(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotSimplemocks(SimpleMockResourceList{simpleMockResource1a, simpleMockResource1b, simpleMockResource2a, simpleMockResource2b}, nil)

		err = simpleMockResourceClient.Delete(simpleMockResource2a.GetMetadata().Namespace, simpleMockResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = simpleMockResourceClient.Delete(simpleMockResource2b.GetMetadata().Namespace, simpleMockResource2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotSimplemocks(SimpleMockResourceList{simpleMockResource1a, simpleMockResource1b}, SimpleMockResourceList{simpleMockResource2a, simpleMockResource2b})

		err = simpleMockResourceClient.Delete(simpleMockResource1a.GetMetadata().Namespace, simpleMockResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = simpleMockResourceClient.Delete(simpleMockResource1b.GetMetadata().Namespace, simpleMockResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotSimplemocks(nil, SimpleMockResourceList{simpleMockResource1a, simpleMockResource1b, simpleMockResource2a, simpleMockResource2b})

		/*
			MockResource
		*/

		assertSnapshotMocks := func(expectMocks MockResourceList, unexpectMocks MockResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectMocks {
						if _, err := snap.Mocks.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectMocks {
						if _, err := snap.Mocks.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := mockResourceClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := mockResourceClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		mockResource1a, err := mockResourceClient.Write(NewMockResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResource1b, err := mockResourceClient.Write(NewMockResource(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotMocks(MockResourceList{mockResource1a, mockResource1b}, nil)
		mockResource2a, err := mockResourceClient.Write(NewMockResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResource2b, err := mockResourceClient.Write(NewMockResource(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotMocks(MockResourceList{mockResource1a, mockResource1b, mockResource2a, mockResource2b}, nil)

		err = mockResourceClient.Delete(mockResource2a.GetMetadata().Namespace, mockResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockResourceClient.Delete(mockResource2b.GetMetadata().Namespace, mockResource2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotMocks(MockResourceList{mockResource1a, mockResource1b}, MockResourceList{mockResource2a, mockResource2b})

		err = mockResourceClient.Delete(mockResource1a.GetMetadata().Namespace, mockResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockResourceClient.Delete(mockResource1b.GetMetadata().Namespace, mockResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotMocks(nil, MockResourceList{mockResource1a, mockResource1b, mockResource2a, mockResource2b})

		/*
			FakeResource
		*/

		assertSnapshotFakes := func(expectFakes FakeResourceList, unexpectFakes FakeResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectFakes {
						if _, err := snap.Fakes.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectFakes {
						if _, err := snap.Fakes.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := fakeResourceClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := fakeResourceClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		fakeResource1a, err := fakeResourceClient.Write(NewFakeResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		fakeResource1b, err := fakeResourceClient.Write(NewFakeResource(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotFakes(FakeResourceList{fakeResource1a, fakeResource1b}, nil)
		fakeResource2a, err := fakeResourceClient.Write(NewFakeResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		fakeResource2b, err := fakeResourceClient.Write(NewFakeResource(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotFakes(FakeResourceList{fakeResource1a, fakeResource1b, fakeResource2a, fakeResource2b}, nil)

		err = fakeResourceClient.Delete(fakeResource2a.GetMetadata().Namespace, fakeResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = fakeResourceClient.Delete(fakeResource2b.GetMetadata().Namespace, fakeResource2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotFakes(FakeResourceList{fakeResource1a, fakeResource1b}, FakeResourceList{fakeResource2a, fakeResource2b})

		err = fakeResourceClient.Delete(fakeResource1a.GetMetadata().Namespace, fakeResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = fakeResourceClient.Delete(fakeResource1b.GetMetadata().Namespace, fakeResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotFakes(nil, FakeResourceList{fakeResource1a, fakeResource1b, fakeResource2a, fakeResource2b})

		/*
			AnotherMockResource
		*/

		assertSnapshotAnothermockresources := func(expectAnothermockresources AnotherMockResourceList, unexpectAnothermockresources AnotherMockResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectAnothermockresources {
						if _, err := snap.Anothermockresources.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectAnothermockresources {
						if _, err := snap.Anothermockresources.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := anotherMockResourceClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := anotherMockResourceClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		anotherMockResource1a, err := anotherMockResourceClient.Write(NewAnotherMockResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		anotherMockResource1b, err := anotherMockResourceClient.Write(NewAnotherMockResource(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotAnothermockresources(AnotherMockResourceList{anotherMockResource1a, anotherMockResource1b}, nil)
		anotherMockResource2a, err := anotherMockResourceClient.Write(NewAnotherMockResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		anotherMockResource2b, err := anotherMockResourceClient.Write(NewAnotherMockResource(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotAnothermockresources(AnotherMockResourceList{anotherMockResource1a, anotherMockResource1b, anotherMockResource2a, anotherMockResource2b}, nil)

		err = anotherMockResourceClient.Delete(anotherMockResource2a.GetMetadata().Namespace, anotherMockResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = anotherMockResourceClient.Delete(anotherMockResource2b.GetMetadata().Namespace, anotherMockResource2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotAnothermockresources(AnotherMockResourceList{anotherMockResource1a, anotherMockResource1b}, AnotherMockResourceList{anotherMockResource2a, anotherMockResource2b})

		err = anotherMockResourceClient.Delete(anotherMockResource1a.GetMetadata().Namespace, anotherMockResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = anotherMockResourceClient.Delete(anotherMockResource1b.GetMetadata().Namespace, anotherMockResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotAnothermockresources(nil, AnotherMockResourceList{anotherMockResource1a, anotherMockResource1b, anotherMockResource2a, anotherMockResource2b})

		/*
			ClusterResource
		*/

		assertSnapshotClusterresources := func(expectClusterresources ClusterResourceList, unexpectClusterresources ClusterResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectClusterresources {
						if _, err := snap.Clusterresources.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectClusterresources {
						if _, err := snap.Clusterresources.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					combined, _ := clusterResourceClient.List(clients.ListOpts{})
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		clusterResource1a, err := clusterResourceClient.Write(NewClusterResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotClusterresources(ClusterResourceList{clusterResource1a}, nil)
		clusterResource2a, err := clusterResourceClient.Write(NewClusterResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotClusterresources(ClusterResourceList{clusterResource1a, clusterResource2a}, nil)

		err = clusterResourceClient.Delete(clusterResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotClusterresources(ClusterResourceList{clusterResource1a}, ClusterResourceList{clusterResource2a})

		err = clusterResourceClient.Delete(clusterResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotClusterresources(nil, ClusterResourceList{clusterResource1a, clusterResource2a})

		/*
			MockCustomType
		*/

		assertSnapshotmcts := func(expectmcts MockCustomTypeList, unexpectmcts MockCustomTypeList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectmcts {
						if _, err := snap.Mcts.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectmcts {
						if _, err := snap.Mcts.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := mockCustomTypeClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := mockCustomTypeClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		mockCustomType1a, err := mockCustomTypeClient.Write(NewMockCustomType(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockCustomType1b, err := mockCustomTypeClient.Write(NewMockCustomType(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotmcts(MockCustomTypeList{mockCustomType1a, mockCustomType1b}, nil)
		mockCustomType2a, err := mockCustomTypeClient.Write(NewMockCustomType(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockCustomType2b, err := mockCustomTypeClient.Write(NewMockCustomType(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotmcts(MockCustomTypeList{mockCustomType1a, mockCustomType1b, mockCustomType2a, mockCustomType2b}, nil)

		err = mockCustomTypeClient.Delete(mockCustomType2a.GetMetadata().Namespace, mockCustomType2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockCustomTypeClient.Delete(mockCustomType2b.GetMetadata().Namespace, mockCustomType2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotmcts(MockCustomTypeList{mockCustomType1a, mockCustomType1b}, MockCustomTypeList{mockCustomType2a, mockCustomType2b})

		err = mockCustomTypeClient.Delete(mockCustomType1a.GetMetadata().Namespace, mockCustomType1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockCustomTypeClient.Delete(mockCustomType1b.GetMetadata().Namespace, mockCustomType1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotmcts(nil, MockCustomTypeList{mockCustomType1a, mockCustomType1b, mockCustomType2a, mockCustomType2b})

		/*
			Pod
		*/

		assertSnapshotpods := func(expectpods github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList, unexpectpods github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectpods {
						if _, err := snap.Pods.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectpods {
						if _, err := snap.Pods.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := podClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := podClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		pod1a, err := podClient.Write(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPod(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		pod1b, err := podClient.Write(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPod(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotpods(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod1a, pod1b}, nil)
		pod2a, err := podClient.Write(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPod(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		pod2b, err := podClient.Write(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPod(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotpods(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod1a, pod1b, pod2a, pod2b}, nil)

		err = podClient.Delete(pod2a.GetMetadata().Namespace, pod2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = podClient.Delete(pod2b.GetMetadata().Namespace, pod2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotpods(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod1a, pod1b}, github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod2a, pod2b})

		err = podClient.Delete(pod1a.GetMetadata().Namespace, pod1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = podClient.Delete(pod1b.GetMetadata().Namespace, pod1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotpods(nil, github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod1a, pod1b, pod2a, pod2b})
	})

	It("tracks snapshots on changes to any resource using AllNamespace", func() {
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		snapshots, errs, err := emitter.Snapshots([]string{""}, clients.WatchOpts{
			Ctx:         ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())

		var snap *TestingSnapshot

		/*
			SimpleMockResource
		*/

		assertSnapshotSimplemocks := func(expectSimplemocks SimpleMockResourceList, unexpectSimplemocks SimpleMockResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectSimplemocks {
						if _, err := snap.Simplemocks.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectSimplemocks {
						if _, err := snap.Simplemocks.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := simpleMockResourceClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := simpleMockResourceClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		simpleMockResource1a, err := simpleMockResourceClient.Write(NewSimpleMockResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		simpleMockResource1b, err := simpleMockResourceClient.Write(NewSimpleMockResource(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotSimplemocks(SimpleMockResourceList{simpleMockResource1a, simpleMockResource1b}, nil)
		simpleMockResource2a, err := simpleMockResourceClient.Write(NewSimpleMockResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		simpleMockResource2b, err := simpleMockResourceClient.Write(NewSimpleMockResource(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotSimplemocks(SimpleMockResourceList{simpleMockResource1a, simpleMockResource1b, simpleMockResource2a, simpleMockResource2b}, nil)

		err = simpleMockResourceClient.Delete(simpleMockResource2a.GetMetadata().Namespace, simpleMockResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = simpleMockResourceClient.Delete(simpleMockResource2b.GetMetadata().Namespace, simpleMockResource2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotSimplemocks(SimpleMockResourceList{simpleMockResource1a, simpleMockResource1b}, SimpleMockResourceList{simpleMockResource2a, simpleMockResource2b})

		err = simpleMockResourceClient.Delete(simpleMockResource1a.GetMetadata().Namespace, simpleMockResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = simpleMockResourceClient.Delete(simpleMockResource1b.GetMetadata().Namespace, simpleMockResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotSimplemocks(nil, SimpleMockResourceList{simpleMockResource1a, simpleMockResource1b, simpleMockResource2a, simpleMockResource2b})

		/*
			MockResource
		*/

		assertSnapshotMocks := func(expectMocks MockResourceList, unexpectMocks MockResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectMocks {
						if _, err := snap.Mocks.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectMocks {
						if _, err := snap.Mocks.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := mockResourceClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := mockResourceClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		mockResource1a, err := mockResourceClient.Write(NewMockResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResource1b, err := mockResourceClient.Write(NewMockResource(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotMocks(MockResourceList{mockResource1a, mockResource1b}, nil)
		mockResource2a, err := mockResourceClient.Write(NewMockResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResource2b, err := mockResourceClient.Write(NewMockResource(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotMocks(MockResourceList{mockResource1a, mockResource1b, mockResource2a, mockResource2b}, nil)

		err = mockResourceClient.Delete(mockResource2a.GetMetadata().Namespace, mockResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockResourceClient.Delete(mockResource2b.GetMetadata().Namespace, mockResource2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotMocks(MockResourceList{mockResource1a, mockResource1b}, MockResourceList{mockResource2a, mockResource2b})

		err = mockResourceClient.Delete(mockResource1a.GetMetadata().Namespace, mockResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockResourceClient.Delete(mockResource1b.GetMetadata().Namespace, mockResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotMocks(nil, MockResourceList{mockResource1a, mockResource1b, mockResource2a, mockResource2b})

		/*
			FakeResource
		*/

		assertSnapshotFakes := func(expectFakes FakeResourceList, unexpectFakes FakeResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectFakes {
						if _, err := snap.Fakes.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectFakes {
						if _, err := snap.Fakes.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := fakeResourceClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := fakeResourceClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		fakeResource1a, err := fakeResourceClient.Write(NewFakeResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		fakeResource1b, err := fakeResourceClient.Write(NewFakeResource(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotFakes(FakeResourceList{fakeResource1a, fakeResource1b}, nil)
		fakeResource2a, err := fakeResourceClient.Write(NewFakeResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		fakeResource2b, err := fakeResourceClient.Write(NewFakeResource(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotFakes(FakeResourceList{fakeResource1a, fakeResource1b, fakeResource2a, fakeResource2b}, nil)

		err = fakeResourceClient.Delete(fakeResource2a.GetMetadata().Namespace, fakeResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = fakeResourceClient.Delete(fakeResource2b.GetMetadata().Namespace, fakeResource2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotFakes(FakeResourceList{fakeResource1a, fakeResource1b}, FakeResourceList{fakeResource2a, fakeResource2b})

		err = fakeResourceClient.Delete(fakeResource1a.GetMetadata().Namespace, fakeResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = fakeResourceClient.Delete(fakeResource1b.GetMetadata().Namespace, fakeResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotFakes(nil, FakeResourceList{fakeResource1a, fakeResource1b, fakeResource2a, fakeResource2b})

		/*
			AnotherMockResource
		*/

		assertSnapshotAnothermockresources := func(expectAnothermockresources AnotherMockResourceList, unexpectAnothermockresources AnotherMockResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectAnothermockresources {
						if _, err := snap.Anothermockresources.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectAnothermockresources {
						if _, err := snap.Anothermockresources.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := anotherMockResourceClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := anotherMockResourceClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		anotherMockResource1a, err := anotherMockResourceClient.Write(NewAnotherMockResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		anotherMockResource1b, err := anotherMockResourceClient.Write(NewAnotherMockResource(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotAnothermockresources(AnotherMockResourceList{anotherMockResource1a, anotherMockResource1b}, nil)
		anotherMockResource2a, err := anotherMockResourceClient.Write(NewAnotherMockResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		anotherMockResource2b, err := anotherMockResourceClient.Write(NewAnotherMockResource(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotAnothermockresources(AnotherMockResourceList{anotherMockResource1a, anotherMockResource1b, anotherMockResource2a, anotherMockResource2b}, nil)

		err = anotherMockResourceClient.Delete(anotherMockResource2a.GetMetadata().Namespace, anotherMockResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = anotherMockResourceClient.Delete(anotherMockResource2b.GetMetadata().Namespace, anotherMockResource2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotAnothermockresources(AnotherMockResourceList{anotherMockResource1a, anotherMockResource1b}, AnotherMockResourceList{anotherMockResource2a, anotherMockResource2b})

		err = anotherMockResourceClient.Delete(anotherMockResource1a.GetMetadata().Namespace, anotherMockResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = anotherMockResourceClient.Delete(anotherMockResource1b.GetMetadata().Namespace, anotherMockResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotAnothermockresources(nil, AnotherMockResourceList{anotherMockResource1a, anotherMockResource1b, anotherMockResource2a, anotherMockResource2b})

		/*
			ClusterResource
		*/

		assertSnapshotClusterresources := func(expectClusterresources ClusterResourceList, unexpectClusterresources ClusterResourceList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectClusterresources {
						if _, err := snap.Clusterresources.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectClusterresources {
						if _, err := snap.Clusterresources.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					combined, _ := clusterResourceClient.List(clients.ListOpts{})
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		clusterResource1a, err := clusterResourceClient.Write(NewClusterResource(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotClusterresources(ClusterResourceList{clusterResource1a}, nil)
		clusterResource2a, err := clusterResourceClient.Write(NewClusterResource(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotClusterresources(ClusterResourceList{clusterResource1a, clusterResource2a}, nil)

		err = clusterResourceClient.Delete(clusterResource2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotClusterresources(ClusterResourceList{clusterResource1a}, ClusterResourceList{clusterResource2a})

		err = clusterResourceClient.Delete(clusterResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotClusterresources(nil, ClusterResourceList{clusterResource1a, clusterResource2a})

		/*
			MockCustomType
		*/

		assertSnapshotmcts := func(expectmcts MockCustomTypeList, unexpectmcts MockCustomTypeList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectmcts {
						if _, err := snap.Mcts.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectmcts {
						if _, err := snap.Mcts.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := mockCustomTypeClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := mockCustomTypeClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		mockCustomType1a, err := mockCustomTypeClient.Write(NewMockCustomType(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockCustomType1b, err := mockCustomTypeClient.Write(NewMockCustomType(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotmcts(MockCustomTypeList{mockCustomType1a, mockCustomType1b}, nil)
		mockCustomType2a, err := mockCustomTypeClient.Write(NewMockCustomType(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockCustomType2b, err := mockCustomTypeClient.Write(NewMockCustomType(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotmcts(MockCustomTypeList{mockCustomType1a, mockCustomType1b, mockCustomType2a, mockCustomType2b}, nil)

		err = mockCustomTypeClient.Delete(mockCustomType2a.GetMetadata().Namespace, mockCustomType2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockCustomTypeClient.Delete(mockCustomType2b.GetMetadata().Namespace, mockCustomType2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotmcts(MockCustomTypeList{mockCustomType1a, mockCustomType1b}, MockCustomTypeList{mockCustomType2a, mockCustomType2b})

		err = mockCustomTypeClient.Delete(mockCustomType1a.GetMetadata().Namespace, mockCustomType1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockCustomTypeClient.Delete(mockCustomType1b.GetMetadata().Namespace, mockCustomType1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotmcts(nil, MockCustomTypeList{mockCustomType1a, mockCustomType1b, mockCustomType2a, mockCustomType2b})

		/*
			Pod
		*/

		assertSnapshotpods := func(expectpods github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList, unexpectpods github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectpods {
						if _, err := snap.Pods.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectpods {
						if _, err := snap.Pods.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := podClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := podClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		pod1a, err := podClient.Write(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPod(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		pod1b, err := podClient.Write(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPod(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotpods(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod1a, pod1b}, nil)
		pod2a, err := podClient.Write(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPod(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		pod2b, err := podClient.Write(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.NewPod(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotpods(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod1a, pod1b, pod2a, pod2b}, nil)

		err = podClient.Delete(pod2a.GetMetadata().Namespace, pod2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = podClient.Delete(pod2b.GetMetadata().Namespace, pod2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotpods(github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod1a, pod1b}, github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod2a, pod2b})

		err = podClient.Delete(pod1a.GetMetadata().Namespace, pod1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = podClient.Delete(pod1b.GetMetadata().Namespace, pod1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotpods(nil, github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.PodList{pod1a, pod1b, pod2a, pod2b})
	})
})
