// Code generated by solo-kit. DO NOT EDIT.

package v1alpha1

import (
	"context"
	fmt "fmt"
	"time"

	"go.opencensus.io/stats"

	"github.com/solo-io/go-utils/errutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
)

type TestingSimpleEmitter interface {
	Snapshots(ctx context.Context) (<-chan *TestingSnapshot, <-chan error, error)
}

func NewTestingSimpleEmitter(aggregatedWatch clients.ResourceWatch) TestingSimpleEmitter {
	return NewTestingSimpleEmitterWithEmit(aggregatedWatch, make(chan struct{}))
}

func NewTestingSimpleEmitterWithEmit(aggregatedWatch clients.ResourceWatch, emit <-chan struct{}) TestingSimpleEmitter {
	return &testingSimpleEmitter{
		aggregatedWatch: aggregatedWatch,
		forceEmit:       emit,
	}
}

type testingSimpleEmitter struct {
	forceEmit       <-chan struct{}
	aggregatedWatch clients.ResourceWatch
}

func (c *testingSimpleEmitter) Snapshots(ctx context.Context) (<-chan *TestingSnapshot, <-chan error, error) {
	snapshots := make(chan *TestingSnapshot)
	errs := make(chan error)

	untyped, watchErrs, err := c.aggregatedWatch(ctx)
	if err != nil {
		return nil, nil, err
	}

	go errutils.AggregateErrs(ctx, errs, watchErrs, "testing-emitter")

	go func() {
		originalSnapshot := TestingSnapshot{}
		currentSnapshot := originalSnapshot.Clone()
		timer := time.NewTicker(time.Second * 1)
		var originalHash uint64
		sync := func() {
			currentHash := currentSnapshot.Hash()
			if originalHash == currentHash {
				return
			}

			originalHash = currentHash

			stats.Record(ctx, mTestingSnapshotOut.M(1))
			originalSnapshot = currentSnapshot.Clone()
			sentSnapshot := currentSnapshot.Clone()
			snapshots <- &sentSnapshot
		}

		defer func() {
			close(snapshots)
			close(errs)
		}()

		for {
			record := func() { stats.Record(ctx, mTestingSnapshotIn.M(1)) }

			select {
			case <-timer.C:
				sync()
			case <-ctx.Done():
				return
			case <-c.forceEmit:
				sentSnapshot := currentSnapshot.Clone()
				snapshots <- &sentSnapshot
			case untypedList := <-untyped:
				record()

				currentSnapshot = TestingSnapshot{}
				for _, res := range untypedList {
					switch typed := res.(type) {
					case *MockResource:
						currentSnapshot.Mocks = append(currentSnapshot.Mocks, typed)
					default:
						select {
						case errs <- fmt.Errorf("TestingSnapshotEmitter "+
							"cannot process resource %v of type %T", res.GetMetadata().Ref(), res):
						case <-ctx.Done():
							return
						}
					}
				}

			}
		}
	}()
	return snapshots, errs, nil
}
