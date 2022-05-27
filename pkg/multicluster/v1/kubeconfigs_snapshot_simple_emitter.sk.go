// Code generated by solo-kit. DO NOT EDIT.

//Source: pkg/code-generator/codegen/templates/snapshot_simple_emitter_template.go
package v1

import (
	"context"
	"fmt"
	"time"

	"go.opencensus.io/stats"
	"go.uber.org/zap"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
)

type KubeconfigsSimpleEmitter interface {
	Snapshots(ctx context.Context) (<-chan *KubeconfigsSnapshot, <-chan error, error)
}

func NewKubeconfigsSimpleEmitter(aggregatedWatch clients.ResourceWatch) KubeconfigsSimpleEmitter {
	return NewKubeconfigsSimpleEmitterWithEmit(aggregatedWatch, make(chan struct{}))
}

func NewKubeconfigsSimpleEmitterWithEmit(aggregatedWatch clients.ResourceWatch, emit <-chan struct{}) KubeconfigsSimpleEmitter {
	return &kubeconfigsSimpleEmitter{
		aggregatedWatch: aggregatedWatch,
		forceEmit:       emit,
	}
}

type kubeconfigsSimpleEmitter struct {
	forceEmit       <-chan struct{}
	aggregatedWatch clients.ResourceWatch
}

func (c *kubeconfigsSimpleEmitter) Snapshots(ctx context.Context) (<-chan *KubeconfigsSnapshot, <-chan error, error) {
	snapshots := make(chan *KubeconfigsSnapshot)
	errs := make(chan error)

	untyped, watchErrs, err := c.aggregatedWatch(ctx)
	if err != nil {
		return nil, nil, err
	}

	go errutils.AggregateErrs(ctx, errs, watchErrs, "kubeconfigs-emitter")

	go func() {
		currentSnapshot := KubeconfigsSnapshot{}
		timer := time.NewTicker(time.Second * 1)
		var previousHash uint64
		sync := func() {
			currentHash, err := currentSnapshot.Hash(nil)
			if err != nil {
				contextutils.LoggerFrom(ctx).Panicw("error while hashing, this should never happen", zap.Error(err))
			}
			if previousHash == currentHash {
				return
			}

			previousHash = currentHash

			stats.Record(ctx, mKubeconfigsSnapshotOut.M(1))
			sentSnapshot := currentSnapshot.Clone()
			snapshots <- &sentSnapshot
		}

		defer func() {
			close(snapshots)
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
			case untypedList := <-untyped:
				record()

				currentSnapshot = KubeconfigsSnapshot{}
				for _, res := range untypedList {
					switch typed := res.(type) {
					case *KubeConfig:
						currentSnapshot.Kubeconfigs = append(currentSnapshot.Kubeconfigs, typed)
					default:
						select {
						case errs <- fmt.Errorf("KubeconfigsSnapshotEmitter "+
							"cannot process resource %v of type %T", res.GetMetadata().Ref(), res):
						case <-ctx.Done():
							return
						}
					}
				}

			}
		}
	}()
	return snapshots, errs, nil
}
