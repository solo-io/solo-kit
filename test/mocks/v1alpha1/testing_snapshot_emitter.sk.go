// Code generated by solo-kit. DO NOT EDIT.

//Generated by pkg/code-generator/codegen/templates/snapshot_emitter_template.go
package v1alpha1

import (
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/errors"
	skstats "github.com/solo-io/solo-kit/pkg/stats"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errutils"
)

var (
	// Deprecated. See mTestingResourcesIn
	mTestingSnapshotIn = stats.Int64("testing.solo.io/emitter/snap_in", "Deprecated. Use testing.solo.io/emitter/resources_in. The number of snapshots in", "1")

	// metrics for emitter
	mTestingResourcesIn    = stats.Int64("testing.solo.io/emitter/resources_in", "The number of resource lists received on open watch channels", "1")
	mTestingSnapshotOut    = stats.Int64("testing.solo.io/emitter/snap_out", "The number of snapshots out", "1")
	mTestingSnapshotMissed = stats.Int64("testing.solo.io/emitter/snap_missed", "The number of snapshots missed", "1")

	// views for emitter
	// deprecated: see testingResourcesInView
	testingsnapshotInView = &view.View{
		Name:        "testing.solo.io/emitter/snap_in",
		Measure:     mTestingSnapshotIn,
		Description: "Deprecated. Use testing.solo.io/emitter/resources_in. The number of snapshots updates coming in.",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}

	testingResourcesInView = &view.View{
		Name:        "testing.solo.io/emitter/resources_in",
		Measure:     mTestingResourcesIn,
		Description: "The number of resource lists received on open watch channels",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			skstats.NamespaceKey,
			skstats.ResourceKey,
		},
	}
	testingsnapshotOutView = &view.View{
		Name:        "testing.solo.io/emitter/snap_out",
		Measure:     mTestingSnapshotOut,
		Description: "The number of snapshots updates going out",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
	testingsnapshotMissedView = &view.View{
		Name:        "testing.solo.io/emitter/snap_missed",
		Measure:     mTestingSnapshotMissed,
		Description: "The number of snapshots updates going missed. this can happen in heavy load. missed snapshot will be re-tried after a second.",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
)

func init() {
	view.Register(
		testingsnapshotInView,
		testingsnapshotOutView,
		testingsnapshotMissedView,
		testingResourcesInView,
	)
}

type TestingSnapshotEmitter interface {
	Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *TestingSnapshot, <-chan error, error)
}

type TestingEmitter interface {
	TestingSnapshotEmitter
	Register() error
	MockResource() MockResourceClient
}

func NewTestingEmitter(mockResourceClient MockResourceClient) TestingEmitter {
	return NewTestingEmitterWithEmit(mockResourceClient, make(chan struct{}))
}

func NewTestingEmitterWithEmit(mockResourceClient MockResourceClient, emit <-chan struct{}) TestingEmitter {
	return &testingEmitter{
		mockResource: mockResourceClient,
		forceEmit:    emit,
	}
}

type testingEmitter struct {
	forceEmit    <-chan struct{}
	mockResource MockResourceClient
}

func (c *testingEmitter) Register() error {
	if err := c.mockResource.Register(); err != nil {
		return err
	}
	return nil
}

func (c *testingEmitter) MockResource() MockResourceClient {
	return c.mockResource
}

func (c *testingEmitter) Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *TestingSnapshot, <-chan error, error) {

	if len(watchNamespaces) == 0 {
		watchNamespaces = []string{""}
	}

	for _, ns := range watchNamespaces {
		if ns == "" && len(watchNamespaces) > 1 {
			return nil, nil, errors.Errorf("the \"\" namespace is used to watch all namespaces. Snapshots can either be tracked for " +
				"specific namespaces or \"\" AllNamespaces, but not both.")
		}
	}

	errs := make(chan error)
	var done sync.WaitGroup
	ctx := opts.Ctx
	/* Create channel for MockResource */
	type mockResourceListWithNamespace struct {
		list      MockResourceList
		namespace string
	}
	mockResourceChan := make(chan mockResourceListWithNamespace)

	var initialMockResourceList MockResourceList

	currentSnapshot := TestingSnapshot{}

	for _, namespace := range watchNamespaces {
		/* Setup namespaced watch for MockResource */
		{
			mocks, err := c.mockResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial MockResource list")
			}
			initialMockResourceList = append(initialMockResourceList, mocks...)
		}
		mockResourceNamespacesChan, mockResourceErrs, err := c.mockResource.Watch(namespace, opts)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "starting MockResource watch")
		}

		done.Add(1)
		go func(namespace string) {
			defer done.Done()
			errutils.AggregateErrs(ctx, errs, mockResourceErrs, namespace+"-mocks")
		}(namespace)

		/* Watch for changes and update snapshot */
		go func(namespace string) {
			for {
				select {
				case <-ctx.Done():
					return
				case mockResourceList, ok := <-mockResourceNamespacesChan:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case mockResourceChan <- mockResourceListWithNamespace{list: mockResourceList, namespace: namespace}:
					}
				}
			}
		}(namespace)
	}
	/* Initialize snapshot for Mocks */
	currentSnapshot.Mocks = initialMockResourceList.Sort()

	snapshots := make(chan *TestingSnapshot)
	go func() {
		// sent initial snapshot to kick off the watch
		initialSnapshot := currentSnapshot.Clone()
		snapshots <- &initialSnapshot

		timer := time.NewTicker(time.Second * 1)
		previousHash, err := currentSnapshot.Hash(nil)
		if err != nil {
			contextutils.LoggerFrom(ctx).Panicw("error while hashing, this should never happen", zap.Error(err))
		}
		sync := func() {
			currentHash, err := currentSnapshot.Hash(nil)
			// this should never happen, so panic if it does
			if err != nil {
				contextutils.LoggerFrom(ctx).Panicw("error while hashing, this should never happen", zap.Error(err))
			}
			if previousHash == currentHash {
				return
			}

			sentSnapshot := currentSnapshot.Clone()
			select {
			case snapshots <- &sentSnapshot:
				stats.Record(ctx, mTestingSnapshotOut.M(1))
				previousHash = currentHash
			default:
				stats.Record(ctx, mTestingSnapshotMissed.M(1))
			}
		}
		mocksByNamespace := make(map[string]MockResourceList)
		defer func() {
			close(snapshots)
			// we must wait for done before closing the error chan,
			// to avoid sending on close channel.
			done.Wait()
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
			case mockResourceNamespacedList, ok := <-mockResourceChan:
				if !ok {
					return
				}
				record()

				namespace := mockResourceNamespacedList.namespace

				skstats.IncrementResourceCount(
					ctx,
					namespace,
					"mock_resource",
					mTestingResourcesIn,
				)

				// merge lists by namespace
				mocksByNamespace[namespace] = mockResourceNamespacedList.list
				var mockResourceList MockResourceList
				for _, mocks := range mocksByNamespace {
					mockResourceList = append(mockResourceList, mocks...)
				}
				currentSnapshot.Mocks = mockResourceList.Sort()
			}
		}
	}()
	return snapshots, errs, nil
}
