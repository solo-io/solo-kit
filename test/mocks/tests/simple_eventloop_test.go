package tests_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("SimpleEventLoop", func() {
	It("sends snapshots from a single aggregated watch", func() {
		emitter := &mockEmitter{}
		shouldSyncer := &mockSyncer{shouldSync: true}
		shouldNotSyncer := &mockSyncer{}

		eventLoop := v1.NewTestingSimpleEventLoop(emitter, shouldSyncer, shouldNotSyncer)

		ctx, cancel := context.WithCancel(context.TODO())

		go func() {
			defer GinkgoRecover()
			errs, err := eventLoop.Run(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(<-errs).NotTo(HaveOccurred())
		}()

		Eventually(func() int {
			return shouldSyncer.syncs
		}).Should(Equal(50))

		Eventually(func() int {
			return shouldSyncer.cancels
		}).Should(Equal(49))

		cancel()
		Eventually(func() int {
			return shouldSyncer.cancels
		}).Should(Equal(50))

		Expect(shouldNotSyncer.syncs).To(Equal(0))
	})
})

type (
	mockSyncer struct {
		shouldSync bool
		syncs      int
		cancels    int
	}
	mockEmitter struct{}
)

func (e *mockEmitter) Snapshots(ctx context.Context) (<-chan *v1.TestingSnapshot, <-chan error, error) {
	snaps := make(chan *v1.TestingSnapshot)
	errs := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case snaps <- &v1.TestingSnapshot{}:
			}
		}
	}()
	return snaps, errs, nil

}

func (m *mockSyncer) Sync(ctx context.Context, snap *v1.TestingSnapshot) error {
	go func() {
		<-ctx.Done()
		m.cancels++
	}()
	m.syncs++
	// set a limit of 50
	if m.syncs >= 50 {
		m.shouldSync = false
	}
	return nil
}

func (m *mockSyncer) ShouldSync(old, new *v1.TestingSnapshot) bool {
	return m.shouldSync
}
