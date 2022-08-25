// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"bytes"
	"sync"
	"time"

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
	// Deprecated. See mKubeconfigsResourcesIn
	mKubeconfigsSnapshotIn = stats.Int64("kubeconfigs.multicluster.solo.io/emitter/snap_in", "Deprecated. Use kubeconfigs.multicluster.solo.io/emitter/resources_in. The number of snapshots in", "1")

	// metrics for emitter
	mKubeconfigsResourcesIn    = stats.Int64("kubeconfigs.multicluster.solo.io/emitter/resources_in", "The number of resource lists received on open watch channels", "1")
	mKubeconfigsSnapshotOut    = stats.Int64("kubeconfigs.multicluster.solo.io/emitter/snap_out", "The number of snapshots out", "1")
	mKubeconfigsSnapshotMissed = stats.Int64("kubeconfigs.multicluster.solo.io/emitter/snap_missed", "The number of snapshots missed", "1")

	// views for emitter
	// deprecated: see kubeconfigsResourcesInView
	kubeconfigssnapshotInView = &view.View{
		Name:        "kubeconfigs.multicluster.solo.io/emitter/snap_in",
		Measure:     mKubeconfigsSnapshotIn,
		Description: "Deprecated. Use kubeconfigs.multicluster.solo.io/emitter/resources_in. The number of snapshots updates coming in.",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}

	kubeconfigsResourcesInView = &view.View{
		Name:        "kubeconfigs.multicluster.solo.io/emitter/resources_in",
		Measure:     mKubeconfigsResourcesIn,
		Description: "The number of resource lists received on open watch channels",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			skstats.NamespaceKey,
			skstats.ResourceKey,
		},
	}
	kubeconfigssnapshotOutView = &view.View{
		Name:        "kubeconfigs.multicluster.solo.io/emitter/snap_out",
		Measure:     mKubeconfigsSnapshotOut,
		Description: "The number of snapshots updates going out",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
	kubeconfigssnapshotMissedView = &view.View{
		Name:        "kubeconfigs.multicluster.solo.io/emitter/snap_missed",
		Measure:     mKubeconfigsSnapshotMissed,
		Description: "The number of snapshots updates going missed. this can happen in heavy load. missed snapshot will be re-tried after a second.",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
)

func init() {
	view.Register(
		kubeconfigssnapshotInView,
		kubeconfigssnapshotOutView,
		kubeconfigssnapshotMissedView,
		kubeconfigsResourcesInView,
	)
}

type KubeconfigsSnapshotEmitter interface {
	Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *KubeconfigsSnapshot, <-chan error, error)
}

type KubeconfigsEmitter interface {
	KubeconfigsSnapshotEmitter
	Register() error
	KubeConfig() KubeConfigClient
}

func NewKubeconfigsEmitter(kubeConfigClient KubeConfigClient, resourceNamespaceLister resources.ResourceNamespaceLister) KubeconfigsEmitter {
	return NewKubeconfigsEmitterWithEmit(kubeConfigClient, resourceNamespaceLister, make(chan struct{}))
}

func NewKubeconfigsEmitterWithEmit(kubeConfigClient KubeConfigClient, resourceNamespaceLister resources.ResourceNamespaceLister, emit <-chan struct{}) KubeconfigsEmitter {
	return &kubeconfigsEmitter{
		kubeConfig:              kubeConfigClient,
		resourceNamespaceLister: resourceNamespaceLister,
		forceEmit:               emit,
	}
}

type kubeconfigsEmitter struct {
	forceEmit               <-chan struct{}
	kubeConfig              KubeConfigClient
	resourceNamespaceLister resources.ResourceNamespaceLister
}

func (c *kubeconfigsEmitter) Register() error {
	if err := c.kubeConfig.Register(); err != nil {
		return err
	}
	return nil
}

func (c *kubeconfigsEmitter) KubeConfig() KubeConfigClient {
	return c.kubeConfig
}

// TODO-JAKE may want to add some comments around how the snapshot_emitter
// event_loop and resource clients -> resource client implementations work in a README.md
// this would be helpful for documentation purposes

