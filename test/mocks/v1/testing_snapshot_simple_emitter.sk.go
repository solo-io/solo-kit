// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"context"
	"fmt"
	"time"

	github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"

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
		needsSync := false
		// intentionally rate-limited so that our sync loops have time to complete before the next snapshot is sent
		timer := time.NewTicker(time.Second * 1)
		defer timer.Stop()
		var previousHash uint64
		sync := func() {
			currentHash, err := currentSnapshot.Hash(nil)
			if err != nil {
				contextutils.LoggerFrom(ctx).Panicw("error while hashing, this should never happen", zap.Error(err))
			}
			if !needsSync && previousHash == currentHash {
				return
			}
			previousHash = currentHash
			stats.Record(ctx, mTestingSnapshotOut.M(1))
			sentSnapshot := currentSnapshot.Clone()
			snapshots <- &sentSnapshot
			needsSync = false
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
				needsSync = true
			case untypedList := <-untyped:
				record()
				currentSnapshot = TestingSnapshot{}
				for _, res := range untypedList {
					switch typed := res.(type) {
					case *SimpleMockResource:
						currentSnapshot.Simplemocks = append(currentSnapshot.Simplemocks, typed)
					case *MockResource:
						currentSnapshot.Mocks = append(currentSnapshot.Mocks, typed)
					case *FakeResource:
						currentSnapshot.Fakes = append(currentSnapshot.Fakes, typed)
					case *AnotherMockResource:
						currentSnapshot.Anothermockresources = append(currentSnapshot.Anothermockresources, typed)
					case *ClusterResource:
						currentSnapshot.Clusterresources = append(currentSnapshot.Clusterresources, typed)
					case *MockCustomType:
						currentSnapshot.Mcts = append(currentSnapshot.Mcts, typed)
					case *github_com_solo_io_solo_kit_pkg_api_v1_resources_common_kubernetes.Pod:
						currentSnapshot.Pods = append(currentSnapshot.Pods, typed)
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
