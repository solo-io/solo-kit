// Code generated by solo-kit. DO NOT EDIT.

package v2alpha1

import (
	"sync"
	"time"

	testing_solo_io "github.com/solo-io/solo-kit/test/mocks/v1"

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
	FrequentlyChangingAnnotationsResource() FrequentlyChangingAnnotationsResourceClient
	FakeResource() testing_solo_io.FakeResourceClient
}

func NewTestingEmitter(mockResourceClient MockResourceClient, frequentlyChangingAnnotationsResourceClient FrequentlyChangingAnnotationsResourceClient, fakeResourceClient testing_solo_io.FakeResourceClient) TestingEmitter {
	return NewTestingEmitterWithEmit(mockResourceClient, frequentlyChangingAnnotationsResourceClient, fakeResourceClient, make(chan struct{}))
}

func NewTestingEmitterWithEmit(mockResourceClient MockResourceClient, frequentlyChangingAnnotationsResourceClient FrequentlyChangingAnnotationsResourceClient, fakeResourceClient testing_solo_io.FakeResourceClient, emit <-chan struct{}) TestingEmitter {
	return &testingEmitter{
		mockResource:                          mockResourceClient,
		frequentlyChangingAnnotationsResource: frequentlyChangingAnnotationsResourceClient,
		fakeResource:                          fakeResourceClient,
		forceEmit:                             emit,
	}
}

type testingEmitter struct {
	forceEmit                             <-chan struct{}
	mockResource                          MockResourceClient
	frequentlyChangingAnnotationsResource FrequentlyChangingAnnotationsResourceClient
	fakeResource                          testing_solo_io.FakeResourceClient
}

func (c *testingEmitter) Register() error {
	if err := c.mockResource.Register(); err != nil {
		return err
	}
	if err := c.frequentlyChangingAnnotationsResource.Register(); err != nil {
		return err
	}
	if err := c.fakeResource.Register(); err != nil {
		return err
	}
	return nil
}

func (c *testingEmitter) MockResource() MockResourceClient {
	return c.mockResource
}

func (c *testingEmitter) FrequentlyChangingAnnotationsResource() FrequentlyChangingAnnotationsResourceClient {
	return c.frequentlyChangingAnnotationsResource
}

