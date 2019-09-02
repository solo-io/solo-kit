package kube_test

import (
	"context"
	"runtime"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned/fake"
	solov1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	mocksv1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/util"
)

var _ = Describe("Test ResourceClientSharedInformerFactory", func() {

	const (
		namespace1 = "test-ns-1"
		namespace2 = "test-ns-2"
		namespace3 = "test-ns-3"
	)

	var (
		kubeCache                   *kube.ResourceClientSharedInformerFactory
		client1, client2, client123 *kube.ResourceClient
		ctx                         context.Context
		cancel                      func()
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.TODO())
		kubeCache = kube.NewKubeCache(ctx).(*kube.ResourceClientSharedInformerFactory)
		Expect(len(kubeCache.Informers())).To(BeZero())

		client1 = util.MockClientForNamespace(kubeCache, []string{namespace1})
		client2 = util.MockClientForNamespace(kubeCache, []string{namespace2})
		client123 = util.MockClientForNamespace(kubeCache, []string{namespace1, namespace2, namespace3})
	})

	Describe("registering resource clients with the factory", func() {

		It("correctly registers a single client", func() {
			err := kubeCache.Register(client1)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(kubeCache.Informers())).To(BeEquivalentTo(1))
		})

		It("correctly registers multiple clients", func() {
			err := kubeCache.Register(client1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(kubeCache.Informers())).To(BeEquivalentTo(1))

			err = kubeCache.Register(client2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(kubeCache.Informers())).To(BeEquivalentTo(2))
		})

		It("creates an informer for each namespace in the client whitelist", func() {
			err := kubeCache.Register(client123)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(kubeCache.Informers())).To(BeEquivalentTo(3))
		})

		It("errors when attempting to register multiple clients for the same resource and namespace", func() {
			err := kubeCache.Register(client1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(kubeCache.Informers())).To(BeEquivalentTo(1))

			err = kubeCache.Register(&*client1)
			Expect(err).To(HaveOccurred())
			Expect(len(kubeCache.Informers())).To(BeEquivalentTo(1))
		})

		It("panics when attempting of register a client with a running factory", func() {
			// Start without registering clients, just to set the "started" flag
			kubeCache.Start()
			Expect(kubeCache.IsRunning()).To(BeTrue())

			Expect(func() { _ = kubeCache.Register(client1) }).To(Panic())
		})
	})

	Describe("starting the factory", func() {

		It("starts without errors", func() {
			err := kubeCache.Register(client1)
			Expect(err).NotTo(HaveOccurred())

			kubeCache.Start()

			Expect(kubeCache.IsRunning()).To(BeTrue())
		})

		It("start operation is idempotent", func() {
			err := kubeCache.Register(client1)
			Expect(err).NotTo(HaveOccurred())

			kubeCache.Start()
			kubeCache.Start()
			kubeCache.Start()

			Expect(kubeCache.IsRunning()).To(BeTrue())
		})
	})

	Describe("creating watches", func() {

		var (
			clientset          *fake.Clientset
			preStartGoroutines int
		)

		BeforeEach(func() {
			clientset = fake.NewSimpleClientset(mocksv1.MockResourceCrd)
			// We need the resourceClient so that we can register its resourceType/namespaces with the cache
			client := util.ClientForClientsetAndResource(clientset, kubeCache, mocksv1.MockResourceCrd, &mocksv1.MockResource{}, []string{namespace1})
			err := kubeCache.Register(client)
			Expect(err).NotTo(HaveOccurred())

			preStartGoroutines = runtime.NumGoroutine()
			kubeCache.Start()
			Expect(kubeCache.IsRunning()).To(BeTrue())
		})

		Context("a single watch", func() {
			var watch <-chan solov1.Resource

			BeforeEach(func() {
				watch = kubeCache.AddWatch(10)
			})

			It("receives an event in that namespace", func() {

				// Add a resource in a separate goroutine
				go Expect(util.CreateMockResource(clientset, namespace1, "mock-res-1", "test")).To(BeNil())

				select {
				case res := <-watch:
					Expect(res.Namespace).To(BeEquivalentTo(namespace1))
					Expect(res.Name).To(BeEquivalentTo("mock-res-1"))
					Expect(res.Kind).To(BeEquivalentTo("MockResource"))
					return
				case <-time.After(1 * time.Second):
					Fail("timed out waiting for watch event")
					return
				}
			})

			It("ignores an event in a different namespace", func() {

				// Add a resource in a different namespace
				go Expect(util.CreateMockResource(clientset, namespace2, "mock-res-1", "test")).To(BeNil())

				select {
				case <-watch:
					Fail("received event for non-watched namespace")
				case <-time.After(1 * time.Second):
					return
				}

			})

			It("correctly handles multiple events", func() {

				var watchResults []string

				ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*100))

				go func() {
					for {
						select {
						case <-ctx.Done():
							return
						case res := <-watch:
							watchResults = append(watchResults, res.ObjectMeta.Name)
						}
					}
				}()

				go Expect(util.CreateMockResource(clientset, namespace1, "mock-res-1", "test")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace2, "mock-res-2", "test")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace1, "mock-res-3", "test")).To(BeNil())
				go Expect(util.DeleteMockResource(clientset, namespace1, "mock-res-1")).To(BeNil())

				<-ctx.Done()

				Expect(len(watchResults)).To(BeEquivalentTo(3))
				Expect(watchResults).To(ConsistOf("mock-res-1", "mock-res-3", "mock-res-1"))
			})
		})

		Context("multiple watches", func() {

			var watches []<-chan solov1.Resource

			BeforeEach(func() {
				watches = []<-chan solov1.Resource{
					kubeCache.AddWatch(10),
					kubeCache.AddWatch(10),
					kubeCache.AddWatch(10),
				}
			})

			It("all watches receive the same events", func() {

				watchResults := [][]string{{}, {}, {}}
				ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*100))
				for i, watch := range watches {
					go func(index int, watchChan <-chan solov1.Resource) {
						for {
							select {
							case <-ctx.Done():
								return
							case res := <-watchChan:
								watchResults[index] = append(watchResults[index], res.ObjectMeta.Name)
							}
						}
					}(i, watch)
				}

				go Expect(util.CreateMockResource(clientset, namespace1, "mock-res-1", "test")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace2, "mock-res-2", "test")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace1, "mock-res-3", "test")).To(BeNil())
				go Expect(util.DeleteMockResource(clientset, namespace1, "mock-res-1")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace1, "mock-res-4", "test")).To(BeNil())
				go Expect(util.DeleteMockResource(clientset, namespace2, "mock-res-2")).To(BeNil())

				// Wait for results to be collected
				time.Sleep(100 * time.Millisecond)

				<-ctx.Done()

				for _, watchResult := range watchResults {
					Expect(len(watchResult)).To(BeEquivalentTo(4))
					Expect(watchResult).To(ConsistOf("mock-res-1", "mock-res-3", "mock-res-1", "mock-res-4"))
				}
			})
		})

		Context("context cancellation", func() {

			var watches []<-chan solov1.Resource

			BeforeEach(func() {
				watches = []<-chan solov1.Resource{
					kubeCache.AddWatch(10),
					kubeCache.AddWatch(10),
					kubeCache.AddWatch(10),
				}
			})

			It("watches stop receiving events after the factory's context is cancelled", func() {

				watchResults := [][]string{{}, {}, {}}
				l := sync.Mutex{}
				for i, watch := range watches {
					preStartGoroutines++
					go func(index int, watchChan <-chan solov1.Resource) {
						for {
							select {
							case res := <-watchChan:
								l.Lock()
								watchResults[index] = append(watchResults[index], res.ObjectMeta.Name)
								l.Unlock()
							}
						}
					}(i, watch)
				}

				go Expect(util.CreateMockResource(clientset, namespace1, "mock-res-1", "test")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace2, "mock-res-2", "test")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace1, "mock-res-3", "test")).To(BeNil())
				go Expect(util.DeleteMockResource(clientset, namespace1, "mock-res-1")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace1, "mock-res-4", "test")).To(BeNil())
				go Expect(util.DeleteMockResource(clientset, namespace2, "mock-res-2")).To(BeNil())

				for i := range watchResults {
					Eventually(func() int {
						l.Lock()
						defer l.Unlock()
						watchResult := watchResults[i]
						return len(watchResult)
					}).Should(BeEquivalentTo(4))
					l.Lock()
					Expect(watchResults[i]).To(ConsistOf("mock-res-1", "mock-res-3", "mock-res-1", "mock-res-4"))
					l.Unlock()
				}

				// cancel the context! zbam
				cancel()
				Eventually(func() int {
					return runtime.NumGoroutine()
				}, time.Second).Should(Equal(preStartGoroutines))

				go Expect(util.CreateMockResource(clientset, namespace1, "another-mock-res-1", "test")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace2, "another-mock-res-2", "test")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace1, "another-mock-res-3", "test")).To(BeNil())
				go Expect(util.DeleteMockResource(clientset, namespace1, "another-mock-res-1")).To(BeNil())
				go Expect(util.CreateMockResource(clientset, namespace1, "another-mock-res-4", "test")).To(BeNil())
				go Expect(util.DeleteMockResource(clientset, namespace2, "another-mock-res-2")).To(BeNil())

				for i := range watchResults {
					Eventually(func() int {
						l.Lock()
						defer l.Unlock()
						watchResult := watchResults[i]
						return len(watchResult)
					}).Should(BeEquivalentTo(4))
					l.Lock()
					Expect(watchResults[i]).NotTo(ConsistOf("another-mock-res-1"))
					Expect(watchResults[i]).NotTo(ConsistOf("another-mock-res-2"))
					Expect(watchResults[i]).NotTo(ConsistOf("another-mock-res-3"))
					Expect(watchResults[i]).NotTo(ConsistOf("another-mock-res-4"))
					l.Unlock()
				}
			})
		})
	})
})
