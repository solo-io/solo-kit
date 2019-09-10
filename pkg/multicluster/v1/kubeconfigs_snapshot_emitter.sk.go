// Code generated by solo-kit. DO NOT EDIT.

package v1

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
	// metrics for sending snapshots
	mKubeconfigsSnapshotIn     = stats.Int64("$emitter_prefix/snap_in", "The number of snapshots in", "1")
	mKubeconfigsSnapshotOut    = stats.Int64("$emitter_prefix/snap_out", "The number of snapshots out", "1")
	mKubeconfigsSnapshotMissed = stats.Int64("$emitter_prefix/snap_missed", "The number of snapshots missed", "1")

	// metrics for resource watches

	mKubeconfigsKubeconfigsListIn = stats.Int64("Kubeconfigs/kube_config_emitter/kubeconfigs_in", "The number of kubeconfigs lists received on watch channel", "1")

	// views for snapshots
	kubeconfigssnapshotInView = &view.View{
		Name:        "kubeconfigs/emitter/snap_in",
		Measure:     mKubeconfigsSnapshotIn,
		Description: "The number of snapshots updates coming in",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
	kubeconfigssnapshotOutView = &view.View{
		Name:        "kubeconfigs/emitter/snap_out",
		Measure:     mKubeconfigsSnapshotOut,
		Description: "The number of snapshots updates going out",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
	kubeconfigssnapshotMissedView = &view.View{
		Name:        "kubeconfigs/emitter/snap_missed",
		Measure:     mKubeconfigsSnapshotMissed,
		Description: "The number of snapshots updates going missed. this can happen in heavy load. missed snapshot will be re-tried after a second.",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}

	kubeconfigsNamespaceKey, _ = tag.NewKey("namespace")

	// views for resource watches
	kubeconfigsKubeconfigsListInView = &view.View{
		Name:        "kubeconfigs/kube_config_emitter/kubeconfigs_in",
		Measure:     mKubeconfigsKubeconfigsListIn,
		Description: "The number of kubeconfigs lists received on watch channel.",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			kubeconfigsNamespaceKey,
		},
	}
)

func init() {
	view.Register(
		kubeconfigssnapshotInView,
		kubeconfigssnapshotOutView,
		kubeconfigssnapshotMissedView,
		kubeconfigsKubeconfigsListInView,
	)
}

type KubeconfigsEmitter interface {
	Register() error
	KubeConfig() KubeConfigClient
	Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *KubeconfigsSnapshot, <-chan error, error)
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
	var done sync.WaitGroup
	ctx := opts.Ctx
	/* Create channel for KubeConfig */
	type kubeConfigListWithNamespace struct {
		list      KubeConfigList
		namespace string
	}
	kubeConfigChan := make(chan kubeConfigListWithNamespace)

	var initialKubeConfigList KubeConfigList

	currentSnapshot := KubeconfigsSnapshot{}

	for _, namespace := range watchNamespaces {
		/* Setup namespaced watch for KubeConfig */
		{
			kubeconfigs, err := c.kubeConfig.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial KubeConfig list")
			}
			initialKubeConfigList = append(initialKubeConfigList, kubeconfigs...)
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
				case kubeConfigList := <-kubeConfigNamespacesChan:
					select {
					case <-ctx.Done():
						return
					case kubeConfigChan <- kubeConfigListWithNamespace{list: kubeConfigList, namespace: namespace}:
					}
				}
			}
		}(namespace)
	}
	/* Initialize snapshot for Kubeconfigs */
	currentSnapshot.Kubeconfigs = initialKubeConfigList.Sort()

	snapshots := make(chan *KubeconfigsSnapshot)
	go func() {
		// sent initial snapshot to kick off the watch
		initialSnapshot := currentSnapshot.Clone()
		snapshots <- &initialSnapshot

		timer := time.NewTicker(time.Second * 1)
		previousHash := currentSnapshot.Hash()
		sync := func() {
			currentHash := currentSnapshot.Hash()
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
		kubeconfigsByNamespace := make(map[string]KubeConfigList)

		for {
			record := func() { stats.Record(ctx, mKubeconfigsSnapshotIn.M(1)) }

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
			case kubeConfigNamespacedList := <-kubeConfigChan:
				record()

				namespace := kubeConfigNamespacedList.namespace

				stats.RecordWithTags(
					ctx,
					[]tag.Mutator{tag.Insert(kubeconfigsNamespaceKey, namespace)},
					mKubeconfigsKubeconfigsListIn.M(1),
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
