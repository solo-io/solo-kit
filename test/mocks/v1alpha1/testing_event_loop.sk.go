// Code generated by solo-kit. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/solo-kit/pkg/errors"
	skstats "github.com/solo-io/solo-kit/pkg/stats"
)

var (
	mTestingSnapshotTimeSec     = stats.Float64("testing.solo.io/sync/time_sec", "The time taken for a given sync", "1")
	mTestingSnapshotTimeSecView = &view.View{
		Name:        "testing.solo.io/sync/time_sec",
		Description: "The time taken for a given sync",
		TagKeys:     []tag.Key{tag.MustNewKey("syncer_name")},
		Measure:     mTestingSnapshotTimeSec,
		Aggregation: view.Distribution(0.01, 0.05, 0.1, 0.25, 0.5, 1, 5, 10, 60),
	}
)

func init() {
	view.Register(
		mTestingSnapshotTimeSecView,
	)
}

type TestingSyncer interface {
	Sync(context.Context, *TestingSnapshot) error
}

type TestingSyncers []TestingSyncer

func (s TestingSyncers) Sync(ctx context.Context, snapshot *TestingSnapshot) error {
	var multiErr *multierror.Error
	for _, syncer := range s {
		if err := syncer.Sync(ctx, snapshot); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr.ErrorOrNil()
}

type testingEventLoop struct {
	emitter TestingSnapshotEmitter
	syncer  TestingSyncer
	ready   chan struct{}
}

func NewTestingEventLoop(emitter TestingSnapshotEmitter, syncer TestingSyncer) eventloop.EventLoop {
	return &testingEventLoop{
		emitter: emitter,
		syncer:  syncer,
		ready:   make(chan struct{}),
	}
}

func (el *testingEventLoop) Ready() <-chan struct{} {
	return el.ready
}

func (el *testingEventLoop) Run(namespaces []string, opts clients.WatchOpts) (<-chan error, error) {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "v1alpha1.event_loop")
	logger := contextutils.LoggerFrom(opts.Ctx)
	logger.Infof("event loop started")

	errs := make(chan error)

	watch, emitterErrs, err := el.emitter.Snapshots(namespaces, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "starting snapshot watch")
	}
	go errutils.AggregateErrs(opts.Ctx, errs, emitterErrs, "v1alpha1.emitter errors")
	go func() {
		var channelClosed bool
		// create a new context for each loop, cancel it before each loop
		var cancel context.CancelFunc = func() {}
		// use closure to allow cancel function to be updated as context changes
		defer func() { cancel() }()
		for {
			select {
			case snapshot, ok := <-watch:
				if !ok {
					return
				}
				// cancel any open watches from previous loop
				cancel()

				startTime := time.Now()
				ctx, span := trace.StartSpan(opts.Ctx, "testing.solo.io.EventLoopSync")
				ctx, canc := context.WithCancel(ctx)
				cancel = canc
				err := el.syncer.Sync(ctx, snapshot)
				stats.RecordWithTags(
					ctx,
					[]tag.Mutator{
						tag.Insert(skstats.SyncerNameKey, fmt.Sprintf("%T", el.syncer)),
					},
					mTestingSnapshotTimeSec.M(time.Now().Sub(startTime).Seconds()),
				)
				span.End()

				if err != nil {
					select {
					case errs <- err:
					default:
						logger.Errorf("write error channel is full! could not propagate err: %v", err)
					}
				} else if !channelClosed {
					channelClosed = true
					close(el.ready)
				}
			case <-opts.Ctx.Done():
				return
			}
		}
	}()
	return errs, nil
}
