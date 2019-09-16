package tests_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

var _ = Describe("Eventloop", func() {

	Context("ready flag works", func() {
		It("should signal ready after first sync", func() {
			emitter := &singleSnapEmitter{}
			syncer := &waitingSyncer{c: make(chan error)}
			eventLoop := v1.NewTestingEventLoop(emitter, syncer)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			eventLoop.Run(nil, clients.WatchOpts{Ctx: ctx})
			Consistently(eventLoop.Ready()).ShouldNot(Receive())
			syncer.c <- fmt.Errorf("error")
			Consistently(eventLoop.Ready()).ShouldNot(Receive())
			syncer.c <- nil
			Eventually(eventLoop.Ready()).Should(BeClosed())
		})
	})
})

type waitingSyncer struct {
	c chan error
}

func (e *waitingSyncer) Sync(context.Context, *v1.TestingSnapshot) error {
	return <-e.c
}

type singleSnapEmitter struct {
}

func (e *singleSnapEmitter) Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *v1.TestingSnapshot, <-chan error, error) {
	snaps := make(chan *v1.TestingSnapshot)
	errs := make(chan error)
	go func() {
		// this test needs two snapshots
		for i := 0; i < 2; i++ {
			select {
			case <-opts.Ctx.Done():
			case snaps <- &v1.TestingSnapshot{}:
			}
		}
	}()
	return snaps, errs, nil

}
