package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/mocks/v1/mocks"
)

var _ = Describe("examples using mocks", func() {

	var (
		ctrl         *gomock.Controller
		ctx          context.Context
		cancel       context.CancelFunc
		syncer       *mocks.MockTestingSyncer
		emitter      *mocks.MockTestingEmitter
		exampleError = fmt.Errorf("example error")
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.TODO())
		ctrl = gomock.NewController(T)
		emitter = mocks.NewMockTestingEmitter(ctrl)
		syncer = mocks.NewMockTestingSyncer(ctrl)
	})
	AfterEach(func() {
		ctrl.Finish()
		cancel()
	})

	It("can create an event loop with a mock emitter and syncer, and error out", func() {
		el := v1.NewTestingEventLoop(emitter, syncer)

		watchNamespaces := []string{"namespace1"}
		// need to make sure args match
		emitter.EXPECT().Snapshots(watchNamespaces, gomock.Any()).Times(1).Return(nil, nil, exampleError)

		_, err := el.Run(watchNamespaces, clients.WatchOpts{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(exampleError.Error()))
	})
	It("can simulate a real event loop with a mocked channel", func() {

		watchNamespaces := []string{"namespace1"}

		watch := make(chan *v1.TestingSnapshot, 10)
		snap := &v1.TestingSnapshot{}
		emitter.EXPECT().Snapshots(watchNamespaces, gomock.Any()).Times(1).Return(watch, nil, nil)
		ackChan := make(chan bool, 10)
		syncer.EXPECT().Sync(gomock.Any(), snap).DoAndReturn(func(ctx context.Context, snap *v1.TestingSnapshot) error {
			select {
			case ackChan <- true:
			case <-time.After(1 * time.Second):
				Fail("waited to long to send message on channel")
			}
			return nil
		}).Times(1)

		el := v1.NewTestingEventLoop(emitter, syncer)

		_, err := el.Run(watchNamespaces, clients.WatchOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		// send snap to el

		go func() {
			defer GinkgoRecover()
			select {
			case watch <- snap:
			case <-time.After(1 * time.Second):
				Fail("could not send message to channel within one second")
			}
		}()

		Eventually(ackChan, time.Second*2, time.Millisecond*100).Should(Receive(Equal(true)))
	})

	It("can simulate an error happening in the sync function", func() {

		watchNamespaces := []string{"namespace1"}

		watch := make(chan *v1.TestingSnapshot)
		snap := &v1.TestingSnapshot{}
		emitter.EXPECT().Snapshots(watchNamespaces, gomock.Any()).Times(1).Return(watch, nil, nil)
		syncer.EXPECT().Sync(gomock.Any(), snap).Return(exampleError).Times(1)

		el := v1.NewTestingEventLoop(emitter, syncer)

		errs, err := el.Run(watchNamespaces, clients.WatchOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		// send snap to el
		go func() {
			defer GinkgoRecover()
			select {
			case watch <- snap:
			case <-time.After(1 * time.Second):
				Fail("could not send message to channel within one second")
			}
		}()

		// wait for error to be returned
		select {
		case err, ok := <-errs:
			Expect(ok).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring(exampleError.Error()))
		case <-time.After(2 * time.Second):
			Fail("waited to long for error to appear")
		}
	})
})
