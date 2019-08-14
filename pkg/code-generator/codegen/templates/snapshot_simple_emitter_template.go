package templates

import (
	"text/template"
)

var SimpleEmitterTemplate = template.Must(template.New("resource_group_emitter").Funcs(Funcs).Parse(
	`package {{ .Project.ProjectConfig.Version }}

import (
	"context"
	"sync"
	"time"

	{{ .Imports }}
	"go.opencensus.io/stats"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/go-utils/errutils"
)


type {{ .GoName }}SimpleEmitter interface {
	Snapshots(ctx context.Context) (<-chan *{{ .GoName }}Snapshot, <-chan error, error)
}

func New{{ .GoName }}SimpleEmitter(aggregatedWatch clients.ResourceWatch) {{ .GoName }}SimpleEmitter {
	return New{{ .GoName }}SimpleEmitterWithEmit(aggregatedWatch, make(chan struct{}))
}

func New{{ .GoName }}SimpleEmitterWithEmit(aggregatedWatch clients.ResourceWatch, emit <-chan struct{}) {{ .GoName }}SimpleEmitter {
	return &{{ lower_camel .GoName }}SimpleEmitter{
		aggregatedWatch: aggregatedWatch,
		forceEmit: emit,
	}
}

type {{ lower_camel .GoName }}SimpleEmitter struct {
	forceEmit <- chan struct{}
	aggregatedWatch clients.ResourceWatch
}

func (c *{{ lower_camel .GoName }}SimpleEmitter) Snapshots(ctx context.Context) (<-chan *{{ .GoName }}Snapshot, <-chan error, error) {
	snapshots := make(chan *{{ .GoName }}Snapshot)
	errs := make(chan error)
	
	untyped, watchErrs, err := c.aggregatedWatch(ctx)
	if err != nil {
		return nil, nil, err
	}

	go errutils.AggregateErrs(ctx, errs, watchErrs, "{{ lower_camel .GoName }}-emitter")

	go func() {
		currentSnapshot := {{ .GoName }}Snapshot{}
		timer := time.NewTicker(time.Second * 1)
		var previousHash uint64
		sync := func() {
			currentHash := currentSnapshot.Hash()
			if previousHash == currentHash {
				return
			}

			previousHash = currentHash

			stats.Record(ctx, m{{ .GoName }}SnapshotOut.M(1))
			sentSnapshot := currentSnapshot.Clone()
			snapshots <- &sentSnapshot
		}

		defer func() {
			close(snapshots)
			close(errs)
		}()

		for {
			record := func() { stats.Record(ctx, m{{ .GoName }}SnapshotIn.M(1)) }

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

				currentSnapshot = {{ .GoName }}Snapshot{}
				for _, res := range untypedList {
					switch typed := res.(type) {
{{- range .Resources}}
					case *{{ .ImportPrefix }}{{ .Name }}:
						currentSnapshot.{{ upper_camel .PluralName }} = append(currentSnapshot.{{ upper_camel .PluralName }}, typed)
{{- end}}
					default:
						select {
						case errs <- fmt.Errorf("{{ .GoName }}SnapshotEmitter "+
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
`))
