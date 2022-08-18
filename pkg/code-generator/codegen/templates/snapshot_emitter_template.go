package templates

import (
	"text/template"
)

var ResourceGroupEmitterTemplate = template.Must(template.New("resource_group_emitter").Funcs(Funcs).Parse(
	`package {{ .Project.ProjectConfig.Version }}

{{- $client_declarations := new_str_slice }}
{{- $clients := new_str_slice }}
{{- range .Resources}}
{{- $client_declarations := (append_str_slice $client_declarations (printf "%vClient %v%vClient"  (lower_camel .Name) .ImportPrefix .Name)) }}
{{- $clients := (append_str_slice $clients (printf "%vClient"  (lower_camel .Name))) }}
{{- end}}
{{- $client_declarations := (join_str_slice $client_declarations ", ") }}
{{- $clients := (join_str_slice $clients ", ") }}

import (
	"bytes"
	"sync"
	"time"

	{{ .Imports }}
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"


	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/errors"
	skstats "github.com/solo-io/solo-kit/pkg/stats"
	
	"github.com/solo-io/go-utils/errutils"
	"github.com/solo-io/go-utils/contextutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubewatch "k8s.io/apimachinery/pkg/watch"
)

{{ $emitter_prefix := (print (snake .Name) "/emitter") }}
{{ $resource_group := upper_camel .GoName }}
var (
	// Deprecated. See m{{ $resource_group }}ResourcesIn
	m{{ $resource_group }}SnapshotIn  = stats.Int64("{{ $emitter_prefix }}/snap_in", "Deprecated. Use {{ $emitter_prefix }}/resources_in. The number of snapshots in", "1")
	
	// metrics for emitter
	m{{ $resource_group }}ResourcesIn = stats.Int64("{{ $emitter_prefix }}/resources_in", "The number of resource lists received on open watch channels", "1")
	m{{ $resource_group }}SnapshotOut = stats.Int64("{{ $emitter_prefix }}/snap_out", "The number of snapshots out", "1")
	m{{ $resource_group }}SnapshotMissed = stats.Int64("{{ $emitter_prefix }}/snap_missed", "The number of snapshots missed", "1")

	// views for emitter
	// deprecated: see {{ lower_camel .GoName }}ResourcesInView
	{{ lower_camel .GoName }}snapshotInView = &view.View{
		Name:        "{{ $emitter_prefix }}/snap_in",
		Measure:     m{{ $resource_group }}SnapshotIn,
		Description: "Deprecated. Use {{ $emitter_prefix }}/resources_in. The number of snapshots updates coming in.",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{
		},
	}

	{{ lower_camel .GoName }}ResourcesInView = &view.View{
			Name:        "{{ $emitter_prefix }}/resources_in",
			Measure:     m{{ $resource_group }}ResourcesIn,
			Description: "The number of resource lists received on open watch channels",
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{
				skstats.NamespaceKey,
				skstats.ResourceKey,
			},
	}
	{{ lower_camel .GoName }}snapshotOutView = &view.View{
		Name:        "{{ $emitter_prefix }}/snap_out",
		Measure:     m{{ $resource_group }}SnapshotOut,
		Description: "The number of snapshots updates going out",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{
		},
	}
	{{ lower_camel .GoName }}snapshotMissedView = &view.View{
			Name:        "{{ $emitter_prefix }}/snap_missed",
			Measure:     m{{ $resource_group }}SnapshotMissed,
			Description: "The number of snapshots updates going missed. this can happen in heavy load. missed snapshot will be re-tried after a second.",
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{
			},
	}


)

func init() {
	view.Register(
		{{ lower_camel .GoName }}snapshotInView, 
		{{ lower_camel .GoName }}snapshotOutView, 
		{{ lower_camel .GoName }}snapshotMissedView,
		{{ lower_camel .GoName }}ResourcesInView,
	)
}

type {{ .GoName }}SnapshotEmitter interface {
	Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *{{ .GoName }}Snapshot, <-chan error, error)
}

type {{ .GoName }}Emitter interface {
	{{ .GoName }}SnapshotEmitter
	Register() error
{{- range .Resources}}
	{{ .Name }}() {{ .ImportPrefix }}{{ .Name }}Client
{{- end}}
}

func New{{ .GoName }}Emitter({{ $client_declarations }}) {{ .GoName }}Emitter {
	return New{{ .GoName }}EmitterWithEmit({{ $clients }}, make(chan struct{}))
}

func New{{ .GoName }}EmitterWithEmit({{ $client_declarations }}, emit <-chan struct{}) {{ .GoName }}Emitter {
	return &{{ lower_camel .GoName }}Emitter{
{{- range .Resources}}
		{{ lower_camel .Name }}:{{ lower_camel .Name }}Client,
{{- end}}
		forceEmit: emit,
	}
}

type {{ lower_camel .GoName }}Emitter struct {
	forceEmit <- chan struct{}
{{- range .Resources}}
	{{ lower_camel .Name }} {{ .ImportPrefix }}{{ .Name }}Client
{{- end}}
}

func (c *{{ lower_camel .GoName }}Emitter) Register() error {
{{- range .Resources}}
	if err := c.{{ lower_camel .Name }}.Register(); err != nil {
		return err
	}
{{- end}}
	return nil
}

{{- range .Resources}}

func (c *{{ lower_camel $.GoName }}Emitter) {{ .Name }}() {{ .ImportPrefix }}{{ .Name }}Client {
	return c.{{ lower_camel .Name }}
}
{{- end}}

func (c *{{ lower_camel .GoName }}Emitter) Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *{{ .GoName }}Snapshot, <-chan error, error) {

	if len(watchNamespaces) == 0 {
		watchNamespaces = []string{""}
	}

	for _, ns := range watchNamespaces {
		if ns == "" && len(watchNamespaces) > 1 {
			return nil, nil, errors.Errorf("the \"\" namespace is used to watch all namespaces. Snapshots can either be tracked for "+
				"specific namespaces or \"\" AllNamespaces, but not both.")
		}
	}

	errs := make(chan error)
	hasWatchedNamespaces :=  len(watchNamespaces) > 1 || (len(watchNamespaces) == 1 && watchNamespaces[0] != "")
	watchNamespacesIsEmpty := ! hasWatchedNamespaces
	var done sync.WaitGroup
	ctx := opts.Ctx

	// if we are watching namespaces, then we do not want to fitler any of the 
	// resources in when listing or watching
	watchedNamespacesListOptions := clients.ListOpts{Ctx: opts.Ctx}
	watchedNamespacesWatchOptions := clients.WatchOpts{Ctx: opts.Ctx}
	if watchNamespacesIsEmpty {
		// if the namespaces that we are watching is empty, then we want to apply
		// the expression Selectors to all the namespaces.
		watchedNamespacesListOptions.ExpressionSelector = opts.ExpressionSelector
		watchedNamespacesWatchOptions.ExpressionSelector = opts.ExpressionSelector
	}

{{- range .Resources}}
	/* Create channel for {{ .Name }} */
{{- if (not .ClusterScoped) }}
type {{ lower_camel .Name }}ListWithNamespace struct {
	list {{ .ImportPrefix }}{{ .Name }}List
	namespace string
}
	{{ lower_camel .Name }}Chan := make(chan {{ lower_camel .Name }}ListWithNamespace)
	var initial{{ upper_camel .Name }}List {{ .ImportPrefix }}{{ .Name }}List
{{- end }}
{{- end }}

	currentSnapshot := {{ .GoName }}Snapshot{}

	{{- range .Resources}}
	{{- if not .ClusterScoped }}
			{{ lower_camel .PluralName }}ByNamespace := make(map[string]{{ .ImportPrefix }}{{ .Name }}List)
	{{- end }}
	{{- end }}

	// watched namespaces
	for _, namespace := range watchNamespaces {
{{- range .Resources}}
{{- if (not .ClusterScoped) }}
		/* Setup namespaced watch for {{ .Name }} */
		{
			// TODO-JAKE not sure if we want to get rid of the Selector in the
			// ListOpts here. the reason that we might want to is because we no
			// longer allow selectors, unless it is on a unwatched namespace.
			{{ lower_camel .PluralName }}, err := c.{{ lower_camel .Name }}.List(namespace, watchedNamespacesListOptions)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial {{ .Name }} list")
			}
			initial{{ upper_camel .Name }}List = append(initial{{ upper_camel .Name }}List, {{ lower_camel .PluralName }}...)
			{{ lower_camel .PluralName }}ByNamespace[namespace] = {{ lower_camel .PluralName }}
		}
		{{ lower_camel .Name }}NamespacesChan, {{ lower_camel .Name }}Errs, err := c.{{ lower_camel .Name }}.Watch(namespace, watchedNamespacesWatchOptions)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "starting {{ .Name }} watch")
		}

		done.Add(1)
		go func(namespace string) {
			defer done.Done()
			errutils.AggregateErrs(ctx, errs, {{ lower_camel .Name }}Errs, namespace+"-{{ lower_camel .PluralName }}")
		}(namespace)

{{- end }}
{{- end }}

		/* Watch for changes and update snapshot */
		go func(namespace string) {
			for {
				select {
				case <-ctx.Done():
					return
{{- range .Resources}}
{{- if (not .ClusterScoped) }}
				case {{ lower_camel .Name }}List, ok := <- {{ lower_camel .Name }}NamespacesChan:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case {{ lower_camel .Name }}Chan <- {{ lower_camel .Name }}ListWithNamespace{list:{{ lower_camel .Name }}List, namespace:namespace}:
					}
{{- end }}
{{- end}}
				}
			}
		}(namespace)
	}
{{- if (not .ClusterScoped) }}
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

		namespacesResources,err := k.CoreV1().Namespaces().List(ctx, metav1.ListOptions{FieldSelector: excludeNamespacesFieldDesciptors})
		if err != nil {
			return nil, nil, err
		}
		allOtherNamespaces := make([]string, len(namespacesResources.Items))
		for _, ns := range namespacesResources.Items {
			allOtherNamespaces = append(allOtherNamespaces, ns.Namespace)
		}

		// nonWatchedNamespaces
		for _, namespace := range allOtherNamespaces {
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
			/* Setup namespaced watch for {{ upper_camel .Name }} */
			{
				{{ lower_camel .Name }}s, err := c.{{ lower_camel .Name }}.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial {{ upper_camel .Name }} list")
				}
				initial{{ upper_camel .Name }}List = append(initial{{ upper_camel .Name }}List, {{ lower_camel .Name }}s...)
				{{ lower_camel .Name }}sByNamespace[namespace] = {{ lower_camel .Name }}s
			}
			{{ lower_camel .Name }}NamespacesChan, {{ lower_camel .Name }}Errs, err := c.{{ lower_camel .Name }}.Watch(namespace, opts)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "starting {{ upper_camel .Name }} watch")
			}

			done.Add(1)
			go func(namespace string) {
				defer done.Done()
				errutils.AggregateErrs(ctx, errs, {{ lower_camel .Name }}Errs, namespace+"-{{ lower_camel .Name }}s")
			}(namespace)
{{- end }}
{{- end }}
			/* Watch for changes and update snapshot */
			go func(namespace string) {
				for {
					select {
					case <-ctx.Done():
						return
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
					case {{ lower_camel .Name }}List, ok := <-{{ lower_camel .Name }}NamespacesChan:
						if !ok {
							return
						}
						select {
						case <-ctx.Done():
							return
						case {{ lower_camel .Name }}Chan <- {{ lower_camel .Name }}ListWithNamespace{list: {{ lower_camel .Name }}List, namespace: namespace}:
						}
{{- end }}
{{- end }}
					}
				}
			}(namespace)
		}
		// create watch on all namespaces, so that we can add resources from new namespaces
		namespaceWatch,err := k.CoreV1().Namespaces().Watch(opts.Ctx, metav1.ListOptions{FieldSelector: excludeNamespacesFieldDesciptors})
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
						// we get an event	
						namespacesResources, err := k.CoreV1().Namespaces().List(opts.Ctx, metav1.ListOptions{FieldSelector: excludeNamespacesFieldDesciptors})
						if err != nil {
							errs <- errors.Wrapf(err, "listing the namespace resources")
						}
						for _,item := range namespacesResources.Items {
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
							namespace := item.Namespace
							_, hit := {{ lower_camel .Name }}sByNamespace[namespace]
							if ! hit {
								/* Setup namespaced watch for {{ upper_camel .Name }} */
								{
									{{ lower_camel .Name }}s, err := c.{{ lower_camel .Name }}.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
									if err != nil {
										// INFO-JAKE pretty sure we want to send
										// an error message here, but we might want
										// to do something else.
										errs <- errors.Wrapf(err, "initial new namespace {{ upper_camel .Name }} list")
										continue
									}
									{{ lower_camel .Name }}sByNamespace[namespace] = {{ lower_camel .Name }}s
								}
								{{ lower_camel .Name }}NamespacesChan, {{ lower_camel .Name }}Errs, err := c.{{ lower_camel .Name }}.Watch(namespace, opts)
								if err != nil {
									// INFO-JAKE is this what we really want to do when there is an error?
									errs <- errors.Wrapf(err, "starting new namespace {{ upper_camel .Name }} watch")
									continue
								}

								// INFO-JAKE I think this is appropriate, becasue
								// we want to watch the errors coming off the namespace
								done.Add(1)
								go func(namespace string) {
									defer done.Done()
									errutils.AggregateErrs(ctx, errs, {{ lower_camel .Name }}Errs, namespace+"-new-namespace-{{ lower_camel .Name }}s")
								}(namespace)
{{- end }}
{{- end }}
								// if that is the case, return the {{ lower_camel .Name }}NamespacesChan....
								/* Watch for changes and update snapshot */
								// INFO-JAKE I would like to refactor this code
								// is there a way we can refactor all this code 
								// in the go routine once?
								go func(namespace string) {
									for {
										select {
										case <-ctx.Done():
											return
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
										case {{ lower_camel .Name }}List, ok := <-{{ lower_camel .Name }}NamespacesChan:
											if !ok {
												return
											}
											select {
											case <-ctx.Done():
												return
											case {{ lower_camel .Name }}Chan <- {{ lower_camel .Name }}ListWithNamespace{list: {{ lower_camel .Name }}List, namespace: namespace}:
											}
{{- end }}
{{- end }}
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
{{- end }}
{{- range .Resources}}
{{- if .ClusterScoped }}
	// TODO-JAKE not sure if ther is anything that will need to be done for cluster
	// scoped resources.  Verify if this is even used in Gloo or not.
	/* Setup cluster-wide watch for {{ .Name }} */
	var err error
	currentSnapshot.{{ upper_camel .PluralName }},err = c.{{ lower_camel .Name }}.List(clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "initial {{ .Name }} list")
	}
	{{ lower_camel .Name }}Chan, {{ lower_camel .Name }}Errs, err := c.{{ lower_camel .Name }}.Watch(opts)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "starting {{ .Name }} watch")
	}
	done.Add(1)
	go func() {
		defer done.Done()
		errutils.AggregateErrs(ctx, errs, {{ lower_camel .Name }}Errs, "{{ lower_camel .PluralName }}")
	}()

{{- else }}
	/* Initialize snapshot for {{ upper_camel .PluralName }} */
	currentSnapshot.{{ upper_camel .PluralName }} = initial{{ upper_camel .Name }}List.Sort()
{{- end }}
{{- end}}

	snapshots := make(chan *{{ .GoName }}Snapshot)
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
				stats.Record(ctx, m{{ $resource_group }}SnapshotOut.M(1))
				previousHash = currentHash
			default:
				stats.Record(ctx, m{{ $resource_group }}SnapshotMissed.M(1))
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
			record := func(){stats.Record(ctx, m{{ $resource_group }}SnapshotIn.M(1))}
			
			select {
			case <-timer.C:
				sync()
			case <-ctx.Done():
				return
			case <-c.forceEmit:
				sentSnapshot := currentSnapshot.Clone()
				snapshots <- &sentSnapshot
{{- range .Resources}}
{{- if .ClusterScoped }}
			case {{ lower_camel .Name }}List, ok := <- {{ lower_camel .Name }}Chan:
				if !ok {
					return
				}
				record()

				skstats.IncrementResourceCount(
					ctx,
					"<all>",
					"{{ snake .Name }}",
					m{{ $resource_group }}ResourcesIn,
				)

				currentSnapshot.{{ upper_camel .PluralName }} = {{ lower_camel .Name }}List
{{- else }}
			case {{ lower_camel .Name }}NamespacedList, ok := <- {{ lower_camel .Name }}Chan:
				if !ok {
					return
				}
				record()

				namespace := {{ lower_camel .Name }}NamespacedList.namespace

				skstats.IncrementResourceCount(
					ctx,
					namespace,
					"{{ snake .Name }}",
					m{{ $resource_group }}ResourcesIn,
				)

				// merge lists by namespace
				{{ lower_camel .PluralName }}ByNamespace[namespace] = {{ lower_camel .Name }}NamespacedList.list
				var {{ lower_camel .Name }}List {{ .ImportPrefix }}{{ .Name }}List
				for _, {{ lower_camel .PluralName }} := range {{ lower_camel .PluralName }}ByNamespace {
					{{ lower_camel .Name }}List  = append({{ lower_camel .Name }}List, {{ lower_camel .PluralName }}...)
				}
				currentSnapshot.{{ upper_camel .PluralName }} = {{ lower_camel .Name }}List.Sort()
{{- end }}
{{- end }}
			}
		}
	}()
	return snapshots, errs, nil
}
`))
