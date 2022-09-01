// Code generated by solo-kit. DO NOT EDIT.

//go:build solokit
// +build solokit

package v1alpha1

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/namespace"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"
	"github.com/solo-io/solo-kit/test/helpers"
	corev1 "k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	// From https://github.com/kubernetes/client-go/blob/53c7adfd0294caa142d961e1f780f74081d5b15f/examples/out-of-cluster-client-configuration/main.go#L31
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("V1Alpha1Emitter", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		ctx                     context.Context
		namespace1, namespace2  string
		namespace3, namespace4  string
		namespace5, namespace6  string
		name1, name2            = "angela" + helpers.RandString(3), "bob" + helpers.RandString(3)
		name3, name4            = "susan" + helpers.RandString(3), "jim" + helpers.RandString(3)
		name5                   = "melisa" + helpers.RandString(3)
		labels1                 = map[string]string{"env": "test"}
		labelExpression1        = "env in (test)"
		cfg                     *rest.Config
		clientset               *apiext.Clientset
		kube                    kubernetes.Interface
		emitter                 TestingEmitter
		mockResourceClient      MockResourceClient
		resourceNamespaceLister resources.ResourceNamespaceLister
		kubeCache               cache.KubeCoreCache
	)
	const (
		TIME_BETWEEN_MESSAGES = 5
	)
	NewMockResourceWithLabels := func(namespace, name string, labels map[string]string) *MockResource {
		resource := NewMockResource(namespace, name)
		resource.GetMetadata().Labels = labels
		return resource
	}

	createNamespaces := func(ctx context.Context, kube kubernetes.Interface, namespaces ...string) {
		err := kubeutils.CreateNamespacesInParallel(ctx, kube, namespaces...)
		Expect(err).NotTo(HaveOccurred())
	}

	createNamespaceWithLabel := func(ctx context.Context, kube kubernetes.Interface, namespace string, labels map[string]string) {
		_, err := kube.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespace,
				Labels: labels,
			},
		}, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred())
	}

	deleteNamespaces := func(ctx context.Context, kube kubernetes.Interface, namespaces ...string) {
		err := kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespaces...)
		Expect(err).NotTo(HaveOccurred())
	}

	// getNewNamespaces is used to generate new namespace names, so that we do not have to wait
	// when deleting namespaces in runNamespacedSelectorsWithWatchNamespaces. Since
	// runNamespacedSelectorsWithWatchNamespaces uses watchNamespaces set to namespace1 and
	// namespace2, this will work. Because the emitter willl only be watching namespaces that are
	// labeled.
	getNewNamespaces := func() {
		namespace3 = helpers.RandString(8)
		namespace4 = helpers.RandString(8)
		namespace5 = helpers.RandString(8)
		namespace6 = helpers.RandString(8)
	}

	// getNewNamespaces1and2 is used to generate new namespaces for namespace 1 and 2.
	// used for the same reason as getNewNamespaces() above
	getNewNamespaces1and2 := func() {
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
	}

	runNamespacedSelectorsWithWatchNamespaces := func() {
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		// There is an error here in the code.
		snapshots, errs, err := emitter.Snapshots([]string{namespace1, namespace2}, clients.WatchOpts{
			Ctx:                ctx,
			RefreshRate:        time.Second,
			ExpressionSelector: labelExpression1,
		})
		Expect(err).NotTo(HaveOccurred())

		var snap *TestingSnapshot

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
		mockResourceWatched := MockResourceList{mockResource1a, mockResource1b}
		assertSnapshotMocks(mockResourceWatched, nil)

		mockResource3a, err := mockResourceClient.Write(NewMockResourceWithLabels(namespace1, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResource3b, err := mockResourceClient.Write(NewMockResourceWithLabels(namespace2, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResourceWatched = append(mockResourceWatched, MockResourceList{mockResource3a, mockResource3b}...)
		assertSnapshotMocks(mockResourceWatched, nil)

		createNamespaceWithLabel(ctx, kube, namespace3, labels1)
		createNamespaces(ctx, kube, namespace4)

		mockResource4a, err := mockResourceClient.Write(NewMockResource(namespace3, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResource4b, err := mockResourceClient.Write(NewMockResource(namespace4, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResourceWatched = append(mockResourceWatched, mockResource4a)
		mockResourceNotWatched := MockResourceList{mockResource4b}
		assertSnapshotMocks(mockResourceWatched, mockResourceNotWatched)

		mockResource5a, err := mockResourceClient.Write(NewMockResourceWithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResource5b, err := mockResourceClient.Write(NewMockResourceWithLabels(namespace4, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResourceWatched = append(mockResourceWatched, mockResource5a)
		mockResourceNotWatched = append(mockResourceNotWatched, mockResource5b)
		assertSnapshotMocks(mockResourceWatched, mockResourceNotWatched)

		for _, r := range mockResourceNotWatched {
			err = mockResourceClient.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
		}

		err = mockResourceClient.Delete(mockResource1a.GetMetadata().Namespace, mockResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockResourceClient.Delete(mockResource1b.GetMetadata().Namespace, mockResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResourceNotWatched = append(mockResourceNotWatched, MockResourceList{mockResource1a, mockResource1b}...)
		mockResourceWatched = MockResourceList{mockResource3a, mockResource3b, mockResource4a, mockResource5a}
		assertSnapshotMocks(mockResourceWatched, mockResourceNotWatched)

		err = mockResourceClient.Delete(mockResource3a.GetMetadata().Namespace, mockResource3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockResourceClient.Delete(mockResource3b.GetMetadata().Namespace, mockResource3b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResourceNotWatched = append(mockResourceNotWatched, MockResourceList{mockResource3a, mockResource3b}...)
		mockResourceWatched = MockResourceList{mockResource4a, mockResource5a}
		assertSnapshotMocks(mockResourceWatched, mockResourceNotWatched)

		err = mockResourceClient.Delete(mockResource4a.GetMetadata().Namespace, mockResource4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = mockResourceClient.Delete(mockResource5a.GetMetadata().Namespace, mockResource5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		mockResourceNotWatched = append(mockResourceNotWatched, MockResourceList{mockResource5a, mockResource5b}...)
		assertSnapshotMocks(nil, mockResourceNotWatched)

		// clean up environment
		deleteNamespaces(ctx, kube, namespace3, namespace4)
		getNewNamespaces()
	}

	BeforeEach(func() {
		err := os.Setenv(statusutils.PodNamespaceEnvName, "default")
		Expect(err).NotTo(HaveOccurred())

		ctx = context.Background()
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
		namespace3 = helpers.RandString(8)
		namespace4 = helpers.RandString(8)
		namespace5 = helpers.RandString(8)
		namespace6 = helpers.RandString(8)

		kube = helpers.MustKubeClient()
		kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		resourceNamespaceLister = namespace.NewKubeClientCacheResourceNamespaceLister(kube, kubeCache)

		createNamespaces(ctx, kube, namespace1, namespace2)

		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())

		clientset, err = apiext.NewForConfig(cfg)
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
		emitter = NewTestingEmitter(mockResourceClient, resourceNamespaceLister)
	})
	AfterEach(func() {
		err := os.Unsetenv(statusutils.PodNamespaceEnvName)
		Expect(err).NotTo(HaveOccurred())

		kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespace1, namespace2)
	})

	Context("Tracking watched namespaces", func() {
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
			mockResource1a, err := mockResourceClient.Write(NewMockResource(namespace1, name5), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResource1b, err := mockResourceClient.Write(NewMockResource(namespace2, name5), clients.WriteOpts{Ctx: ctx})
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
		})

		It("should be able to track all resources that are on labeled namespaces", func() {
			runNamespacedSelectorsWithWatchNamespaces()
		})
	})

	Context("Tracking empty watched namespaces", func() {
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
		})

		It("should be able to track resources only made with the matching labels", func() {
			ctx := context.Background()
			err := emitter.Register()
			Expect(err).NotTo(HaveOccurred())

			snapshots, errs, err := emitter.Snapshots([]string{""}, clients.WatchOpts{
				Ctx:                ctx,
				RefreshRate:        time.Second,
				ExpressionSelector: labelExpression1,
			})
			Expect(err).NotTo(HaveOccurred())

			var snap *TestingSnapshot

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
			mockResourceNotWatched := MockResourceList{mockResource1a, mockResource1b}

			createNamespaceWithLabel(ctx, kube, namespace3, labels1)
			createNamespaceWithLabel(ctx, kube, namespace4, labels1)

			mockResource2a, err := mockResourceClient.Write(NewMockResource(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResource2b, err := mockResourceClient.Write(NewMockResource(namespace4, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResourceWatched := MockResourceList{mockResource2a, mockResource2b}
			assertSnapshotMocks(mockResourceWatched, mockResourceNotWatched)

			createNamespaces(ctx, kube, namespace5)
			createNamespaceWithLabel(ctx, kube, namespace6, labels1)

			mockResource5a, err := mockResourceClient.Write(NewMockResource(namespace5, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResource5b, err := mockResourceClient.Write(NewMockResource(namespace6, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResourceNotWatched = append(mockResourceNotWatched, mockResource5a)
			mockResourceWatched = append(mockResourceWatched, mockResource5b)
			assertSnapshotMocks(mockResourceWatched, mockResourceNotWatched)

			mockResource7a, err := mockResourceClient.Write(NewMockResource(namespace5, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResource7b, err := mockResourceClient.Write(NewMockResource(namespace6, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResourceNotWatched = append(mockResourceNotWatched, mockResource7a)
			mockResourceWatched = append(mockResourceWatched, mockResource7b)
			assertSnapshotMocks(mockResourceWatched, mockResourceNotWatched)

			for _, r := range mockResourceNotWatched {
				err = mockResourceClient.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
			}

			for _, r := range mockResourceWatched {
				err = mockResourceClient.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				mockResourceNotWatched = append(mockResourceNotWatched, r)
			}
			assertSnapshotMocks(nil, mockResourceNotWatched)

			// clean up environment
			deleteNamespaces(ctx, kube, namespace3, namespace4, namespace5, namespace6)
			getNewNamespaces()
		})
	})

	Context("Tracking resources on namespaces that are deleted", func() {
		It("Should not contain resources from a deleted namespace", func() {
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
			mockResource1b, err := mockResourceClient.Write(NewMockResource(namespace2, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResourceWatched := MockResourceList{mockResource1a, mockResource1b}
			assertSnapshotMocks(mockResourceWatched, nil)
			err = mockResourceClient.Delete(mockResource1a.GetMetadata().Namespace, mockResource1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = mockResourceClient.Delete(mockResource1b.GetMetadata().Namespace, mockResource1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())

			mockResourceNotWatched := MockResourceList{mockResource1a, mockResource1b}
			assertSnapshotMocks(nil, mockResourceNotWatched)

			deleteNamespaces(ctx, kube, namespace1, namespace2)

			getNewNamespaces1and2()
			createNamespaces(ctx, kube, namespace1, namespace2)
		})

		It("Should not contain resources from a deleted namespace, that is filtered", func() {
			ctx := context.Background()
			err := emitter.Register()
			Expect(err).NotTo(HaveOccurred())

			snapshots, errs, err := emitter.Snapshots([]string{""}, clients.WatchOpts{
				Ctx:                ctx,
				RefreshRate:        time.Second,
				ExpressionSelector: labelExpression1,
			})
			Expect(err).NotTo(HaveOccurred())

			var snap *TestingSnapshot

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

			// create namespaces
			createNamespaceWithLabel(ctx, kube, namespace3, labels1)
			createNamespaceWithLabel(ctx, kube, namespace4, labels1)

			mockResource2a, err := mockResourceClient.Write(NewMockResource(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResource2b, err := mockResourceClient.Write(NewMockResource(namespace4, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			mockResourceNotWatched := MockResourceList{}
			mockResourceWatched := MockResourceList{mockResource2a, mockResource2b}
			assertSnapshotMocks(mockResourceWatched, mockResourceNotWatched)

			deleteNamespaces(ctx, kube, namespace3)

			mockResourceWatched = MockResourceList{mockResource2b}
			mockResourceNotWatched = append(mockResourceNotWatched, mockResource2a)
			assertSnapshotMocks(mockResourceWatched, mockResourceNotWatched)

			for _, r := range mockResourceWatched {
				err = mockResourceClient.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				mockResourceNotWatched = append(mockResourceNotWatched, r)
			}
			assertSnapshotMocks(nil, mockResourceNotWatched)

			deleteNamespaces(ctx, kube, namespace4)
			getNewNamespaces()
		})
	})

	Context("use different resource namespace listers", func() {
		BeforeEach(func() {
			resourceNamespaceLister = namespace.NewKubeClientResourceNamespaceLister(kube)
			emitter = NewTestingEmitter(mockResourceClient, resourceNamespaceLister)
		})

		It("Should work with the Kube Client Namespace Lister", func() {
			runNamespacedSelectorsWithWatchNamespaces()
		})
	})

})
