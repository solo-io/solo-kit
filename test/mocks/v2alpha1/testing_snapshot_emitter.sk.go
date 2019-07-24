// Code generated by solo-kit. DO NOT EDIT.

package v2alpha1

import (
	"sync"
	"time"

	testing_solo_io "github.com/solo-io/solo-kit/test/mocks/v1"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/solo-io/go-utils/errutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/errors"
)

var (
	mTestingSnapshotIn  = stats.Int64("testing.solo.io/snap_emitter/snap_in", "The number of snapshots in", "1")
	mTestingSnapshotOut = stats.Int64("testing.solo.io/snap_emitter/snap_out", "The number of snapshots out", "1")

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
)

func init() {
	view.Register(testingsnapshotInView, testingsnapshotOutView)
}

type TestingEmitter interface {
	Register() error
	MockResource() MockResourceClient
	FakeResource() testing_solo_io.FakeResourceClient
	Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *TestingSnapshot, <-chan error, error)
}

func NewTestingEmitter(mockResourceClient MockResourceClient, fakeResourceClient testing_solo_io.FakeResourceClient) TestingEmitter {
	return NewTestingEmitterWithEmit(mockResourceClient, fakeResourceClient, make(chan struct{}))
}

func NewTestingEmitterWithEmit(mockResourceClient MockResourceClient, fakeResourceClient testing_solo_io.FakeResourceClient, emit <-chan struct{}) TestingEmitter {
	return &testingEmitter{
		mockResource: mockResourceClient,
		fakeResource: fakeResourceClient,
		forceEmit:    emit,
	}
}

type testingEmitter struct {
	forceEmit    <-chan struct{}
	mockResource MockResourceClient
	fakeResource testing_solo_io.FakeResourceClient
}

func (c *testingEmitter) Register() error {
	if err := c.mockResource.Register(); err != nil {
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
	/* Create channel for FakeResource */
	type fakeResourceListWithNamespace struct {
		list      testing_solo_io.FakeResourceList
		namespace string
	}
	fakeResourceChan := make(chan fakeResourceListWithNamespace)

	var initialFakeResourceList testing_solo_io.FakeResourceList

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
		/* Setup namespaced watch for FakeResource */
		{
			fakes, err := c.fakeResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial FakeResource list")
			}
			initialFakeResourceList = append(initialFakeResourceList, fakes...)
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
				case mockResourceList := <-mockResourceNamespacesChan:
					select {
					case <-ctx.Done():
						return
					case mockResourceChan <- mockResourceListWithNamespace{list: mockResourceList, namespace: namespace}:
					}
				case fakeResourceList := <-fakeResourceNamespacesChan:
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
	/* Initialize snapshot for Fakes */
	currentSnapshot.Fakes = initialFakeResourceList.Sort()

	snapshots := make(chan *TestingSnapshot)
	go func() {
		originalSnapshot := TestingSnapshot{}
		timer := time.NewTicker(time.Second * 1)

		sync := func() {
			if originalSnapshot.Hash() == currentSnapshot.Hash() {
				return
			}

			stats.Record(ctx, mTestingSnapshotOut.M(1))
			originalSnapshot = currentSnapshot.Clone()
			sentSnapshot := currentSnapshot.Clone()
			snapshots <- &sentSnapshot
		}
		mocksByNamespace := make(map[string]MockResourceList)
		fakesByNamespace := make(map[string]testing_solo_io.FakeResourceList)

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
			case fakeResourceNamespacedList := <-fakeResourceChan:
				record()

				namespace := fakeResourceNamespacedList.namespace

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
