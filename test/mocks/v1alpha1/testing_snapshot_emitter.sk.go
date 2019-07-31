// Code generated by solo-kit. DO NOT EDIT.

package v1alpha1

import (
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/solo-io/go-utils/errutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/errors"
)

var (
	mTestingSnapshotIn     = stats.Int64("testing.solo.io/snap_emitter/snap_in", "The number of snapshots in", "1")
	mTestingSnapshotOut    = stats.Int64("testing.solo.io/snap_emitter/snap_out", "The number of snapshots out", "1")
	mTestingSnapshotMissed = stats.Int64("testing.solo.io/snap_emitter/snap_missed", "The number of snapshots missed", "1")

	testingsnapshotInView = &view.View{
		Name:        "testing.solo.io_snap_emitter/snap_in",
		Measure:     mTestingSnapshotIn,
		Description: "The number of snapshots updates coming in",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
	testingsnapshotOutView = &view.View{
		Name:        "testing.solo.io/snap_emitter/snap_out",
		Measure:     mTestingSnapshotOut,
		Description: "The number of snapshots updates going out",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
	testingsnapshotMissedView = &view.View{
		Name:        "testing.solo.io/snap_emitter/snap_missed",
		Measure:     mTestingSnapshotMissed,
		Description: "The number of snapshots updates going missed. this can happen in heavy load. missed snapshot will be re-tried after a second.",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
)

func init() {
	view.Register(testingsnapshotInView, testingsnapshotOutView, testingsnapshotMissedView)
}

type TestingEmitter interface {
	Register() error
	MockResource() MockResourceClient
	Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *TestingSnapshot, <-chan error, error)
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
				case mockResourceList := <-mockResourceNamespacesChan:
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

		originalSnapshot := TestingSnapshot{}
		timer := time.NewTicker(time.Second * 1)

		sync := func() {
			if originalSnapshot.Hash() == currentSnapshot.Hash() {
				return
			}

			sentSnapshot := currentSnapshot.Clone()
			select {
			case snapshots <- &sentSnapshot:
				stats.Record(ctx, mTestingSnapshotOut.M(1))
				originalSnapshot = currentSnapshot.Clone()
			default:
				stats.Record(ctx, mTestingSnapshotMissed.M(1))
			}
		}
		mocksByNamespace := make(map[string]MockResourceList)

		for {
			record := func() { stats.Record(ctx, mTestingSnapshotIn.M(1)) }

			select {
			case <-timer.C:
				sync()
			case <-ctx.Done():
				close(snapshots)
				done.Wait()
				close(errs)
				return
			case <-c.forceEmit:
				sentSnapshot := currentSnapshot.Clone()
				snapshots <- &sentSnapshot
			case mockResourceNamespacedList := <-mockResourceChan:
				record()

				namespace := mockResourceNamespacedList.namespace

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