func (c *testingEmitter) FakeResource() testing_solo_io.FakeResourceClient {
	return c.fakeResource
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
	/* Create channel for FrequentlyChangingAnnotationsResource */
	type frequentlyChangingAnnotationsResourceListWithNamespace struct {
		list      FrequentlyChangingAnnotationsResourceList
		namespace string
	}
	frequentlyChangingAnnotationsResourceChan := make(chan frequentlyChangingAnnotationsResourceListWithNamespace)

	var initialFrequentlyChangingAnnotationsResourceList FrequentlyChangingAnnotationsResourceList
	/* Create channel for FakeResource */
	type fakeResourceListWithNamespace struct {
		list      testing_solo_io.FakeResourceList
		namespace string
	}
	fakeResourceChan := make(chan fakeResourceListWithNamespace)

	var initialFakeResourceList testing_solo_io.FakeResourceList

	currentSnapshot := TestingSnapshot{}
	mocksByNamespace := make(map[string]MockResourceList)
	fcarsByNamespace := make(map[string]FrequentlyChangingAnnotationsResourceList)
	fakesByNamespace := make(map[string]testing_solo_io.FakeResourceList)

	for _, namespace := range watchNamespaces {
		/* Setup namespaced watch for MockResource */
		{
			mocks, err := c.mockResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial MockResource list")
			}
			initialMockResourceList = append(initialMockResourceList, mocks...)
			mocksByNamespace[namespace] = mocks
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
		/* Setup namespaced watch for FrequentlyChangingAnnotationsResource */
		{
			fcars, err := c.frequentlyChangingAnnotationsResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial FrequentlyChangingAnnotationsResource list")
			}
			initialFrequentlyChangingAnnotationsResourceList = append(initialFrequentlyChangingAnnotationsResourceList, fcars...)
			fcarsByNamespace[namespace] = fcars
		}
		frequentlyChangingAnnotationsResourceNamespacesChan, frequentlyChangingAnnotationsResourceErrs, err := c.frequentlyChangingAnnotationsResource.Watch(namespace, opts)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "starting FrequentlyChangingAnnotationsResource watch")
		}

		done.Add(1)
		go func(namespace string) {
			defer done.Done()
			errutils.AggregateErrs(ctx, errs, frequentlyChangingAnnotationsResourceErrs, namespace+"-fcars")
		}(namespace)
		/* Setup namespaced watch for FakeResource */
		{
			fakes, err := c.fakeResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial FakeResource list")
			}
			initialFakeResourceList = append(initialFakeResourceList, fakes...)
			fakesByNamespace[namespace] = fakes
		}
		fakeResourceNamespacesChan, fakeResourceErrs, err := c.fakeResource.Watch(namespace, opts)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "starting FakeResource watch")
		}

		done.Add(1)
		go func(namespace string) {
			defer done.Done()
			errutils.AggregateErrs(ctx, errs, fakeResourceErrs, namespace+"-fakes")
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
				case frequentlyChangingAnnotationsResourceList, ok := <-frequentlyChangingAnnotationsResourceNamespacesChan:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case frequentlyChangingAnnotationsResourceChan <- frequentlyChangingAnnotationsResourceListWithNamespace{list: frequentlyChangingAnnotationsResourceList, namespace: namespace}:
					}
				case fakeResourceList, ok := <-fakeResourceNamespacesChan:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case fakeResourceChan <- fakeResourceListWithNamespace{list: fakeResourceList, namespace: namespace}:
					}
				}
			}
		}(namespace)
	}
	/* Initialize snapshot for Mocks */
	currentSnapshot.Mocks = initialMockResourceList.Sort()
	/* Initialize snapshot for Fcars */
	currentSnapshot.Fcars = initialFrequentlyChangingAnnotationsResourceList.Sort()
	/* Initialize snapshot for Fakes */
	currentSnapshot.Fakes = initialFakeResourceList.Sort()

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
			case frequentlyChangingAnnotationsResourceNamespacedList, ok := <-frequentlyChangingAnnotationsResourceChan:
				if !ok {
					return
				}
				record()

				namespace := frequentlyChangingAnnotationsResourceNamespacedList.namespace

				skstats.IncrementResourceCount(
					ctx,
					namespace,
					"frequently_changing_annotations_resource",
					mTestingResourcesIn,
				)

				// merge lists by namespace
				fcarsByNamespace[namespace] = frequentlyChangingAnnotationsResourceNamespacedList.list
				var frequentlyChangingAnnotationsResourceList FrequentlyChangingAnnotationsResourceList
				for _, fcars := range fcarsByNamespace {
					frequentlyChangingAnnotationsResourceList = append(frequentlyChangingAnnotationsResourceList, fcars...)
				}
				currentSnapshot.Fcars = frequentlyChangingAnnotationsResourceList.Sort()
			case fakeResourceNamespacedList, ok := <-fakeResourceChan:
				if !ok {
					return
				}
				record()

				namespace := fakeResourceNamespacedList.namespace

				skstats.IncrementResourceCount(
					ctx,
					namespace,
					"fake_resource",
					mTestingResourcesIn,
				)

				// merge lists by namespace
				fakesByNamespace[namespace] = fakeResourceNamespacedList.list
				var fakeResourceList testing_solo_io.FakeResourceList
				for _, fakes := range fakesByNamespace {
					fakeResourceList = append(fakeResourceList, fakes...)
				}
				currentSnapshot.Fakes = fakeResourceList.Sort()
			}
		}
	}()
	return snapshots, errs, nil
}
