package kubesecret_test

import (
	"context"
	"os"
	"time"

	kubehelpers "github.com/solo-io/go-utils/testutils/kube"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	"github.com/solo-io/solo-kit/test/helpers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/tests/generic"
	"k8s.io/client-go/kubernetes"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("Base", func() {
	BeforeEach(func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			Skip("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		}
	})

	var (
		ns1, ns2       string
		kube           kubernetes.Interface
		client         *ResourceClient
		kubeCache      cache.KubeCoreCache
		localTestLabel string
	)
	BeforeEach(func() {
		ns1 = helpers.RandString(8)
		ns2 = helpers.RandString(8)
		localTestLabel = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(kube, ns1, ns2)
		kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client, err = NewResourceClient(kube, &v1.MockResource{}, false, kubeCache)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kube, ns1, ns2)
		Expect(err).NotTo(HaveOccurred())
		kubehelpers.WaitForNamespaceTeardown(ns1)
		kubehelpers.WaitForNamespaceTeardown(ns2)
	})
	It("CRUDs resources", func() {
		selectors := map[string]string{
			helpers.TestLabel: localTestLabel,
		}
		generic.TestCrudClient(ns1, ns2, client, clients.WatchOpts{
			Ctx:         context.TODO(),
			Selector:    selectors,
			RefreshRate: time.Minute,
		})
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
			client, err = NewResourceClient(kube, &v1.MockResource{}, false, kubeCache)
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
