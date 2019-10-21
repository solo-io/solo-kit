package multicluster_test

import (
	"context"
	"log"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	skapi "github.com/solo-io/solo-kit/api/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/clientgetter"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	multicluster2 "github.com/solo-io/solo-kit/pkg/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	skpkg "github.com/solo-io/solo-kit/pkg/multicluster/v1"
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
		subject         v1.AnotherMockResourceClient
		watchAggregator wrapper.WatchAggregator
		localRestConfig *rest.Config
		localKubeClient kubernetes.Interface
		ctx             context.Context
		cancel          context.CancelFunc
		cfgSecret       *corev1.Secret
		remoteCluster   = "cluster-two"
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
		clientGetter := clientgetter.NewKubeResourceClientGetter(
			cacheManager,
			v1.AnotherMockResourceCrd,
			false,
			nil,
			0,
			factory.NewResourceClientParams{ResourceType: &v1.AnotherMockResource{}},
		)
		watchAggregator = wrapper.NewWatchAggregator()
		mcrc := multicluster.NewMultiClusterResourceClient(&v1.AnotherMockResource{}, clientGetter, watchAggregator)

		configWatcher := multicluster2.NewKubeConfigWatcher()
		localRestConfig, err = kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
		Expect(err).NotTo(HaveOccurred())
		localKubeClient, err = kubernetes.NewForConfig(localRestConfig)
		Expect(err).NotTo(HaveOccurred())
		localCache, err := cache.NewKubeCoreCache(ctx, localKubeClient)
		Expect(err).NotTo(HaveOccurred())

		restConfigHandler := multicluster2.NewRestConfigHandler(configWatcher, cacheManager, mcrc)
		_, err = restConfigHandler.Run(ctx, localRestConfig, localKubeClient, localCache)
		Expect(err).NotTo(HaveOccurred())

		subject = v1.NewAnotherMockResourceClientWithBase(mcrc)
		EventuallyClusterAvailable(subject, multicluster2.LocalCluster)
	})

	AfterEach(func() {
		err := localKubeClient.CoreV1().Secrets(namespace).Delete(cfgSecret.GetName(), &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		cancel()
	})

	It("cruds across clusters", func() {
		// write a resource to the local cluster
		localResource := getAnotherMockResource(multicluster2.LocalCluster, namespace, resourceName, "foo")
		localWritten, err := subject.Write(localResource, clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		ExpectEqualAnotherMockResources(localWritten, localResource)

		// list local resources
		localList, err := subject.List(namespace, clients.ListOpts{Ctx: ctx, Cluster: multicluster2.LocalCluster})
		Expect(err).NotTo(HaveOccurred())
		Expect(localList).To(HaveLen(1))
		ExpectEqualAnotherMockResources(localResource, localList[0])

		// register another cluster
		anotherKubeConfig, err := kubeutils.GetKubeConfig("", os.Getenv("ALT_CLUSTER_KUBECONFIG"))
		Expect(err).NotTo(HaveOccurred())
		skKubeConfig := &skpkg.KubeConfig{
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
		EventuallyClusterAvailable(subject, remoteCluster)

		// write a resource to the other cluster
		remoteResource := getAnotherMockResource(remoteCluster, namespace, resourceName, "bar")
		remoteWritten, err := subject.Write(remoteResource, clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		ExpectEqualAnotherMockResources(remoteWritten, remoteResource)

		// list local and remote resources, verify they are distinct
		localList, err = subject.List(namespace, clients.ListOpts{Ctx: ctx, Cluster: multicluster2.LocalCluster})
		Expect(err).NotTo(HaveOccurred())
		Expect(localList).To(HaveLen(1))
		ExpectEqualAnotherMockResources(localResource, localList[0])
		remoteList, err := subject.List(namespace, clients.ListOpts{Ctx: ctx, Cluster: remoteCluster})
		Expect(err).NotTo(HaveOccurred())
		Expect(remoteList).To(HaveLen(1))
		ExpectEqualAnotherMockResources(remoteResource, remoteList[0])
		Expect(remoteList).NotTo(Equal(localList))

		// read
		localRead, err := subject.Read(namespace, resourceName, clients.ReadOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		ExpectEqualAnotherMockResources(localResource, localRead)
		remoteRead, err := subject.Read(namespace, resourceName, clients.ReadOpts{Ctx: ctx, Cluster: remoteCluster})
		Expect(err).NotTo(HaveOccurred())
		ExpectEqualAnotherMockResources(remoteResource, remoteRead)

		// local watch
		w, errs, err := subject.Watch(namespace, clients.WatchOpts{
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
			ExpectEqualAnotherMockResources(list[0], localResource)
		}

		// remote watch
		w, errs, err = subject.Watch(namespace, clients.WatchOpts{
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
			ExpectEqualAnotherMockResources(list[0], remoteResource)
		}

		// aggregated watch
		aw, errs, err := watchAggregator.Watch(namespace, clients.WatchOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		Eventually(errs).ShouldNot(Receive())
		Eventually(aw).Should(Receive(And(ContainElement(localWritten), ContainElement(remoteWritten))))

		// delete
		err = subject.Delete(namespace, resourceName, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = subject.Delete(namespace, resourceName, clients.DeleteOpts{Ctx: ctx, Cluster: remoteCluster})
		Expect(err).NotTo(HaveOccurred())
	})
})

func ExpectEqualAnotherMockResources(a, b *v1.AnotherMockResource) {
	Expect(a.Metadata.Cluster).To(Equal(b.Metadata.Cluster))
	Expect(a.Metadata.Namespace).To(Equal(b.Metadata.Namespace))
	Expect(a.Metadata.Name).To(Equal(b.Metadata.Name))
	Expect(a.BasicField).To(Equal(b.BasicField))
}

// Wait for client to become available
func EventuallyClusterAvailable(client v1.AnotherMockResourceClient, cluster string) {
	Eventually(func() error {
		_, err := client.List("", clients.ListOpts{Cluster: cluster})
		return err
	}, 10*time.Second, 1*time.Second).Should(BeNil())
}
