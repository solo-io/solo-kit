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
	// resourceNamespaceLister is used to watch for new namespaces when they are created.
	// It is used when Expression Selector is in the Watch Opts set in Snapshot().
	resourceNamespaceLister resources.ResourceNamespaceLister
	// namespacesWatching is the set of namespaces that we are watching. This is helpful
	// when Expression Selector is set on the Watch Opts in Snapshot().
	namespacesWatching sync.Map
	// updateNamespaces is used to perform locks and unlocks when watches on namespaces are being updated/created
	updateNamespaces sync.Mutex
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

// Snapshots will return a channel that can be used to receive snapshots of the
// state of the resources it is watching
// when watching resources, you can set the watchNamespaces, and you can set the
// ExpressionSelector of the WatchOpts.  Setting watchNamespaces will watch for all resources
// that are in the specified namespaces. In addition if ExpressionSelector of the WatchOpts is
// set, then all namespaces that meet the label criteria of the ExpressionSelector will
// also be watched.
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
	hasWatchedNamespaces := len(watchNamespaces) > 1 || (len(watchNamespaces) == 1 && watchNamespaces[0] != "")
	watchingLabeledNamespaces := !(opts.ExpressionSelector == "")
	var done sync.WaitGroup
	ctx := opts.Ctx

	// setting up the options for both listing and watching resources in namespaces
	watchedNamespacesListOptions := clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector}
	watchedNamespacesWatchOptions := clients.WatchOpts{Ctx: opts.Ctx, Selector: opts.Selector}
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
	if hasWatchedNamespaces || !watchingLabeledNamespaces {
		// then watch all resources on watch Namespaces

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
				defer func() {
					c.namespacesWatching.Delete(namespace)
				}()
				c.namespacesWatching.Store(namespace, true)
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
	// watch all other namespaces that are labeled and fit the Expression Selector
	if opts.ExpressionSelector != "" {
		// watch resources of non-watched namespaces that fit the expression selectors
		namespaceListOptions := resources.ResourceNamespaceListOptions{
			Ctx:                opts.Ctx,
			ExpressionSelector: opts.ExpressionSelector,
		}
		namespaceWatchOptions := resources.ResourceNamespaceWatchOptions{
			Ctx:                opts.Ctx,
			ExpressionSelector: opts.ExpressionSelector,
		}

		filterNamespaces := resources.ResourceNamespaceList{}
		for _, ns := range watchNamespaces {
			// we do not want to filter out "" which equals all namespaces
			// the reason is because we will never create a watch on ""(all namespaces) because
			// doing so means we watch all resources regardless of namespace. Our intent is to
			// watch only certain namespaces.
			if ns != "" {
				filterNamespaces = append(filterNamespaces, resources.ResourceNamespace{Name: ns})
			}
		}
		namespacesResources, err := c.resourceNamespaceLister.GetResourceNamespaceList(namespaceListOptions, filterNamespaces)
		if err != nil {
			return nil, nil, err
		}
		newlyRegisteredNamespaces := make([]string, len(namespacesResources))
		// non watched namespaces that are labeled
		for i, resourceNamespace := range namespacesResources {
			namespace := resourceNamespace.Name
			newlyRegisteredNamespaces[i] = namespace
			err = c.mockResource.RegisterNamespace(namespace)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "there was an error registering the namespace to the mockResource")
			}
			/* Setup namespaced watch for MockResource */
			{
				mocks, err := c.mockResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial MockResource list with new namespace")
				}
				initialMockResourceList = append(initialMockResourceList, mocks...)
				mocksByNamespace.Store(namespace, mocks)
			}
			mockResourceNamespacesChan, mockResourceErrs, err := c.mockResource.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx})
			if err != nil {
				return nil, nil, errors.Wrapf(err, "starting MockResource watch")
			}

			done.Add(1)
			go func(namespace string) {
				defer done.Done()
				errutils.AggregateErrs(ctx, errs, mockResourceErrs, namespace+"-mocks")
			}(namespace)
			err = c.frequentlyChangingAnnotationsResource.RegisterNamespace(namespace)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "there was an error registering the namespace to the frequentlyChangingAnnotationsResource")
			}
			/* Setup namespaced watch for FrequentlyChangingAnnotationsResource */
			{
				fcars, err := c.frequentlyChangingAnnotationsResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial FrequentlyChangingAnnotationsResource list with new namespace")
				}
				initialFrequentlyChangingAnnotationsResourceList = append(initialFrequentlyChangingAnnotationsResourceList, fcars...)
				fcarsByNamespace.Store(namespace, fcars)
			}
			frequentlyChangingAnnotationsResourceNamespacesChan, frequentlyChangingAnnotationsResourceErrs, err := c.frequentlyChangingAnnotationsResource.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx})
			if err != nil {
				return nil, nil, errors.Wrapf(err, "starting FrequentlyChangingAnnotationsResource watch")
			}

			done.Add(1)
			go func(namespace string) {
				defer done.Done()
				errutils.AggregateErrs(ctx, errs, frequentlyChangingAnnotationsResourceErrs, namespace+"-fcars")
			}(namespace)
			err = c.fakeResource.RegisterNamespace(namespace)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "there was an error registering the namespace to the fakeResource")
			}
			/* Setup namespaced watch for FakeResource */
			{
				fakes, err := c.fakeResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial FakeResource list with new namespace")
				}
				initialFakeResourceList = append(initialFakeResourceList, fakes...)
				fakesByNamespace.Store(namespace, fakes)
			}
			fakeResourceNamespacesChan, fakeResourceErrs, err := c.fakeResource.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx})
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
		if len(newlyRegisteredNamespaces) > 0 {
			contextutils.LoggerFrom(ctx).Infof("registered the new namespace %v", newlyRegisteredNamespaces)
		}

		// create watch on all namespaces, so that we can add all resources from new namespaces
		// we will be watching namespaces that meet the Expression Selector filter

		namespaceWatch, errsReceiver, err := c.resourceNamespaceLister.GetResourceNamespaceWatch(namespaceWatchOptions, filterNamespaces)
		if err != nil {
			return nil, nil, err
		}
		if errsReceiver != nil {
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case err = <-errsReceiver:
						errs <- errors.Wrapf(err, "received error from watch on resource namespaces")
					}
				}
			}()
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
					// get the list of new namespaces, if there is a new namespace
					// get the list of resources from that namespace, and add
					// a watch for new resources created/deleted on that namespace
					c.updateNamespaces.Lock()

					// get the new namespaces, and get a map of the namespaces
					mapOfResourceNamespaces := make(map[string]bool, len(resourceNamespaces))
					newNamespaces := []string{}
					for _, ns := range resourceNamespaces {
						if _, hit := c.namespacesWatching.Load(ns.Name); !hit {
							newNamespaces = append(newNamespaces, ns.Name)
						}
						mapOfResourceNamespaces[ns.Name] = true
					}

					for _, ns := range watchNamespaces {
						mapOfResourceNamespaces[ns] = true
					}

					missingNamespaces := []string{}
					// use the map of namespace resources to find missing/deleted namespaces
					c.namespacesWatching.Range(func(key interface{}, value interface{}) bool {
						name := key.(string)
						if _, hit := mapOfResourceNamespaces[name]; !hit {
							missingNamespaces = append(missingNamespaces, name)
						}
						return true
					})

					for _, ns := range missingNamespaces {
						mockResourceChan <- mockResourceListWithNamespace{list: MockResourceList{}, namespace: ns}
						frequentlyChangingAnnotationsResourceChan <- frequentlyChangingAnnotationsResourceListWithNamespace{list: FrequentlyChangingAnnotationsResourceList{}, namespace: ns}
						fakeResourceChan <- fakeResourceListWithNamespace{list: testing_solo_io.FakeResourceList{}, namespace: ns}
					}

					for _, namespace := range newNamespaces {
						var err error
						err = c.mockResource.RegisterNamespace(namespace)
						if err != nil {
							errs <- errors.Wrapf(err, "there was an error registering the namespace to the mockResource")
							continue
						}
						/* Setup namespaced watch for MockResource for new namespace */
						{
							mocks, err := c.mockResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
							if err != nil {
								errs <- errors.Wrapf(err, "initial new namespace MockResource list in namespace watch")
								continue
							}
							mocksByNamespace.Store(namespace, mocks)
						}
						mockResourceNamespacesChan, mockResourceErrs, err := c.mockResource.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx, Selector: opts.Selector})
						if err != nil {
							errs <- errors.Wrapf(err, "starting new namespace MockResource watch")
							continue
						}

						done.Add(1)
						go func(namespace string) {
							defer done.Done()
							errutils.AggregateErrs(ctx, errs, mockResourceErrs, namespace+"-new-namespace-mocks")
						}(namespace)
						err = c.frequentlyChangingAnnotationsResource.RegisterNamespace(namespace)
						if err != nil {
							errs <- errors.Wrapf(err, "there was an error registering the namespace to the frequentlyChangingAnnotationsResource")
							continue
						}
						/* Setup namespaced watch for FrequentlyChangingAnnotationsResource for new namespace */
						{
							fcars, err := c.frequentlyChangingAnnotationsResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
							if err != nil {
								errs <- errors.Wrapf(err, "initial new namespace FrequentlyChangingAnnotationsResource list in namespace watch")
								continue
							}
							fcarsByNamespace.Store(namespace, fcars)
						}
						frequentlyChangingAnnotationsResourceNamespacesChan, frequentlyChangingAnnotationsResourceErrs, err := c.frequentlyChangingAnnotationsResource.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx, Selector: opts.Selector})
						if err != nil {
							errs <- errors.Wrapf(err, "starting new namespace FrequentlyChangingAnnotationsResource watch")
							continue
						}

						done.Add(1)
						go func(namespace string) {
							defer done.Done()
							errutils.AggregateErrs(ctx, errs, frequentlyChangingAnnotationsResourceErrs, namespace+"-new-namespace-fcars")
						}(namespace)
						err = c.fakeResource.RegisterNamespace(namespace)
						if err != nil {
							errs <- errors.Wrapf(err, "there was an error registering the namespace to the fakeResource")
							continue
						}
						/* Setup namespaced watch for FakeResource for new namespace */
						{
							fakes, err := c.fakeResource.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
							if err != nil {
								errs <- errors.Wrapf(err, "initial new namespace FakeResource list in namespace watch")
								continue
							}
							fakesByNamespace.Store(namespace, fakes)
						}
						fakeResourceNamespacesChan, fakeResourceErrs, err := c.fakeResource.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx, Selector: opts.Selector})
						if err != nil {
							errs <- errors.Wrapf(err, "starting new namespace FakeResource watch")
							continue
						}

						done.Add(1)
						go func(namespace string) {
							defer done.Done()
							errutils.AggregateErrs(ctx, errs, fakeResourceErrs, namespace+"-new-namespace-fakes")
						}(namespace)
						/* Watch for changes and update snapshot */
						go func(namespace string) {
							defer func() {
								c.namespacesWatching.Delete(namespace)
							}()
							c.namespacesWatching.Store(namespace, true)
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
					if len(newNamespaces) > 0 {
						contextutils.LoggerFrom(ctx).Infof("registered the new namespace %v", newNamespaces)
						c.updateNamespaces.Unlock()
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
