package controller_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned/fake"
	solov1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	mocksv1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// The setup for these tests is intentionally redundant. Given the many nested closures and the nature of the objects
// to be tested, having global variables to be reused between tests increases the chance of errors significantly.
var _ = Describe("Test KubeController", func() {

	const (
		namespace1 = "test-ns-1"
		namespace2 = "test-ns-2"
	)

	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("one registered informer", func() {

		var (
			kubeController *controller.Controller
			resultChan     chan solov1.Resource
			clientset      *fake.Clientset
			resyncPeriod   time.Duration
			stopChan       chan struct{}
			err            error
		)

		const (
			name1  = "res-1"
			value1 = "test"
			name2  = "res-2"
			value2 = "secondNamespaceValue"
		)

		getResultFromkMockResource := func(namespace, name, value string) {
			select {
			case res := <-resultChan:
				Expect(res.Namespace).To(BeEquivalentTo(namespace))
				Expect(res.Name).To(BeEquivalentTo(name))
				Expect(res.Kind).To(BeEquivalentTo("MockResource"))
				Expect(res.Spec).To(Not(BeNil()))

				fieldValue, ok := (*res.Spec)["someDumbField"]
				Expect(ok).To(BeTrue())
				Expect(fieldValue).To(BeEquivalentTo(value))
			case <-time.After(50 * time.Millisecond):
				Fail("timed out waiting for watch event")
			}
		}

		BeforeEach(func() {
			clientset = fake.NewSimpleClientset(mocksv1.MockResourceCrd)
			resyncPeriod = time.Duration(0)
			resultChan = make(chan solov1.Resource, 100)
			stopChan = make(chan struct{})

			kubeController = controller.NewController(
				"test-controller",
				controller.NewLockingCallbackHandler(func(resource solov1.Resource) {
					// block until someone receives from the channel
					resultChan <- resource
				}),
				cache.NewSharedIndexInformer(
					listWatchForClientAndNamespace(ctx, clientset, namespace1),
					&solov1.Resource{},
					resyncPeriod,
					cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
			)
			err = kubeController.Run(2, stopChan)
		})

		AfterEach(func() {
			close(stopChan)
		})

		It("controller starts correctly", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("sends the correct notification through the event handler", func() {
			err = util.CreateMockResource(ctx, clientset, namespace1, "res-1", "test")
			Expect(err).NotTo(HaveOccurred())

			getResultFromkMockResource(namespace1, name1, value1)
		})

		It("does not react to events in a non relevant namespace", func() {
			err = util.CreateMockResource(ctx, clientset, "ns-X", "res-1", "test")
			Expect(err).NotTo(HaveOccurred())

			select {
			case <-resultChan:
				Fail("should not have received event")
				return
			case <-time.After(100 * time.Millisecond):
				Succeed()
			}
		})

		It("can add new informers so that events can be received on the new informer", func() {
			err = util.CreateMockResource(ctx, clientset, namespace1, name1, value1)
			Expect(err).NotTo(HaveOccurred())

			newInformer := cache.NewSharedIndexInformer(
				listWatchForClientAndNamespace(ctx, clientset, namespace2),
				&solov1.Resource{},
				resyncPeriod,
				cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			)

			getResultFromkMockResource(namespace1, name1, value1)

			// create the second value that we want to look at, but do not set up the informer quit
			// yet, we still want to ensure that the controller does not learn about the new resource
			// until the new informer has been added to the kube controller
			err = util.CreateMockResource(ctx, clientset, namespace2, name2, value2)
			Expect(err).NotTo(HaveOccurred())

			select {
			case res := <-resultChan:
				Fail(fmt.Sprintf("Should not have received the resource %s from Namespace %s as the informer has not yet been added to the KubeController yet", res.Name, res.Namespace))
			case <-time.After(100 * time.Millisecond):
			}

			err = kubeController.AddNewOfInformers(newInformer)
			Expect(err).NotTo(HaveOccurred())

			getResultFromkMockResource(namespace2, name2, value2)
		})
	})

	Context("controller is configured with a resync period", func() {

		var (
			kubeController *controller.Controller
			resultChan     chan solov1.Resource
			clientset      *fake.Clientset
			resyncPeriod   time.Duration
			stopChan       chan struct{}
		)

		BeforeEach(func() {
			clientset = fake.NewSimpleClientset(mocksv1.MockResourceCrd)
			resyncPeriod = time.Second
			resultChan = make(chan solov1.Resource, 100)
			stopChan = make(chan struct{})

			kubeController = controller.NewController(
				"test-controller",
				controller.NewLockingCallbackHandler(func(resource solov1.Resource) {
					// block until someone receives from the channel
					resultChan <- resource
				}),
				cache.NewSharedIndexInformer(
					listWatchForClientAndNamespace(ctx, clientset, namespace1),
					&solov1.Resource{},
					resyncPeriod,
					cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
			)
			Expect(kubeController.Run(2, stopChan)).To(BeNil())
		})

		AfterEach(func() {
			close(stopChan)
		})

		It("resyncs every period", func() {
			// Put an object into the store so the ListWatch has something to list
			Expect(util.CreateMockResource(ctx, clientset, namespace1, "res-1", "test")).To(BeNil())
			// drain channel from creation event to have accurate resync count
			<-resultChan

			count := 0
			after := time.After(2200 * time.Millisecond)
		LOOP:
			for {
				select {
				case <-resultChan:
					count = count + 1
				case <-after:
					break LOOP
				}
			}

			Expect(count).To(BeEquivalentTo(2))
		})
	})

	Context("two registered informers", func() {

		var (
			kubeController         *controller.Controller
			resultChan             chan solov1.Resource
			clientset1, clientset2 *fake.Clientset
			resyncPeriod           time.Duration
			stopChan               chan struct{}
		)

		BeforeEach(func() {
			clientset1 = fake.NewSimpleClientset(mocksv1.MockResourceCrd)
			clientset2 = fake.NewSimpleClientset(mocksv1.MockResourceCrd)

			resyncPeriod = time.Duration(0)
			resultChan = make(chan solov1.Resource, 100)
			stopChan = make(chan struct{})

			kubeController = controller.NewController(
				"test-controller",
				controller.NewLockingCallbackHandler(func(resource solov1.Resource) {
					// block until someone receives from the channel
					resultChan <- resource
				}),
				informerWith(listWatchForClientAndNamespace(ctx, clientset1, namespace1), resyncPeriod),
				informerWith(listWatchForClientAndNamespace(ctx, clientset2, namespace2), resyncPeriod),
			)

			Expect(kubeController.Run(2, stopChan)).To(BeNil())
		})

		AfterEach(func() {
			close(stopChan)
		})

		It("correctly sends notification for both informers", func() {
			Expect(util.CreateMockResource(ctx, clientset1, namespace1, "res-1", "test-1")).To(BeNil())
			Expect(util.CreateMockResource(ctx, clientset2, namespace2, "res-2", "test-2")).To(BeNil())

			results := make(map[string]solov1.Resource)
			after := time.After(100 * time.Millisecond)
		LOOP:
			for {
				select {
				case res := <-resultChan:
					results[res.ObjectMeta.Namespace] = res
				case <-after:
					break LOOP
				}
			}

			Expect(results).To(HaveLen(2))

			res1, ok := results[namespace1]
			Expect(ok).To(BeTrue())
			Expect(res1.Namespace).To(BeEquivalentTo(namespace1))
			Expect(res1.Name).To(BeEquivalentTo("res-1"))
			Expect(res1.Kind).To(BeEquivalentTo("MockResource"))

			res2, ok := results[namespace2]
			Expect(ok).To(BeTrue())
			Expect(res2.Namespace).To(BeEquivalentTo(namespace2))
			Expect(res2.Name).To(BeEquivalentTo("res-2"))
			Expect(res2.Kind).To(BeEquivalentTo("MockResource"))
		})
	})

})

func listWatchForClientAndNamespace(ctx context.Context, clientset *fake.Clientset, namespace string) *cache.ListWatch {
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.ResourcesV1().Resources(namespace).List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.ResourcesV1().Resources(namespace).Watch(ctx, options)
		},
	}
}

func informerWith(listWatch *cache.ListWatch, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		listWatch,
		&solov1.Resource{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}
