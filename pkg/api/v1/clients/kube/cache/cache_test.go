package cache

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/k8s-utils/kubeutils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ = Describe("kube core cache tests", func() {

	Context("kube tests", func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
			return
		}
		var (
			cfg    *rest.Config
			client *kubernetes.Clientset

			cache     *kubeCoreCaches
			ctx       context.Context
			cancel    context.CancelFunc
			selectors = labels.SelectorFromSet(make(map[string]string))
		)

		createNamespaceAndResource := func(namespace string) {
			_, err := client.CoreV1().Namespaces().Create(ctx, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cfg"}}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}

		validateNamespaceResource := func(namespace string) {
			_, err := cache.NamespacedPodLister(namespace).List(selectors)
			Expect(err).NotTo(HaveOccurred())
			cfgMap, err := cache.NamespacedConfigMapLister(namespace).List(selectors)
			Expect(err).NotTo(HaveOccurred())
			cfgMap = cleanConfigMaps(cfgMap)
			_, err = cache.NamespacedSecretLister(namespace).List(selectors)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfgMap).To(HaveLen(1))
			Expect(cfgMap[0].Namespace).To(Equal(namespace))
		}

		BeforeEach(func() {
			var err error
			cfg, err = kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())

			ctx, cancel = context.WithCancel(context.TODO())
			client = kubernetes.NewForConfigOrDie(cfg)
		})

		Context("all namespaces", func() {

			BeforeEach(func() {
				var err error
				cache, err = NewKubeCoreCache(ctx, client, true)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				cancel()
			})

			It("Allows for multiple subscribers", func() {
				watches := make([]<-chan struct{}, 3)
				for i := 0; i < 4; i++ {
					watches = append(watches, cache.Subscribe())
				}

				for _, watch := range watches {
					cache.Unsubscribe(watch)
					Expect(cache.cacheUpdatedWatchers).NotTo(ContainElement(watch))
				}
			})

			It("can list resources for all listers", func() {
				namespaces, err := cache.NamespaceLister().List(selectors)
				Expect(err).NotTo(HaveOccurred())
				Expect(namespaces).NotTo(HaveLen(0))
				_, err = cache.PodLister().List(selectors)
				Expect(err).NotTo(HaveOccurred())
				_, err = cache.ConfigMapLister().List(selectors)
				Expect(err).NotTo(HaveOccurred())
				_, err = cache.SecretLister().List(selectors)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("single namespaces", func() {

			BeforeEach(func() {
				var err error
				cache, err = NewKubeCoreCacheWithOptions(ctx, client, time.Hour, []string{"default"}, true)
				Expect(err).NotTo(HaveOccurred())
			})

			It("can list resources for all listers", func() {
				Expect(cache.NamespaceLister()).ToNot(BeNil())
				_, err := cache.NamespacedPodLister("default").List(selectors)
				Expect(err).NotTo(HaveOccurred())
				_, err = cache.NamespacedConfigMapLister("default").List(selectors)
				Expect(err).NotTo(HaveOccurred())
				_, err = cache.NamespacedSecretLister("default").List(selectors)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("2 namespaces", func() {
			Context("valid namespaces", func() {
				var (
					testns  string
					testns2 string
				)

				BeforeEach(func() {
					randomvalue := rand.Int31()
					testns = fmt.Sprintf("test-%d", randomvalue)
					testns2 = fmt.Sprintf("test2-%d", randomvalue)
					for _, ns := range []string{testns, testns2} {
						createNamespaceAndResource(ns)
					}
					var err error
					cache, err = NewKubeCoreCacheWithOptions(ctx, client, time.Hour, []string{testns, testns2}, true)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					for _, ns := range []string{testns, testns2} {
						client.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
					}
				})

				It("can list resources for all listers", func() {
					Expect(cache.NamespaceLister()).ToNot(BeNil())
					validateNamespaceResource(testns)
					validateNamespaceResource(testns2)
				})
			})

			Context("Invalid namespaces", func() {
				It("should error with invalid namespace config", func() {
					var err error
					_, err = NewKubeCoreCacheWithOptions(ctx, client, time.Hour, []string{"default", ""}, true)
					Expect(err).To(HaveOccurred())
				})
			})
			Context("Register a new namespace", func() {
				var (
					initialNs    string
					registeredNs string
				)

				BeforeEach(func() {
					randomvalue := rand.Int31()
					initialNs = fmt.Sprintf("initial-%d", randomvalue)
					registeredNs = fmt.Sprintf("registered-%d", randomvalue)

					createNamespaceAndResource(initialNs)

					var err error
					cache, err = NewKubeCoreCacheWithOptions(ctx, client, time.Hour, []string{initialNs}, true)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					client.CoreV1().Namespaces().Delete(ctx, initialNs, metav1.DeleteOptions{})
					client.CoreV1().Namespaces().Delete(ctx, registeredNs, metav1.DeleteOptions{})
				})

				It("should be able to register a new namespace", func() {
					createNamespaceAndResource(registeredNs)

					err := cache.RegisterNewNamespaceCache(registeredNs)
					Expect(err).NotTo(HaveOccurred())

					validateNamespaceResource(initialNs)
					validateNamespaceResource(registeredNs)
				})

				It("should be able to register a new namespace after the namespace was previously registered then deleted", func() {
					createNamespaceAndResource(registeredNs)

					err := cache.RegisterNewNamespaceCache(registeredNs)
					Expect(err).NotTo(HaveOccurred())

					validateNamespaceResource(initialNs)
					validateNamespaceResource(registeredNs)

					client.CoreV1().Namespaces().Delete(ctx, registeredNs, metav1.DeleteOptions{})
					// let the namespace be deleted
					Eventually(func() bool {
						_, err := client.CoreV1().Namespaces().Get(ctx, registeredNs, metav1.GetOptions{})
						return err != nil
					}, 10*time.Second, time.Second).Should(BeTrue())
					createNamespaceAndResource(registeredNs)
					// have to ensure that the configmap is created in the new namespace
					time.Sleep(50 * time.Millisecond)
					validateNamespaceResource(registeredNs)
				})
			})
		})
	})
})

func init() {
	rand.Seed(time.Now().UnixNano())
}

// remove the auto-generated kube-root-ca.crt ConfigMap, if it exists
func cleanConfigMaps(configMaps []*v1.ConfigMap) []*v1.ConfigMap {
	cleanedConfigMaps := make([]*v1.ConfigMap, 0, len(configMaps))
	for _, cfgMap := range configMaps {
		if cfgMap.GetName() != "kube-root-ca.crt" {
			cleanedConfigMaps = append(cleanedConfigMaps, cfgMap)
		}
	}
	return cleanedConfigMaps
}
