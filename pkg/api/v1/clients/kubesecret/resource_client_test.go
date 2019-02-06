package kubesecret_test

import (
	"context"
	"os"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/test/helpers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/solo-kit/test/tests/generic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("Base", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace string
		cfg       *rest.Config
		client    *ResourceClient
		kubeCache cache.KubeCoreCache
	)
	BeforeEach(func() {
		namespace = helpers.RandString(8)
		err := setup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		kube, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		kubeCache = cache.NewKubeCoreCache(context.TODO(), kube)
		client, err = NewResourceClient(kube, &v1.MockResource{}, false, kubeCache)
	})
	AfterEach(func() {
		setup.TeardownKube(namespace)
	})
	It("CRUDs resources", func() {
		generic.TestCrudClient(namespace, client, time.Minute)
	})

	Context("multiple namespaces", func() {
		var (
			ns1, ns2 string
			cfg      *rest.Config
			client   *ResourceClient
		)
		BeforeEach(func() {
			ns1 = helpers.RandString(8)
			ns2 = helpers.RandString(8)
			err := setup.SetupKubeForTest(ns1)
			Expect(err).NotTo(HaveOccurred())
			err = setup.SetupKubeForTest(ns2)
			Expect(err).NotTo(HaveOccurred())

			cfg, err = kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())

			clientset, err := kubernetes.NewForConfig(cfg)
			Expect(err).NotTo(HaveOccurred())

			kubeCache = cache.NewKubeCoreCache(context.TODO(), clientset)
			client, err = NewResourceClient(clientset, &v1.MockResource{}, false, kubeCache)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			setup.TeardownKube(ns1)
			setup.TeardownKube(ns2)
		})
		It("can watch resources across namespaces when using NamespaceAll", func() {
			namespace := ""
			boo := "hoo"
			goo := "goo"
			data := "hi"

			err := client.Register()
			Expect(err).NotTo(HaveOccurred())

			w, errs, err := client.Watch(namespace, clients.WatchOpts{Ctx: context.TODO()})
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
					},
				}, clients.WriteOpts{})
				Expect(err).NotTo(HaveOccurred())

				r2, err = client.Write(&v1.MockResource{
					Data: data,
					Metadata: core.Metadata{
						Name:      goo,
						Namespace: ns2,
					},
				}, clients.WriteOpts{})
				Expect(err).NotTo(HaveOccurred())
			}()
			select {
			case <-wait:
			case <-time.After(time.Second * 5):
				Fail("expected wait to be closed before 5s")
			}

			list, err := client.List(namespace, clients.ListOpts{})
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

			var timesDrained int
		drain:
			for {
				select {
				case list = <-w:
					timesDrained++
					if timesDrained > 50 {
						Fail("drained the watch channel 50 times, something is wrong")
					}
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second / 4):
					break drain
				}
			}

			Expect(list).To(ContainElement(r1))
			Expect(list).To(ContainElement(r2))
		})
	})
})
