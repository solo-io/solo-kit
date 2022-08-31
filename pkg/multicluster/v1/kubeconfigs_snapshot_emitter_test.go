// Code generated by solo-kit. DO NOT EDIT.

// +build solokit

package v1

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
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"
	"github.com/solo-io/solo-kit/test/helpers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

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
		ctx                     context.Context
		namespace1, namespace2  string
		namespace3, namespace4  string
		namespace5, namespace6  string
		name1, name2            = "angela" + helpers.RandString(3), "bob" + helpers.RandString(3)
		name3, name4            = "susan" + helpers.RandString(3), "jim" + helpers.RandString(3)
		name5, name6            = "melisa" + helpers.RandString(3), "blake" + helpers.RandString(3)
		name7, name8            = "britany" + helpers.RandString(3), "john" + helpers.RandString(3)
		labels1                 = map[string]string{"env": "test"}
		labels2                 = map[string]string{"env": "testenv", "owner": "foo"}
		labelExpression1        = "env in (test)"
		kube                    kubernetes.Interface
		emitter                 KubeconfigsEmitter
		kubeConfigClient        KubeConfigClient
		resourceNamespaceLister resources.ResourceNamespaceLister
		kubeCache               cache.KubeCoreCache
	)
	const (
		TIME_BETWEEN_MESSAGES = 5
	)
	NewKubeConfigWithLabels := func(namespace, name string, labels map[string]string) *KubeConfig {
		resource := NewKubeConfig(namespace, name)
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

	deleteNonDefaultKubeNamespaces := func(ctx context.Context, kube kubernetes.Interface) {
		// clean up your local environment
		namespaces, err := kube.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		Expect(err).ToNot(HaveOccurred())
		defaultNamespaces := map[string]bool{"kube-node-lease": true, "kube-public": true, "kube-system": true, "local-path-storage": true, "default": true}
		var namespacesToDelete []string
		for _, ns := range namespaces.Items {
			if _, hit := defaultNamespaces[ns.Name]; !hit {
				namespacesToDelete = append(namespacesToDelete, ns.Name)
			}
		}
		err = kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespacesToDelete...)
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

		var snap *KubeconfigsSnapshot

		assertNoMessageSent := func() {
			for {
				select {
				case snap = <-snapshots:
					Fail("expected that no snapshots would be recieved " + log.Sprintf("%v", snap))
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 5):
					// this means that we have not recieved any mocks that we are not expecting
					return
				}
			}
		}

		/*
			KubeConfig
		*/
		assertSnapshotkubeconfigs := func(expectkubeconfigs KubeConfigList, unexpectkubeconfigs KubeConfigList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectkubeconfigs {
						if _, err := snap.Kubeconfigs.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectkubeconfigs {
						if _, err := snap.Kubeconfigs.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := kubeConfigClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := kubeConfigClient.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}

		kubeConfig1a, err := kubeConfigClient.Write(NewKubeConfig(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfig1b, err := kubeConfigClient.Write(NewKubeConfig(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigWatched := KubeConfigList{kubeConfig1a, kubeConfig1b}
		assertSnapshotkubeconfigs(kubeConfigWatched, nil)

		kubeConfig2a, err := kubeConfigClient.Write(NewKubeConfig(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfig2b, err := kubeConfigClient.Write(NewKubeConfig(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigWatched = append(kubeConfigWatched, KubeConfigList{kubeConfig2a, kubeConfig2b}...)
		assertSnapshotkubeconfigs(kubeConfigWatched, nil)

		kubeConfig3a, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace1, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfig3b, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace2, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigWatched = append(kubeConfigWatched, KubeConfigList{kubeConfig3a, kubeConfig3b}...)
		assertSnapshotkubeconfigs(kubeConfigWatched, nil)

		createNamespaceWithLabel(ctx, kube, namespace3, labels1)
		createNamespaces(ctx, kube, namespace4)

		kubeConfig4a, err := kubeConfigClient.Write(NewKubeConfig(namespace3, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfig4b, err := kubeConfigClient.Write(NewKubeConfig(namespace4, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigWatched = append(kubeConfigWatched, kubeConfig4a)
		kubeConfigNotWatched := KubeConfigList{kubeConfig4b}
		assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

		kubeConfig5a, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfig5b, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace4, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigWatched = append(kubeConfigWatched, kubeConfig5a)
		kubeConfigNotWatched = append(kubeConfigNotWatched, kubeConfig5b)
		assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

		kubeConfig6a, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace3, name3, labels2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfig6b, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace4, name3, labels2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigWatched = append(kubeConfigWatched, kubeConfig6a)
		kubeConfigNotWatched = append(kubeConfigNotWatched, kubeConfig6b)
		assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

		createNamespaceWithLabel(ctx, kube, namespace5, labels1)
		createNamespaces(ctx, kube, namespace6)

		kubeConfig7a, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace5, name1, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfig7b, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace6, name1, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigWatched = append(kubeConfigWatched, kubeConfig7a)
		kubeConfigNotWatched = append(kubeConfigNotWatched, kubeConfig7b)
		assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

		kubeConfig8a, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace6, name2, labels2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfig8b, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace6, name3, labels2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig8a, kubeConfig8b}...)
		assertNoMessageSent()

		for _, r := range kubeConfigNotWatched {
			err = kubeConfigClient.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
		}
		assertNoMessageSent()

		err = kubeConfigClient.Delete(kubeConfig1a.GetMetadata().Namespace, kubeConfig1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = kubeConfigClient.Delete(kubeConfig1b.GetMetadata().Namespace, kubeConfig1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig1a, kubeConfig1b}...)
		kubeConfigWatched = KubeConfigList{kubeConfig2a, kubeConfig2b, kubeConfig3a, kubeConfig3b, kubeConfig4a, kubeConfig5a, kubeConfig6a, kubeConfig7a}
		assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

		err = kubeConfigClient.Delete(kubeConfig2a.GetMetadata().Namespace, kubeConfig2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = kubeConfigClient.Delete(kubeConfig2b.GetMetadata().Namespace, kubeConfig2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig2a, kubeConfig2b}...)
		kubeConfigWatched = KubeConfigList{kubeConfig3a, kubeConfig3b, kubeConfig4a, kubeConfig5a, kubeConfig6a, kubeConfig7a}
		assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

		err = kubeConfigClient.Delete(kubeConfig3a.GetMetadata().Namespace, kubeConfig3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = kubeConfigClient.Delete(kubeConfig3b.GetMetadata().Namespace, kubeConfig3b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig3a, kubeConfig3b}...)
		kubeConfigWatched = KubeConfigList{kubeConfig4a, kubeConfig5a, kubeConfig6a, kubeConfig7a}
		assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

		err = kubeConfigClient.Delete(kubeConfig4a.GetMetadata().Namespace, kubeConfig4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = kubeConfigClient.Delete(kubeConfig5a.GetMetadata().Namespace, kubeConfig5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig5a, kubeConfig5b}...)
		kubeConfigWatched = KubeConfigList{kubeConfig6a, kubeConfig7a}
		assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

		err = kubeConfigClient.Delete(kubeConfig6a.GetMetadata().Namespace, kubeConfig6a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = kubeConfigClient.Delete(kubeConfig7a.GetMetadata().Namespace, kubeConfig7a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig6a, kubeConfig7a}...)
		assertSnapshotkubeconfigs(nil, kubeConfigNotWatched)

		// clean up environment
		deleteNamespaces(ctx, kube, namespace3, namespace4, namespace5, namespace6)
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
		// KubeConfig Constructor
		kubeConfigClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}

		kubeConfigClient, err = NewKubeConfigClient(ctx, kubeConfigClientFactory)
		Expect(err).NotTo(HaveOccurred())
		emitter = NewKubeconfigsEmitter(kubeConfigClient, resourceNamespaceLister)
	})
	AfterEach(func() {
		err := os.Unsetenv(statusutils.PodNamespaceEnvName)
		Expect(err).NotTo(HaveOccurred())

		deleteNonDefaultKubeNamespaces(ctx, kube)
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

			var snap *KubeconfigsSnapshot

			/*
				KubeConfig
			*/

			assertSnapshotkubeconfigs := func(expectkubeconfigs KubeConfigList, unexpectkubeconfigs KubeConfigList) {
			drain:
				for {
					select {
					case snap = <-snapshots:
						for _, expected := range expectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(expected.GetMetadata().Ref().Strings()); err != nil {
								continue drain
							}
						}
						for _, unexpected := range unexpectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
								continue drain
							}
						}
						break drain
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 10):
						nsList1, _ := kubeConfigClient.List(namespace1, clients.ListOpts{})
						nsList2, _ := kubeConfigClient.List(namespace2, clients.ListOpts{})
						combined := append(nsList1, nsList2...)
						Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
					}
				}
			}
			kubeConfig1a, err := kubeConfigClient.Write(NewKubeConfig(namespace1, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig1b, err := kubeConfigClient.Write(NewKubeConfig(namespace2, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())

			assertSnapshotkubeconfigs(KubeConfigList{kubeConfig1a, kubeConfig1b}, nil)
			kubeConfig2a, err := kubeConfigClient.Write(NewKubeConfig(namespace1, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig2b, err := kubeConfigClient.Write(NewKubeConfig(namespace2, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())

			assertSnapshotkubeconfigs(KubeConfigList{kubeConfig1a, kubeConfig1b, kubeConfig2a, kubeConfig2b}, nil)

			err = kubeConfigClient.Delete(kubeConfig2a.GetMetadata().Namespace, kubeConfig2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = kubeConfigClient.Delete(kubeConfig2b.GetMetadata().Namespace, kubeConfig2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())

			assertSnapshotkubeconfigs(KubeConfigList{kubeConfig1a, kubeConfig1b}, KubeConfigList{kubeConfig2a, kubeConfig2b})

			err = kubeConfigClient.Delete(kubeConfig1a.GetMetadata().Namespace, kubeConfig1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = kubeConfigClient.Delete(kubeConfig1b.GetMetadata().Namespace, kubeConfig1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())

			assertSnapshotkubeconfigs(nil, KubeConfigList{kubeConfig1a, kubeConfig1b, kubeConfig2a, kubeConfig2b})
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

			var snap *KubeconfigsSnapshot

			/*
				KubeConfig
			*/

			assertSnapshotkubeconfigs := func(expectkubeconfigs KubeConfigList, unexpectkubeconfigs KubeConfigList) {
			drain:
				for {
					select {
					case snap = <-snapshots:
						for _, expected := range expectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(expected.GetMetadata().Ref().Strings()); err != nil {
								continue drain
							}
						}
						for _, unexpected := range unexpectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
								continue drain
							}
						}
						break drain
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 10):
						nsList1, _ := kubeConfigClient.List(namespace1, clients.ListOpts{})
						nsList2, _ := kubeConfigClient.List(namespace2, clients.ListOpts{})
						combined := append(nsList1, nsList2...)
						Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
					}
				}
			}

			kubeConfig1a, err := kubeConfigClient.Write(NewKubeConfig(namespace1, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig1b, err := kubeConfigClient.Write(NewKubeConfig(namespace2, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			assertSnapshotkubeconfigs(KubeConfigList{kubeConfig1a, kubeConfig1b}, nil)

			kubeConfig2a, err := kubeConfigClient.Write(NewKubeConfig(namespace1, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig2b, err := kubeConfigClient.Write(NewKubeConfig(namespace2, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			assertSnapshotkubeconfigs(KubeConfigList{kubeConfig1a, kubeConfig1b, kubeConfig2a, kubeConfig2b}, nil)

			err = kubeConfigClient.Delete(kubeConfig2a.GetMetadata().Namespace, kubeConfig2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = kubeConfigClient.Delete(kubeConfig2b.GetMetadata().Namespace, kubeConfig2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			assertSnapshotkubeconfigs(KubeConfigList{kubeConfig1a, kubeConfig1b}, KubeConfigList{kubeConfig2a, kubeConfig2b})

			err = kubeConfigClient.Delete(kubeConfig1a.GetMetadata().Namespace, kubeConfig1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = kubeConfigClient.Delete(kubeConfig1b.GetMetadata().Namespace, kubeConfig1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			assertSnapshotkubeconfigs(nil, KubeConfigList{kubeConfig1a, kubeConfig1b, kubeConfig2a, kubeConfig2b})
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

			var snap *KubeconfigsSnapshot

			assertNoMessageSent := func() {
				for {
					select {
					case snap = <-snapshots:
						Fail("expected that no snapshots wouldbe recieved " + log.Sprintf("%v", snap))
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 5):
						// this means that we have not recieved any mocks that we are not expecting
						return
					}
				}
			}

			/*
				KubeConfig
			*/
			assertNokubeconfigsSent := func() {
			drain:
				for {
					select {
					case snap = <-snapshots:
						if len(snap.Kubeconfigs) == 0 {
							continue drain
						}
						Fail("expected that no snapshots containing resources would be recieved " + log.Sprintf("%v", snap))
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 5):
						// this means that we have not recieved any mocks that we are not expecting
						return
					}
				}
			}

			assertSnapshotkubeconfigs := func(expectkubeconfigs KubeConfigList, unexpectkubeconfigs KubeConfigList) {
			drain:
				for {
					select {
					case snap = <-snapshots:
						for _, expected := range expectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(expected.GetMetadata().Ref().Strings()); err != nil {
								continue drain
							}
						}
						for _, unexpected := range unexpectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
								continue drain
							}
						}
						break drain
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 10):
						nsList1, _ := kubeConfigClient.List(namespace1, clients.ListOpts{})
						nsList2, _ := kubeConfigClient.List(namespace2, clients.ListOpts{})
						combined := append(nsList1, nsList2...)
						Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
					}
				}
			}

			kubeConfig1a, err := kubeConfigClient.Write(NewKubeConfig(namespace1, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig1b, err := kubeConfigClient.Write(NewKubeConfig(namespace2, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigNotWatched := KubeConfigList{kubeConfig1a, kubeConfig1b}
			assertNokubeconfigsSent()

			createNamespaceWithLabel(ctx, kube, namespace3, labels1)
			createNamespaceWithLabel(ctx, kube, namespace4, labels1)

			kubeConfig2a, err := kubeConfigClient.Write(NewKubeConfig(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig2b, err := kubeConfigClient.Write(NewKubeConfig(namespace4, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigWatched := KubeConfigList{kubeConfig2a, kubeConfig2b}
			assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

			kubeConfig3a, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace1, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig3b, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace2, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig3a, kubeConfig3b}...)
			assertNokubeconfigsSent()

			kubeConfig4a, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig4b, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace4, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigWatched = append(kubeConfigWatched, KubeConfigList{kubeConfig4a, kubeConfig4b}...)
			assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

			createNamespaces(ctx, kube, namespace5, namespace6)

			kubeConfig5a, err := kubeConfigClient.Write(NewKubeConfig(namespace5, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig5b, err := kubeConfigClient.Write(NewKubeConfig(namespace6, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig5a, kubeConfig5b}...)
			assertNoMessageSent()

			kubeConfig6a, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace5, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig6b, err := kubeConfigClient.Write(NewKubeConfigWithLabels(namespace6, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig6a, kubeConfig6b}...)
			assertNoMessageSent()

			kubeConfig7a, err := kubeConfigClient.Write(NewKubeConfig(namespace5, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig7b, err := kubeConfigClient.Write(NewKubeConfig(namespace6, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig7a, kubeConfig7b}...)
			assertNoMessageSent()

			for _, r := range kubeConfigNotWatched {
				err = kubeConfigClient.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
			}
			assertNoMessageSent()

			err = kubeConfigClient.Delete(kubeConfig2a.GetMetadata().Namespace, kubeConfig2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = kubeConfigClient.Delete(kubeConfig2b.GetMetadata().Namespace, kubeConfig2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig2a, kubeConfig2b}...)
			kubeConfigWatched = KubeConfigList{kubeConfig4a, kubeConfig4b}
			assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

			err = kubeConfigClient.Delete(kubeConfig4a.GetMetadata().Namespace, kubeConfig4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = kubeConfigClient.Delete(kubeConfig4b.GetMetadata().Namespace, kubeConfig4b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigNotWatched = append(kubeConfigNotWatched, KubeConfigList{kubeConfig4a, kubeConfig4b}...)
			assertSnapshotkubeconfigs(nil, kubeConfigNotWatched)

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

			var snap *KubeconfigsSnapshot

			/*
				KubeConfig
			*/
			assertSnapshotkubeconfigs := func(expectkubeconfigs KubeConfigList, unexpectkubeconfigs KubeConfigList) {
			drain:
				for {
					select {
					case snap = <-snapshots:
						for _, expected := range expectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(expected.GetMetadata().Ref().Strings()); err != nil {
								continue drain
							}
						}
						for _, unexpected := range unexpectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
								continue drain
							}
						}
						break drain
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 10):
						nsList1, _ := kubeConfigClient.List(namespace1, clients.ListOpts{})
						nsList2, _ := kubeConfigClient.List(namespace2, clients.ListOpts{})
						combined := append(nsList1, nsList2...)
						Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
					}
				}
			}

			kubeConfig1a, err := kubeConfigClient.Write(NewKubeConfig(namespace1, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig1b, err := kubeConfigClient.Write(NewKubeConfig(namespace2, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigWatched := KubeConfigList{kubeConfig1a, kubeConfig1b}
			assertSnapshotkubeconfigs(kubeConfigWatched, nil)

			deleteNamespaces(ctx, kube, namespace1, namespace2)
			kubeConfigNotWatched := KubeConfigList{kubeConfig1a, kubeConfig1b}
			assertSnapshotkubeconfigs(nil, kubeConfigNotWatched)
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

			var snap *KubeconfigsSnapshot

			/*
				KubeConfig
			*/

			assertNokubeconfigsSent := func() {
			drain:
				for {
					select {
					case snap = <-snapshots:
						if len(snap.Kubeconfigs) == 0 {
							continue drain
						}
						Fail("expected that no snapshots containing resources would be recieved " + log.Sprintf("%v", snap))
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 5):
						// this means that we have not recieved any mocks that we are not expecting
						return
					}
				}
			}

			assertSnapshotkubeconfigs := func(expectkubeconfigs KubeConfigList, unexpectkubeconfigs KubeConfigList) {
			drain:
				for {
					select {
					case snap = <-snapshots:
						for _, expected := range expectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(expected.GetMetadata().Ref().Strings()); err != nil {
								continue drain
							}
						}
						for _, unexpected := range unexpectkubeconfigs {
							if _, err := snap.Kubeconfigs.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
								continue drain
							}
						}
						break drain
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 10):
						nsList1, _ := kubeConfigClient.List(namespace1, clients.ListOpts{})
						nsList2, _ := kubeConfigClient.List(namespace2, clients.ListOpts{})
						combined := append(nsList1, nsList2...)
						Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
					}
				}
			}

			kubeConfig1a, err := kubeConfigClient.Write(NewKubeConfig(namespace1, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig1b, err := kubeConfigClient.Write(NewKubeConfig(namespace2, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigNotWatched := KubeConfigList{kubeConfig1a, kubeConfig1b}
			assertNokubeconfigsSent()

			deleteNamespaces(ctx, kube, namespace1, namespace2)
			assertNokubeconfigsSent()

			// create namespaces
			createNamespaceWithLabel(ctx, kube, namespace3, labels1)
			createNamespaceWithLabel(ctx, kube, namespace4, labels1)

			kubeConfig2a, err := kubeConfigClient.Write(NewKubeConfig(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfig2b, err := kubeConfigClient.Write(NewKubeConfig(namespace4, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigWatched := KubeConfigList{kubeConfig2a, kubeConfig2b}
			assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

			deleteNamespaces(ctx, kube, namespace3)
			kubeConfigWatched = KubeConfigList{kubeConfig2b}
			kubeConfigNotWatched = append(kubeConfigNotWatched, kubeConfig2a)
			assertSnapshotClusterresources(kubeConfigWatched, kubeConfigNotWatched)

			createNamespaceWithLabel(ctx, kube, namespace5, labels1)

			kubeConfig3a, err := kubeConfigClient.Write(NewKubeConfig(namespace5, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigWatched = append(kubeConfigWatched, kubeConfig3a)
			assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

			deleteNamespaces(ctx, kube, namespace4)
			kubeConfigNotWatched = append(kubeConfigNotWatched, kubeConfig2b)
			kubeConfigWatched = KubeConfigList{kubeConfig3a}
			assertSnapshotkubeconfigs(kubeConfigWatched, kubeConfigNotWatched)

			for _, r := range kubeConfigWatched {
				err = kubeConfigClient.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				kubeConfigNotWatched = append(kubeConfigNotWatched, r)
			}
			assertSnapshotkubeconfigs(nil, kubeConfigNotWatched)

			deleteNamespaces(ctx, kube, namespace5)

		})
	})

	Context("use different resource namespace listers", func() {
		BeforeEach(func() {
			resourceNamespaceLister = namespace.NewKubeClientResourceNamespaceLister(kube)
			emitter = NewKubeconfigsEmitter(kubeConfigClient, resourceNamespaceLister)
		})

		It("Should work with the Kube Client Namespace Lister", func() {
			runNamespacedSelectorsWithWatchNamespaces()
		})
	})

})