func (c *kubeconfigsEmitter) Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *KubeconfigsSnapshot, <-chan error, error) {

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
	/* Create channel for KubeConfig */
	type kubeConfigListWithNamespace struct {
		list      KubeConfigList
		namespace string
	}
	kubeConfigChan := make(chan kubeConfigListWithNamespace)
	var initialKubeConfigList KubeConfigList

	currentSnapshot := KubeconfigsSnapshot{}
	kubeconfigsByNamespace := sync.Map{}

	if !watchNamespacesIsEmpty {
		// then watch all resources on watch Namespaces

		// watched namespaces
		for _, namespace := range watchNamespaces {
			/* Setup namespaced watch for KubeConfig */
			{
				kubeconfigs, err := c.kubeConfig.List(namespace, watchedNamespacesListOptions)
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial KubeConfig list")
				}
				initialKubeConfigList = append(initialKubeConfigList, kubeconfigs...)
				kubeconfigsByNamespace.Store(namespace, kubeconfigs)
			}
			kubeConfigNamespacesChan, kubeConfigErrs, err := c.kubeConfig.Watch(namespace, watchedNamespacesWatchOptions)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "starting KubeConfig watch")
			}

			done.Add(1)
			go func(namespace string) {
				defer done.Done()
				errutils.AggregateErrs(ctx, errs, kubeConfigErrs, namespace+"-kubeconfigs")
			}(namespace)
			/* Watch for changes and update snapshot */
			go func(namespace string) {
				for {
					select {
					case <-ctx.Done():
						return
					case kubeConfigList, ok := <-kubeConfigNamespacesChan:
						if !ok {
							return
						}
						select {
						case <-ctx.Done():
							return
						case kubeConfigChan <- kubeConfigListWithNamespace{list: kubeConfigList, namespace: namespace}:
						}
					}
				}
			}(namespace)
		}
	}
	// watch all other namespaces that fit the Expression Selectors
	if opts.ExpressionSelector != "" {
		// watch resources of non-watched namespaces that fit the Expression
		// Selector filters.

		// first get the renaiming namespaces
		excludeNamespacesFieldDesciptors := ""

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

		// TODO-JAKE need to add in the other namespaces that will not be allowed, IE the exclusion list.
		// this could be built dyynamically

		// TODO-JAKE test that we can create a huge field selector of massive size
		namespacesResources, err := c.resourceNamespaceLister.GetNamespaceResourceList(ctx, resources.ResourceNamespaceListOptions{
			// TODO-JAKE field selectors are not working
			FieldSelectors:      excludeNamespacesFieldDesciptors,
			ExpressionSelectors: opts.ExpressionSelector,
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
					break
				}
			}
			if add {
				allOtherNamespaces = append(allOtherNamespaces, ns.Name)
			}
		}

		// non Watched Namespaces
		// REFACTOR
		for _, namespace := range allOtherNamespaces {
			/* Setup namespaced watch for KubeConfig */
			{
				clien
				kubeconfigs, err := c.kubeConfig.List(namespace, clients.ListOpts{Ctx: opts.Ctx})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial KubeConfig list")
				}
				initialKubeConfigList = append(initialKubeConfigList, kubeconfigs...)
				kubeconfigsByNamespace.Store(namespace, kubeconfigs)
			}
			kubeConfigNamespacesChan, kubeConfigErrs, err := c.kubeConfig.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx})
			if err != nil {
				return nil, nil, errors.Wrapf(err, "starting KubeConfig watch")
			}

			done.Add(1)
			go func(namespace string) {
				defer done.Done()
				errutils.AggregateErrs(ctx, errs, kubeConfigErrs, namespace+"-kubeconfigs")
			}(namespace)
			/* Watch for changes and update snapshot */
			go func(namespace string) {
				for {
					select {
					case <-ctx.Done():
						return
					case kubeConfigList, ok := <-kubeConfigNamespacesChan:
						if !ok {
							return
						}
						select {
						case <-ctx.Done():
							return
						case kubeConfigChan <- kubeConfigListWithNamespace{list: kubeConfigList, namespace: namespace}:
						}
					}
				}
			}(namespace)
		}
		// create watch on all namespaces, so that we can add resources from new namespaces
		// TODO-JAKE this interface has to deal with the event types of kubernetes independently without the interface knowing about it.
		// we will need a way to deal with DELETES and CREATES and updates seperately
		namespaceWatch, _, err := c.resourceNamespaceLister.GetNamespaceResourceWatch(ctx, resources.ResourceNamespaceWatchOptions{
			FieldSelectors:      excludeNamespacesFieldDesciptors,
			ExpressionSelectors: opts.ExpressionSelector,
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
					// TODO-JAKE with the interface, we have lost the ability to know the event type.
					// so the interface must be able to identify the type of event that occured as well
					// not just return the list of namespaces
					newNamespaces := []string{}

					for _, ns := range resourceNamespaces {
						// TODO-JAKE are we sure we need this. Looks like there is a cocurrent map read and map write here
						// TODO-JAKE we willl only need to do this once, I might be best to keep a set/map of the current
						// namespaces that are used
						if _, hit := kubeconfigsByNamespace.Load(ns.Name); !hit {
							newNamespaces = append(newNamespaces, ns.Name)
							continue
						}
					}
					// add a watch for all the new namespaces
					for _, namespace := range newNamespaces {
						/* Setup namespaced watch for KubeConfig for new namespace */
						{
							kubeconfigs, err := c.kubeConfig.List(namespace, clients.ListOpts{Ctx: opts.Ctx})
							if err != nil {
								// INFO-JAKE not sure if we want to do something else
								// but since this is occuring in async I think it should be fine
								errs <- errors.Wrapf(err, "initial new namespace KubeConfig list")
								continue
							}
							kubeconfigsByNamespace.Store(namespace, kubeconfigs)
						}
						kubeConfigNamespacesChan, kubeConfigErrs, err := c.kubeConfig.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx})
						if err != nil {
							// TODO-JAKE if we do decide to have the namespaceErrs from the watch namespaces functionality
							// , then we could add it here namespaceErrs <- error(*) . the namespaceErrs is coming from the
							// ResourceNamespaceLister currently
							// INFO-JAKE is this what we really want to do when there is an error?
							errs <- errors.Wrapf(err, "starting new namespace KubeConfig watch")
							continue
						}

						// INFO-JAKE I think this is appropriate, becasue
						// we want to watch the errors coming off the namespace
						done.Add(1)
						go func(namespace string) {
							defer done.Done()
							errutils.AggregateErrs(ctx, errs, kubeConfigErrs, namespace+"-new-namespace-kubeconfigs")
						}(namespace)
						/* Watch for changes and update snapshot */
						// REFACTOR
						go func(namespace string) {
							for {
								select {
								case <-ctx.Done():
									return
								case kubeConfigList, ok := <-kubeConfigNamespacesChan:
									if !ok {
										return
									}
									select {
									case <-ctx.Done():
										return
									case kubeConfigChan <- kubeConfigListWithNamespace{list: kubeConfigList, namespace: namespace}:
									}
								}
							}
						}(namespace)
					}
				}
			}
		}()
	}
	/* Initialize snapshot for Kubeconfigs */
	currentSnapshot.Kubeconfigs = initialKubeConfigList.Sort()

	snapshots := make(chan *KubeconfigsSnapshot)
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
				stats.Record(ctx, mKubeconfigsSnapshotOut.M(1))
				previousHash = currentHash
			default:
				stats.Record(ctx, mKubeconfigsSnapshotMissed.M(1))
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
			record := func() { stats.Record(ctx, mKubeconfigsSnapshotIn.M(1)) }

			select {
			case <-timer.C:
				sync()
			case <-ctx.Done():
				return
			case <-c.forceEmit:
				sentSnapshot := currentSnapshot.Clone()
				snapshots <- &sentSnapshot
			case kubeConfigNamespacedList, ok := <-kubeConfigChan:
				if !ok {
					return
				}
				record()

				namespace := kubeConfigNamespacedList.namespace

				skstats.IncrementResourceCount(
					ctx,
					namespace,
					"kube_config",
					mKubeconfigsResourcesIn,
				)

				// merge lists by namespace
				kubeconfigsByNamespace.Store(namespace, kubeConfigNamespacedList.list)
				var kubeConfigList KubeConfigList
				kubeconfigsByNamespace.Range(func(key interface{}, value interface{}) bool {
					mocks := value.(KubeConfigList)
					kubeConfigList = append(kubeConfigList, mocks...)
					return true
				})
				currentSnapshot.Kubeconfigs = kubeConfigList.Sort()
			}
		}
	}()
	return snapshots, errs, nil
}
