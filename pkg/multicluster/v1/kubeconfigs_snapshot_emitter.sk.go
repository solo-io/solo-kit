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
	"github.com/solo-io/solo-kit/pkg/errors"
	skstats "github.com/solo-io/solo-kit/pkg/stats"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
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

func NewKubeconfigsEmitter(kubeConfigClient KubeConfigClient) KubeconfigsEmitter {
	return NewKubeconfigsEmitterWithEmit(kubeConfigClient, make(chan struct{}))
}

func NewKubeconfigsEmitterWithEmit(kubeConfigClient KubeConfigClient, emit <-chan struct{}) KubeconfigsEmitter {
	return &kubeconfigsEmitter{
		kubeConfig: kubeConfigClient,
		forceEmit:  emit,
	}
}

type kubeconfigsEmitter struct {
	forceEmit  <-chan struct{}
	kubeConfig KubeConfigClient
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
	kubeconfigsByNamespace := make(map[string]KubeConfigList)

	// watched namespaces
	for _, namespace := range watchNamespaces {
		/* Setup namespaced watch for KubeConfig */
		{
			kubeconfigs, err := c.kubeConfig.List(namespace, watchedNamespacesListOptions)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial KubeConfig list")
			}
			initialKubeConfigList = append(initialKubeConfigList, kubeconfigs...)
			kubeconfigsByNamespace[namespace] = kubeconfigs
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
	if hasWatchedNamespaces && opts.ExpressionSelector != "" {
		// watch resources using non-watched namespaces. With these namespaces we
		// will watch only those that are filted using the label selectors defined
		// by Expression Selectors

		// first get the renaiming namespaces
		var k kubernetes.Interface
		excludeNamespacesFieldDesciptors := ""

		var buffer bytes.Buffer
		for i, ns := range watchNamespaces {
			buffer.WriteString("metadata.namespace!=")
			buffer.WriteString(ns)
			if i < len(watchNamespaces)-1 {
				buffer.WriteByte(',')
			}
		}
		excludeNamespacesFieldDesciptors = buffer.String()

		namespacesResources, err := k.CoreV1().Namespaces().List(ctx, metav1.ListOptions{FieldSelector: excludeNamespacesFieldDesciptors})
		if err != nil {
			return nil, nil, err
		}
		allOtherNamespaces := make([]string, len(namespacesResources.Items))
		for _, ns := range namespacesResources.Items {
			allOtherNamespaces = append(allOtherNamespaces, ns.Namespace)
		}

		// nonWatchedNamespaces
		// REFACTOR
		for _, namespace := range allOtherNamespaces {
			/* Setup namespaced watch for KubeConfig */
			{
				kubeconfigs, err := c.kubeConfig.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial KubeConfig list")
				}
				initialKubeConfigList = append(initialKubeConfigList, kubeconfigs...)
				kubeconfigsByNamespace[namespace] = kubeconfigs
			}
			kubeConfigNamespacesChan, kubeConfigErrs, err := c.kubeConfig.Watch(namespace, opts)
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
		namespaceWatch, err := k.CoreV1().Namespaces().Watch(opts.Ctx, metav1.ListOptions{FieldSelector: excludeNamespacesFieldDesciptors})
		if err != nil {
			return nil, nil, err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-namespaceWatch.ResultChan():
					if !ok {
						return
					}
					switch event.Type {
					case kubewatch.Error:
						errs <- errors.Errorf("receiving namespace event", event)
					default:
						namespacesResources, err := k.CoreV1().Namespaces().List(opts.Ctx, metav1.ListOptions{FieldSelector: excludeNamespacesFieldDesciptors})
						if err != nil {
							errs <- errors.Wrapf(err, "listing the namespace resources")
						}

						hit := false
						newNamespaces := []string{}

						for _, item := range namespacesResources.Items {
							namespace := item.Namespace
							_, hit = kubeconfigsByNamespace[namespace]
							if !hit {
								newNamespaces = append(newNamespaces, namespace)
								continue
							}
						}
						if hit {
							// add a watch for all the new namespaces
							// REFACTOR
							for _, namespace := range newNamespaces {
								/* Setup namespaced watch for KubeConfig for new namespace */
								{
									kubeconfigs, err := c.kubeConfig.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
									if err != nil {
										// INFO-JAKE not sure if we want to do something else
										// but since this is occuring in async I think it should be fine
										errs <- errors.Wrapf(err, "initial new namespace KubeConfig list")
										continue
									}
									kubeconfigsByNamespace[namespace] = kubeconfigs
								}
								kubeConfigNamespacesChan, kubeConfigErrs, err := c.kubeConfig.Watch(namespace, opts)
								if err != nil {
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
				kubeconfigsByNamespace[namespace] = kubeConfigNamespacedList.list
				var kubeConfigList KubeConfigList
				for _, kubeconfigs := range kubeconfigsByNamespace {
					kubeConfigList = append(kubeConfigList, kubeconfigs...)
				}
				currentSnapshot.Kubeconfigs = kubeConfigList.Sort()
			}
		}
	}()
	return snapshots, errs, nil
}
