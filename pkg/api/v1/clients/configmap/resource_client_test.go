package configmap_test

import (
	"context"
	"os"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/go-utils/kubeutils"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/solo-io/solo-kit/test/mocks/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/configmap"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/generic"
	"k8s.io/client-go/kubernetes"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("Base", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		ns1            string
		kube           kubernetes.Interface
		client         *ResourceClient
		kubeCache      cache.KubeCoreCache
		localTestLabel string
	)
	BeforeEach(func() {
		ns1 = helpers.RandString(8)
		localTestLabel = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(kube, ns1)
		kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client, err = NewResourceClient(kube, &v1.MockResource{}, kubeCache, false)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kube, ns1)
		Expect(err).NotTo(HaveOccurred())
	})
	It("CRUDs resources", func() {
		selector := map[string]string{
			helpers.TestLabel: localTestLabel,
		}
		generic.TestCrudClient(ns1, client, clients.WatchOpts{
			Selector:    selector,
			Ctx:         context.TODO(),
			RefreshRate: time.Minute,
		})
	})
	It("uses json keys when serializing", func() {
		foo := "test-data-keys"
		input := v1.NewMockResource(ns1, foo)
		data := "hello: goodbye"
		input.Data = data
		labels := map[string]string{"pick": "me"}
		input.Metadata.Labels = labels

		err := client.Register()
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Write(input, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		cm, err := kube.CoreV1().ConfigMaps(input.Metadata.Namespace).Get(input.Metadata.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(cm.Data).To(HaveKey("data.json"))
		Expect(cm.Data["data.json"]).To(ContainSubstring("'hello: goodbye'"))
	})

	Context("multiple namespaces", func() {
		var (
			ns2 string
		)
		BeforeEach(func() {
			ns2 = helpers.RandString(8)

			err := kubeutils.CreateNamespacesInParallel(kube, ns2)
			Expect(err).NotTo(HaveOccurred())

			kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
			Expect(err).NotTo(HaveOccurred())
			client, err = NewResourceClient(kube, &v1.MockResource{}, kubeCache, false)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := kubeutils.DeleteNamespacesInParallelBlocking(kube, ns2)
			Expect(err).NotTo(HaveOccurred())
		})
		It("can watch resources across namespaces when using NamespaceAll", func() {
			watchNamespace := ""
			selectors := map[string]string{helpers.TestLabel: localTestLabel}
			boo := "hoo"
			goo := "goo"
			data := "hi"

			err := client.Register()
			Expect(err).NotTo(HaveOccurred())

			w, errs, err := client.Watch(watchNamespace, clients.WatchOpts{Ctx: context.TODO(), Selector: selectors})
			Expect(err).NotTo(HaveOccurred())

			var r1, r2 resources.Resource
			wait := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				defer func() {
					close(wait)
				}()
				r1, err = client.Write(&v1.MockResource{
					Data: data,
					Metadata: core.Metadata{
						Name:      boo,
						Namespace: ns1,
						Labels:    selectors,
					},
				}, clients.WriteOpts{})
				Expect(err).NotTo(HaveOccurred())

				r2, err = client.Write(&v1.MockResource{
					Data: data,
					Metadata: core.Metadata{
						Name:      goo,
						Namespace: ns2,
						Labels:    selectors,
					},
				}, clients.WriteOpts{})
				Expect(err).NotTo(HaveOccurred())
			}()
			select {
			case <-wait:
			case <-time.After(time.Second * 5):
				Fail("expected wait to be closed before 5s")
			}

			list, err := client.List(watchNamespace, clients.ListOpts{
				Selector: selectors,
				Ctx:      context.TODO(),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(ContainElement(r1))
			Expect(list).To(ContainElement(r2))

			select {
			case err := <-errs:
				Expect(err).NotTo(HaveOccurred())
			case list = <-w:
			case <-time.After(time.Millisecond * 5):
				Fail("expected a message in channel")
			}

			go func() {
				defer GinkgoRecover()
				for {
					select {
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second / 4):
						return
					}
				}
			}()

			Eventually(w, time.Second*5, time.Second/10).Should(Receive(And(ContainElement(r1), ContainElement(r2))))
		})
	})
})
