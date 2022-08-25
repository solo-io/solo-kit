// Code generated by solo-kit. DO NOT EDIT.

package v2alpha1

import (
	"bytes"
	"sync"
	"time"

	testing_solo_io "github.com/solo-io/solo-kit/test/mocks/v1"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
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

func NewTestingEmitter(mockResourceClient MockResourceClient, frequentlyChangingAnnotationsResourceClient FrequentlyChangingAnnotationsResourceClient, fakeResourceClient testing_solo_io.FakeResourceClient, resourceNamespaceLister resources.ResourceNamespaceLister) TestingEmitter {
	return NewTestingEmitterWithEmit(mockResourceClient, frequentlyChangingAnnotationsResourceClient, fakeResourceClient, resourceNamespaceLister, make(chan struct{}))
}

func NewTestingEmitterWithEmit(mockResourceClient MockResourceClient, frequentlyChangingAnnotationsResourceClient FrequentlyChangingAnnotationsResourceClient, fakeResourceClient testing_solo_io.FakeResourceClient, resourceNamespaceLister resources.ResourceNamespaceLister, emit <-chan struct{}) TestingEmitter {
	return &testingEmitter{
		mockResource:                          mockResourceClient,
		frequentlyChangingAnnotationsResource: frequentlyChangingAnnotationsResourceClient,
		fakeResource:                          fakeResourceClient,
		resourceNamespaceLister:               resourceNamespaceLister,
		forceEmit:                             emit,
	}
}

type testingEmitter struct {
	forceEmit                             <-chan struct{}
	mockResource                          MockResourceClient
	frequentlyChangingAnnotationsResource FrequentlyChangingAnnotationsResourceClient
	fakeResource                          testing_solo_io.FakeResourceClient
	resourceNamespaceLister               resources.ResourceNamespaceLister
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

	// TODO-JAKE some of this should only be present if scoped by namespace
	errs := make(chan error)
	hasWatchedNamespaces := len(watchNamespaces) > 1 || (len(watchNamespaces) == 1 && watchNamespaces[0] != "")
	watchNamespacesIsEmpty := !hasWatchedNamespaces
	var done sync.WaitGroup
	ctx := opts.Ctx

	// if we are watching namespaces, then we do not want to fitler any of the
	// resources in when listing or watching
	// TODO-JAKE not sure if we want to get rid of the Selector in the
	// ListOpts here. the reason that we might want to is because we no
	// longer allow selectors, unless it is on a unwatched namespace.
	watchedNamespacesListOptions := clients.ListOpts{Ctx: opts.Ctx}
	watchedNamespacesWatchOptions := clients.WatchOpts{Ctx: opts.Ctx}
	if watchNamespacesIsEmpty {
		// if the namespaces that we are watching is empty, then we want to apply
		// the expression Selectors to all the namespaces.
		watchedNamespacesListOptions.ExpressionSelector = opts.ExpressionSelector
		watchedNamespacesWatchOptions.ExpressionSelector = opts.ExpressionSelector
	}
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
	mocksByNamespace := sync.Map{}
	fcarsByNamespace := sync.Map{}
	fakesByNamespace := sync.Map{}

	// watched namespaces
	for _, namespace := range watchNamespaces {
		/* Setup namespaced watch for MockResource */
		{
			mocks, err := c.mockResource.List(namespace, watchedNamespacesListOptions)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial MockResource list")
			}
			initialMockResourceList = append(initialMockResourceList, mocks...)
			mocksByNamespace.Store(namespace, mocks)
		}
		mockResourceNamespacesChan, mockResourceErrs, err := c.mockResource.Watch(namespace, watchedNamespacesWatchOptions)
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
			fcars, err := c.frequentlyChangingAnnotationsResource.List(namespace, watchedNamespacesListOptions)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial FrequentlyChangingAnnotationsResource list")
			}
			initialFrequentlyChangingAnnotationsResourceList = append(initialFrequentlyChangingAnnotationsResourceList, fcars...)
			fcarsByNamespace.Store(namespace, fcars)
		}
		frequentlyChangingAnnotationsResourceNamespacesChan, frequentlyChangingAnnotationsResourceErrs, err := c.frequentlyChangingAnnotationsResource.Watch(namespace, watchedNamespacesWatchOptions)
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
			fakes, err := c.fakeResource.List(namespace, watchedNamespacesListOptions)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial FakeResource list")
			}
			initialFakeResourceList = append(initialFakeResourceList, fakes...)
			fakesByNamespace.Store(namespace, fakes)
		}
		fakeResourceNamespacesChan, fakeResourceErrs, err := c.fakeResource.Watch(namespace, watchedNamespacesWatchOptions)
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
	if hasWatchedNamespaces && opts.ExpressionSelector != "" {
		// watch resources using non-watched namespaces. With these namespaces we
		// will watch only those that are filted using the label selectors defined
		// by Expression Selectors

		// first get the renaiming namespaces
		excludeNamespacesFieldDesciptors := ""

		// TODO-JAKE may want to add some comments around how the snapshot_emitter
		// event_loop and resource clients -> resource client implementations work in a README.md
		// this would be helpful for documentation purposes

		// TODO implement how we will be able to delete resources from namespaces that are deleted

		// TODO-JAKE REFACTOR, we can refactor how the watched namespaces are added up to make a exclusion namespaced fields
		var buffer bytes.Buffer
		for i, ns := range watchNamespaces {
			buffer.WriteString("metadata.name!=")
			buffer.WriteString(ns)
			if i < len(watchNamespaces)-1 {
				buffer.WriteByte(',')
			}
		}
		excludeNamespacesFieldDesciptors = buffer.String()

		// we should only be watching namespaces that have the selectors that we want to be watching
		// TODO-JAKE need to add in the other namespaces that will not be allowed, IE the exclusion list
		// TODO-JAKE test that we can create a huge field selector of massive size
		namespacesResources, err := c.resourceNamespaceLister.GetNamespaceResourceList(ctx, resources.ResourceNamespaceListOptions{
			FieldSelectors: excludeNamespacesFieldDesciptors,
		})

		if err != nil {
			return nil, nil, err
		}
		allOtherNamespaces := make([]string, 0)
		for _, ns := range namespacesResources {
			// TODO-JAKE get the filters on the namespacing working
			add := true
			// TODO-JAKE need to implement the filtering of the field selectors in the resourceNamespaceLister
			for _, wns := range watchNamespaces {
				if ns.Name == wns {
					add = false
				}
			}
			if add {
				allOtherNamespaces = append(allOtherNamespaces, ns.Name)
			}
		}

		// nonWatchedNamespaces
		// REFACTOR
		for _, namespace := range allOtherNamespaces {
			/* Setup namespaced watch for MockResource */
			{
				mocks, err := c.mockResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial MockResource list")
				}
				initialMockResourceList = append(initialMockResourceList, mocks...)
				mocksByNamespace.Store(namespace, mocks)
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
				fcars, err := c.frequentlyChangingAnnotationsResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial FrequentlyChangingAnnotationsResource list")
				}
				initialFrequentlyChangingAnnotationsResourceList = append(initialFrequentlyChangingAnnotationsResourceList, fcars...)
				fcarsByNamespace.Store(namespace, fcars)
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
				fakes, err := c.fakeResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial FakeResource list")
				}
				initialFakeResourceList = append(initialFakeResourceList, fakes...)
				fakesByNamespace.Store(namespace, fakes)
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
		// create watch on all namespaces, so that we can add resources from new namespaces
		// TODO-JAKE this interface has to deal with the event types of kubernetes independently without the interface knowing about it.
		// we will need a way to deal with DELETES and CREATES and updates seperately
		namespaceWatch, _, err := c.resourceNamespaceLister.GetNamespaceResourceWatch(ctx, resources.ResourceNamespaceWatchOptions{
			FieldSelectors: excludeNamespacesFieldDesciptors,
		})
		if err != nil {
			return nil, nil, err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case resourceNamespaces, ok := <-namespaceWatch:
					if !ok {
						return
					}
					newNamespaces := []string{}

					for _, ns := range resourceNamespaces {
						// TODO-JAKE are we sure we need this. Looks like there is a cocurrent map read and map write here
						// TODO-JAKE we willl only need to do this once, I might be best to keep a set/map of the current
						// namespaces that are used
						if _, hit := mocksByNamespace.Load(ns.Name); !hit {
							newNamespaces = append(newNamespaces, ns.Name)
							continue
						}
						// TODO-JAKE we willl only need to do this once, I might be best to keep a set/map of the current
						// namespaces that are used
						if _, hit := fcarsByNamespace.Load(ns.Name); !hit {
							newNamespaces = append(newNamespaces, ns.Name)
							continue
						}
						// TODO-JAKE we willl only need to do this once, I might be best to keep a set/map of the current
						// namespaces that are used
						if _, hit := fakesByNamespace.Load(ns.Name); !hit {
							newNamespaces = append(newNamespaces, ns.Name)
							continue
						}
					}
					// TODO-JAKE I think we could get rid of this if statement if needed.
					if len(newNamespaces) > 0 {
						// add a watch for all the new namespaces
						// REFACTOR
						for _, namespace := range newNamespaces {
							/* Setup namespaced watch for MockResource for new namespace */
							{
								mocks, err := c.mockResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
								if err != nil {
									// INFO-JAKE not sure if we want to do something else
									// but since this is occuring in async I think it should be fine
									errs <- errors.Wrapf(err, "initial new namespace MockResource list")
									continue
								}
								mocksByNamespace.Store(namespace, mocks)
							}
							mockResourceNamespacesChan, mockResourceErrs, err := c.mockResource.Watch(namespace, opts)
							if err != nil {
								// TODO-JAKE if we do decide to have the namespaceErrs from the watch namespaces functionality
								// , then we could add it here namespaceErrs <- error(*) . the namespaceErrs is coming from the
								// ResourceNamespaceLister currently
								// INFO-JAKE is this what we really want to do when there is an error?
								errs <- errors.Wrapf(err, "starting new namespace MockResource watch")
								continue
							}

							// INFO-JAKE I think this is appropriate, becasue
							// we want to watch the errors coming off the namespace
							done.Add(1)
							go func(namespace string) {
								defer done.Done()
								errutils.AggregateErrs(ctx, errs, mockResourceErrs, namespace+"-new-namespace-mocks")
							}(namespace)
							/* Setup namespaced watch for FrequentlyChangingAnnotationsResource for new namespace */
							{
								fcars, err := c.frequentlyChangingAnnotationsResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
								if err != nil {
									// INFO-JAKE not sure if we want to do something else
									// but since this is occuring in async I think it should be fine
									errs <- errors.Wrapf(err, "initial new namespace FrequentlyChangingAnnotationsResource list")
									continue
								}
								fcarsByNamespace.Store(namespace, fcars)
							}
							frequentlyChangingAnnotationsResourceNamespacesChan, frequentlyChangingAnnotationsResourceErrs, err := c.frequentlyChangingAnnotationsResource.Watch(namespace, opts)
							if err != nil {
								// TODO-JAKE if we do decide to have the namespaceErrs from the watch namespaces functionality
								// , then we could add it here namespaceErrs <- error(*) . the namespaceErrs is coming from the
								// ResourceNamespaceLister currently
								// INFO-JAKE is this what we really want to do when there is an error?
								errs <- errors.Wrapf(err, "starting new namespace FrequentlyChangingAnnotationsResource watch")
								continue
							}

							// INFO-JAKE I think this is appropriate, becasue
							// we want to watch the errors coming off the namespace
							done.Add(1)
							go func(namespace string) {
								defer done.Done()
								errutils.AggregateErrs(ctx, errs, frequentlyChangingAnnotationsResourceErrs, namespace+"-new-namespace-fcars")
							}(namespace)
							/* Setup namespaced watch for FakeResource for new namespace */
							{
								fakes, err := c.fakeResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
								if err != nil {
									// INFO-JAKE not sure if we want to do something else
									// but since this is occuring in async I think it should be fine
									errs <- errors.Wrapf(err, "initial new namespace FakeResource list")
									continue
								}
								fakesByNamespace.Store(namespace, fakes)
							}
							fakeResourceNamespacesChan, fakeResourceErrs, err := c.fakeResource.Watch(namespace, opts)
							if err != nil {
								// TODO-JAKE if we do decide to have the namespaceErrs from the watch namespaces functionality
								// , then we could add it here namespaceErrs <- error(*) . the namespaceErrs is coming from the
								// ResourceNamespaceLister currently
								// INFO-JAKE is this what we really want to do when there is an error?
								errs <- errors.Wrapf(err, "starting new namespace FakeResource watch")
								continue
							}

							// INFO-JAKE I think this is appropriate, becasue
							// we want to watch the errors coming off the namespace
							done.Add(1)
							go func(namespace string) {
								defer done.Done()
								errutils.AggregateErrs(ctx, errs, fakeResourceErrs, namespace+"-new-namespace-fakes")
							}(namespace)
							/* Watch for changes and update snapshot */
							// REFACTOR
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
					}
				}
			}
		}()
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
				mocksByNamespace.Store(namespace, mockResourceNamespacedList.list)
				var mockResourceList MockResourceList
				mocksByNamespace.Range(func(key interface{}, value interface{}) bool {
					mocks := value.(MockResourceList)
					mockResourceList = append(mockResourceList, mocks...)
					return true
				})
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
				fcarsByNamespace.Store(namespace, frequentlyChangingAnnotationsResourceNamespacedList.list)
				var frequentlyChangingAnnotationsResourceList FrequentlyChangingAnnotationsResourceList
				fcarsByNamespace.Range(func(key interface{}, value interface{}) bool {
					mocks := value.(FrequentlyChangingAnnotationsResourceList)
					frequentlyChangingAnnotationsResourceList = append(frequentlyChangingAnnotationsResourceList, mocks...)
					return true
				})
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
				fakesByNamespace.Store(namespace, fakeResourceNamespacedList.list)
				var fakeResourceList testing_solo_io.FakeResourceList
				fakesByNamespace.Range(func(key interface{}, value interface{}) bool {
					mocks := value.(testing_solo_io.FakeResourceList)
					fakeResourceList = append(fakeResourceList, mocks...)
					return true
				})
				currentSnapshot.Fakes = fakeResourceList.Sort()
			}
		}
	}()
	return snapshots, errs, nil
}
