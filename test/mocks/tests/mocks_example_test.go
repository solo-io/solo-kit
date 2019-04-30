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
		ctrl    *gomock.Controller
		ctx     context.Context
		cancel  context.CancelFunc
		syncer  *mocks.MockTestingSyncer
		emitter *mocks.MockTestingEmitter
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
	})

	It("can create an event loop with a mock emitter and syncer, and error out", func() {
		el := v1.NewTestingEventLoop(emitter, syncer)

		watchNamespaces := []string{"namespace1"}
		// need to make sure args match
		emitter.EXPECT().Snapshots(watchNamespaces, gomock.Any()).Times(1).Return(nil, nil, exampleError)

		_, err := el.Run(watchNamespaces, clients.WatchOpts{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(exampleError.Error()))
		cancel()
	})
	It("can simulate a real event loop with a mocked channel", func() {

		watchNamespaces := []string{"namespace1"}

		watch := make(chan *v1.TestingSnapshot)
		snap := &v1.TestingSnapshot{}
		emitter.EXPECT().Snapshots(watchNamespaces, gomock.Any()).Times(1).Return(watch, nil, nil)
		syncerHasBeenCalled := false
		syncer.EXPECT().Sync(gomock.Any(), snap).Return(nil).Do(func(ctx context.Context, snap *v1.TestingSnapshot) {
			syncerHasBeenCalled = true
		}).Times(1)

		el := v1.NewTestingEventLoop(emitter, syncer)


		_, err := el.Run(watchNamespaces, clients.WatchOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		// send snap to el
		watch <- snap
		Eventually(func() bool {
			return syncerHasBeenCalled
		}, time.Second*2, time.Millisecond * 100).Should(BeTrue())
		cancel()
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
		watch <- snap

		// wait for error to be returned
		select {
		case err, ok := <- errs:
			Expect(ok).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring(exampleError.Error()))
		}


		cancel()
	})
})
