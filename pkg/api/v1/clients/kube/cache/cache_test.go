package cache

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/log"
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

})
