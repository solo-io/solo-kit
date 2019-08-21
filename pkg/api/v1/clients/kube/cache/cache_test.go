package cache

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/log"
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

		BeforeEach(func() {
			var err error
			cfg, err = kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())

			ctx, cancel = context.WithCancel(context.TODO())
			client = kubernetes.NewForConfigOrDie(cfg)
		})

		Context("all namesapces", func() {

			BeforeEach(func() {
				var err error
				cache, err = NewKubeCoreCache(ctx, client)
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

		Context("single namesapces", func() {

			BeforeEach(func() {
				var err error
				cache, err = NewKubeCoreCacheWithOptions(ctx, client, time.Hour, []string{"default"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("can list resources for all listers", func() {
				Expect(cache.NamespaceLister()).To(BeNil())
				_, err := cache.NamespacedPodLister("default").List(selectors)
				Expect(err).NotTo(HaveOccurred())
				_, err = cache.NamespacedConfigMapLister("default").List(selectors)
				Expect(err).NotTo(HaveOccurred())
				_, err = cache.NamespacedSecretLister("default").List(selectors)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("2 namesapces", func() {
			Context("valid namesapces", func() {
				var (
					testns  string
					testns2 string
				)

				BeforeEach(func() {
					randomvalue := rand.Int31()
					testns = fmt.Sprintf("test-%d", randomvalue)
					testns2 = fmt.Sprintf("test2-%d", randomvalue)
					for _, ns := range []string{testns, testns2} {
						_, err := client.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
						Expect(err).NotTo(HaveOccurred())
						_, err = client.CoreV1().ConfigMaps(ns).Create(&v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cfg"}})
						Expect(err).NotTo(HaveOccurred())

					}
					var err error
					cache, err = NewKubeCoreCacheWithOptions(ctx, client, time.Hour, []string{testns, testns2})
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					for _, ns := range []string{testns, testns2} {
						client.CoreV1().Namespaces().Delete(ns, nil)
					}
				})

				It("can list resources for all listers", func() {
					Expect(cache.NamespaceLister()).To(BeNil())
					_, err := cache.NamespacedPodLister(testns).List(selectors)
					Expect(err).NotTo(HaveOccurred())
					cfgMaps, err := cache.NamespacedConfigMapLister(testns).List(selectors)
					Expect(err).NotTo(HaveOccurred())
					_, err = cache.NamespacedSecretLister(testns).List(selectors)
					Expect(err).NotTo(HaveOccurred())

					Expect(cache.NamespaceLister()).To(BeNil())
					_, err = cache.NamespacedPodLister(testns2).List(selectors)
					Expect(err).NotTo(HaveOccurred())
					cfgMaps2, err := cache.NamespacedConfigMapLister(testns2).List(selectors)
					Expect(err).NotTo(HaveOccurred())
					_, err = cache.NamespacedSecretLister(testns2).List(selectors)
					Expect(err).NotTo(HaveOccurred())

					Expect(cfgMaps).To(HaveLen(1))
					Expect(cfgMaps2).To(HaveLen(1))
					Expect(cfgMaps[0].Namespace).To(Equal(testns))
					Expect(cfgMaps2[0].Namespace).To(Equal(testns2))
				})
			})

			Context("Invalid namesapces", func() {
				It("should error with invalid namespace config", func() {
					var err error
					_, err = NewKubeCoreCacheWithOptions(ctx, client, time.Hour, []string{"default", ""})
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})

func init() {
	rand.Seed(time.Now().UnixNano())
}
