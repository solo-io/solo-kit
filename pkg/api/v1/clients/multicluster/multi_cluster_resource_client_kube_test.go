package multicluster

import (
	"context"
	"log"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	kubehelpers "github.com/solo-io/go-utils/testutils/kube"
	skapi "github.com/solo-io/solo-kit/api/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/ClientFactory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	sk_multicluster "github.com/solo-io/solo-kit/pkg/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	multicluster_v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
	"github.com/solo-io/solo-kit/test/helpers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/testutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ = Describe("MultiClusterResourceClient e2e test", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}

	var (
		namespace                 string
		anotherMockResourceClient v1.AnotherMockResourceClient
		watchAggregator           wrapper.WatchAggregator
		localRestConfig           *rest.Config
		localKubeClient           kubernetes.Interface
		ctx                       context.Context
		cancel                    context.CancelFunc
		cfgSecret                 *corev1.Secret
		remoteCluster             = "cluster-two"
		// resources on different clusters can have the same name and namespace
		resourceName = "name"
	)

	getAnotherMockResource := func(cluster, ns, name, basicField string) *v1.AnotherMockResource {
		return &v1.AnotherMockResource{
			Metadata: core.Metadata{
				Name:      name,
				Namespace: ns,
				Cluster:   cluster,
			},
			BasicField: basicField,
		}
	}

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		cacheManager, err := clustercache.NewCacheManager(ctx, kube.NewKubeSharedCacheForConfig)
		Expect(err).NotTo(HaveOccurred())
		ClientFactory := ClientFactory.NewKubeResourceClientFactory(
			cacheManager,
			v1.AnotherMockResourceCrd,
			false,
			nil,
			0,
			factory.NewResourceClientParams{ResourceType: &v1.AnotherMockResource{}},
		)
		watchAggregator = wrapper.NewWatchAggregator()
		watchHandler := NewAggregatedWatchClusterClientHandler(watchAggregator)
		clientSet := NewClusterClientManager(context.Background(), ClientFactory, watchHandler)
		mcrc := NewMultiClusterResourceClient(&v1.AnotherMockResource{}, clientSet)

		configWatcher := sk_multicluster.NewKubeConfigWatcher()
		localRestConfig, err = kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
		Expect(err).NotTo(HaveOccurred())
		localKubeClient, err = kubernetes.NewForConfig(localRestConfig)
		Expect(err).NotTo(HaveOccurred())
		localCache, err := cache.NewKubeCoreCache(ctx, localKubeClient)
		Expect(err).NotTo(HaveOccurred())

		restConfigHandler := sk_multicluster.NewRestConfigHandler(configWatcher, cacheManager, clientSet)
		errs, err := restConfigHandler.Run(ctx, localRestConfig, localKubeClient, localCache)
		Eventually(errs).ShouldNot(Receive())
		Expect(err).NotTo(HaveOccurred())

		anotherMockResourceClient = v1.NewAnotherMockResourceClientWithBase(mcrc)
		eventuallyClusterAvailable(anotherMockResourceClient, sk_multicluster.LocalCluster)

		// Create namespaces
		namespace = helpers.RandString(6)
		err = kubeutils.CreateNamespacesInParallel(localKubeClient, namespace)
		Expect(err).NotTo(HaveOccurred())
		err = kubeutils.CreateNamespacesInParallel(remoteKubeClient(), namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		// Delete remote cluster kube config secret
		err := localKubeClient.CoreV1().Secrets(namespace).Delete(cfgSecret.GetName(), &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)

		remoteClient := remoteKubeClient()

		// Delete namespaces
		err = kubeutils.DeleteNamespacesInParallelBlocking(localKubeClient, namespace)
		Expect(err).NotTo(HaveOccurred())
		err = kubeutils.DeleteNamespacesInParallelBlocking(remoteClient, namespace)
		Expect(err).NotTo(HaveOccurred())
		kubehelpers.WaitForNamespaceTeardownWithClient(namespace, localKubeClient)
		kubehelpers.WaitForNamespaceTeardownWithClient(namespace, remoteClient)
		cancel()
	})

	It("cruds across clusters", func() {
		// write a resource to the local cluster
		localResource := getAnotherMockResource(sk_multicluster.LocalCluster, namespace, resourceName, "foo")
		localWritten, err := anotherMockResourceClient.Write(localResource, clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		expectEqualAnotherMockResources(localWritten, localResource)

		// list local resources
		localList, err := anotherMockResourceClient.List(namespace, clients.ListOpts{Ctx: ctx, Cluster: sk_multicluster.LocalCluster})
		Expect(err).NotTo(HaveOccurred())
		Expect(localList).To(HaveLen(1))
		expectEqualAnotherMockResources(localResource, localList[0])

		// register another cluster
		anotherKubeConfig, err := kubeutils.GetKubeConfig("", os.Getenv("ALT_CLUSTER_KUBECONFIG"))
		Expect(err).NotTo(HaveOccurred())
		skKubeConfig := &multicluster_v1.KubeConfig{
			KubeConfig: skapi.KubeConfig{
				Metadata: core.Metadata{Name: "remote-cluster", Namespace: namespace},
				Config:   *anotherKubeConfig,
				Cluster:  remoteCluster,
			},
		}
		cfgSecret, err = secretconverter.KubeConfigToSecret(skKubeConfig)
		Expect(err).NotTo(HaveOccurred())
		_, err = localKubeClient.CoreV1().Secrets(namespace).Create(cfgSecret)
		Expect(err).NotTo(HaveOccurred())
		eventuallyClusterAvailable(anotherMockResourceClient, remoteCluster)

		// write a resource to the other cluster
		remoteResource := getAnotherMockResource(remoteCluster, namespace, resourceName, "bar")
		remoteWritten, err := anotherMockResourceClient.Write(remoteResource, clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		expectEqualAnotherMockResources(remoteWritten, remoteResource)

		// list local and remote resources, verify they are distinct
		localList, err = anotherMockResourceClient.List(namespace, clients.ListOpts{Ctx: ctx, Cluster: sk_multicluster.LocalCluster})
		Expect(err).NotTo(HaveOccurred())
		Expect(localList).To(HaveLen(1))
		expectEqualAnotherMockResources(localResource, localList[0])
		remoteList, err := anotherMockResourceClient.List(namespace, clients.ListOpts{Ctx: ctx, Cluster: remoteCluster})
		Expect(err).NotTo(HaveOccurred())
		Expect(remoteList).To(HaveLen(1))
		expectEqualAnotherMockResources(remoteResource, remoteList[0])
		Expect(remoteList).NotTo(Equal(localList))

		// read
		localRead, err := anotherMockResourceClient.Read(namespace, resourceName, clients.ReadOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		expectEqualAnotherMockResources(localResource, localRead)
		remoteRead, err := anotherMockResourceClient.Read(namespace, resourceName, clients.ReadOpts{Ctx: ctx, Cluster: remoteCluster})
		Expect(err).NotTo(HaveOccurred())
		expectEqualAnotherMockResources(remoteResource, remoteRead)

		// local watch
		w, errs, err := anotherMockResourceClient.Watch(namespace, clients.WatchOpts{
			Ctx:         ctx,
			Selector:    nil,
			RefreshRate: 10 * time.Second,
		})
		Expect(err).NotTo(HaveOccurred())
		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		case <-time.After(5 * time.Millisecond):
			Fail("expected message in local watch channel")
		case list := <-w:
			Expect(list).To(HaveLen(1))
			expectEqualAnotherMockResources(list[0], localResource)
		}

		// remote watch
		w, errs, err = anotherMockResourceClient.Watch(namespace, clients.WatchOpts{
			Ctx:         ctx,
			Selector:    nil,
			RefreshRate: 10 * time.Second,
			Cluster:     remoteCluster,
		})
		Expect(err).NotTo(HaveOccurred())
		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		case <-time.After(5 * time.Millisecond):
			Fail("expected a message in channel")
		case list := <-w:
			Expect(list).To(HaveLen(1))
			expectEqualAnotherMockResources(list[0], remoteResource)
		}

		// aggregated watch
		aw, errs, err := watchAggregator.Watch(namespace, clients.WatchOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		Eventually(errs).ShouldNot(Receive())
		Eventually(aw).Should(Receive(And(ContainElement(localWritten), ContainElement(remoteWritten))))

		// delete
		err = anotherMockResourceClient.Delete(namespace, resourceName, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = anotherMockResourceClient.Delete(namespace, resourceName, clients.DeleteOpts{Ctx: ctx, Cluster: remoteCluster})
		Expect(err).NotTo(HaveOccurred())
	})
})

func expectEqualAnotherMockResources(a, b *v1.AnotherMockResource) {
	Expect(a.Metadata.Cluster).To(Equal(b.Metadata.Cluster))
	Expect(a.Metadata.Namespace).To(Equal(b.Metadata.Namespace))
	Expect(a.Metadata.Name).To(Equal(b.Metadata.Name))
	Expect(a.BasicField).To(Equal(b.BasicField))
}

// Wait for client to become available
func eventuallyClusterAvailable(client v1.AnotherMockResourceClient, cluster string) {
	Eventually(func() error {
		_, err := client.List("", clients.ListOpts{Cluster: cluster})
		return err
	}, 10*time.Second, 1*time.Second).Should(BeNil())
}
