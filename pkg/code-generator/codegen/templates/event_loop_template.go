package templates

import (
	"text/template"
)

var ResourceGroupEventLoopTemplate = template.Must(template.New("resource_group_event_loop").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"github.com/hashicorp/go-multierror"

	skstats "github.com/solo-io/solo-kit/pkg/stats"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errutils"
)


var (
	m{{ .GoName }}SnapshotTimeSec     = stats.Float64("{{ .Name }}/sync/time_sec", "The time taken for a given sync", "1")
	m{{ .GoName }}SnapshotTimeSecView = &view.View{
		Name:        "{{ .Name }}/sync/time_sec",
		Description: "The time taken for a given sync",
		TagKeys:     []tag.Key{tag.MustNewKey("syncer_name")},
		Measure:     m{{ .GoName }}SnapshotTimeSec,
		Aggregation: view.Distribution(0.01, 0.05, 0.1, 0.25, 0.5, 1, 5, 10, 60),
	}
)


func init() {
	view.Register(
		m{{ .GoName }}SnapshotTimeSecView,
	)
}

type {{ .GoName }}Syncer interface {
	Sync(context.Context, *{{ .GoName }}Snapshot) error
}

type {{ .GoName }}Syncers []{{ .GoName }}Syncer

func (s {{ .GoName }}Syncers) Sync(ctx context.Context, snapshot *{{ .GoName }}Snapshot) error {
	var multiErr *multierror.Error
	for _, syncer := range s {
		if err := syncer.Sync(ctx, snapshot); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr.ErrorOrNil()
}

type {{ lower_camel .GoName }}EventLoop struct {
	emitter {{ .GoName }}SnapshotEmitter
	syncer  {{ .GoName }}Syncer
	ready chan struct{}
}

func New{{ .GoName }}EventLoop(emitter {{ .GoName }}SnapshotEmitter, syncer {{ .GoName }}Syncer) eventloop.EventLoop {
	return &{{ lower_camel .GoName }}EventLoop{
		emitter: emitter,
		syncer:  syncer,
		ready: make(chan struct{}),
	}
}


func (el *{{ lower_camel .GoName }}EventLoop) Ready() <-chan struct{} {
	return el.ready
}

func (el *{{ lower_camel .GoName }}EventLoop) Run(namespaces []string, opts clients.WatchOpts) (<-chan error, error) {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "{{ .Project.ProjectConfig.Version }}.event_loop")
	logger := contextutils.LoggerFrom(opts.Ctx)
	logger.Infof("event loop started")

	errs := make(chan error)

	watch, emitterErrs, err := el.emitter.Snapshots(namespaces, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "starting snapshot watch")
	}
	go errutils.AggregateErrs(opts.Ctx, errs, emitterErrs, "{{ .Project.ProjectConfig.Version }}.emitter errors")
	go func() {
		var channelClosed bool
		// create a new context for each loop, cancel it before each loop
		var cancel context.CancelFunc = func() {}
		// use closure to allow cancel function to be updated as context changes
		defer func() { cancel() }()
		for {
			select {
			case snapshot, ok := <-watch:
				if !ok {
					return
				}
				// cancel any open watches from previous loop
				cancel()

				startTime := time.Now()
				ctx, span := trace.StartSpan(opts.Ctx, "{{ .Name }}.EventLoopSync")
				ctx, canc := context.WithCancel(ctx)
				cancel = canc
				err := el.syncer.Sync(ctx, snapshot)
				stats.RecordWithTags(
					ctx,
					[]tag.Mutator{
						tag.Insert(skstats.SyncerNameKey, fmt.Sprintf("%T", el.syncer)),
					},
					m{{ .GoName }}SnapshotTimeSec.M(time.Now().Sub(startTime).Seconds()),
				)
				span.End()

				if err != nil {
					select {
					case errs <- err:
					default:
						logger.Errorf("write error channel is full! could not propagate err: %v", err)
					}
				} else if !channelClosed {
					channelClosed = true
					close(el.ready)
				}
			case <-opts.Ctx.Done():
				return
			}
		}
	}()
	return errs, nil
}
`))
