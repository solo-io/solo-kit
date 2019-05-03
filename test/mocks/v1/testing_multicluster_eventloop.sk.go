package v1

//
//import (
//	"context"
//	"sync"
//
//	"go.opencensus.io/trace"
//
//	"github.com/solo-io/go-utils/contextutils"
//	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
//	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
//	"github.com/solo-io/solo-kit/pkg/errors"
//	"github.com/solo-io/solo-kit/pkg/utils/errutils"
//)
//
//type testingMcEventLoop struct {
//	emitters []TestingEmitter
//	syncer   TestingSyncer
//}
//
//func NewTestingMcEventLoop(emitters []TestingEmitter, syncer TestingSyncer) eventloop.EventLoop {
//	return &testingMcEventLoop{
//		emitters: emitters,
//		syncer:   syncer,
//	}
//}
//
//func (el *testingMcEventLoop) Run(namespaces []string, opts clients.WatchOpts) (<-chan error, error) {
//	opts = opts.WithDefaults()
//	opts.Ctx = contextutils.WithLogger(opts.Ctx, "v1.event_loop")
//	logger := contextutils.LoggerFrom(opts.Ctx)
//	logger.Infof("event loop started")
//
//	errs := make(chan error)
//
//	aggregatedSnapshots := make(chan *TestingSnapshot)
//	snapshotsByEmitter := make(map[TestingEmitter]*TestingSnapshot)
//	var access sync.RWMutex
//	resync := func(emitter TestingEmitter, snap *TestingSnapshot) {
//		access.Lock()
//		snapshotsByEmitter[emitter] = snap
//		access.Unlock()
//	}
//	mergeSnapshots := func() {
//		access.RLock()
//		mergedSnap := &TestingSnapshot{}
//		for _, snap := range snapshotsByEmitter {
//			for _, res := range snap.Mocks {
//				mergedSnap.Mocks = append(mergedSnap.Mocks, res)
//			}
//		}
//		defer access.RUnlock()
//	}
//	for _, emitter := range el.emitters {
//		emitter := emitter
//		watch, emitterErrs, err := emitter.Snapshots(namespaces, opts)
//		if err != nil {
//			return nil, errors.Wrapf(err, "starting snapshot watch")
//		}
//		go errutils.AggregateErrs(opts.Ctx, errs, emitterErrs, "v1.emitter errors")
//		go func() {
//			for {
//				select {
//				case snapshot, ok := <-watch:
//					if !ok {
//						return
//					}
//					resync(emitter, snapshot)
//				case <-opts.Ctx.Done():
//					return
//				}
//			}
//		}()
//	}
//	go func() {
//		// create a new context for each loop, cancel it before each loop
//		var cancel context.CancelFunc = func() {}
//		// use closure to allow cancel function to be updated as context changes
//		defer func() { cancel() }()
//		for {
//			select {
//			case snapshot, ok := <-aggregatedSnapshots:
//				if !ok {
//					return
//				}
//				// cancel any open watches from previous loop
//				cancel()
//
//				ctx, span := trace.StartSpan(opts.Ctx, "testing.solo.io.EventLoopSync")
//				ctx, canc := context.WithCancel(ctx)
//				cancel = canc
//				err := el.syncer.Sync(ctx, snapshot)
//				span.End()
//
//				if err != nil {
//					select {
//					case errs <- err:
//					default:
//						logger.Errorf("write error channel is full! could not propagate err: %v", err)
//					}
//				}
//			case <-opts.Ctx.Done():
//				return
//			}
//		}
//	}()
//	return errs, nil
//}
