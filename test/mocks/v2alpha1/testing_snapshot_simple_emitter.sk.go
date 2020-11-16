// Code generated by solo-kit. DO NOT EDIT.

package v2alpha1

import (
	"context"
	"fmt"
	"time"

	testing_solo_io "github.com/solo-io/solo-kit/test/mocks/v1"

	"go.opencensus.io/stats"
	"go.uber.org/zap"

	"github.com/solo-io/go-utils/contextutils"
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
		currentSnapshot := TestingSnapshot{}
		timer := time.NewTicker(time.Second * 1)
		var previousHash uint64
		sync := func() {
			currentHash, err := currentSnapshot.Hash(nil)
			if err != nil {
				contextutils.LoggerFrom(ctx).Panicw("error while hashing, this should never happen", zap.Error(err))
			}
			if previousHash == currentHash {
				return
			}

			previousHash = currentHash

			stats.Record(ctx, mTestingSnapshotOut.M(1))
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
					case *FrequentlyChangingAnnotationsResource:
						currentSnapshot.Fcars = append(currentSnapshot.Fcars, typed)
					case *testing_solo_io.FakeResource:
						currentSnapshot.Fakes = append(currentSnapshot.Fakes, typed)
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
