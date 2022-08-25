package templates

import (
	"text/template"
)

// Snapshot Emitters are used to take snapshots of the current system, using either
// cluster scoped or non namespaced scoped selection. Watches are used to notify
// the snapshot emitter when new resources have been created or updated.
// Snapshot Emitters will delegate to Resource Clients to list and watch defined
// resources.

// ClusterScoped - without namespacing, get all the resources within the entire cluster. There is one watch per resource.
// Not using ClusterScoped - allows for using namespacing, so that each namespace has it's own watch per resource.
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
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	skstats "github.com/solo-io/solo-kit/pkg/stats"
	
	"github.com/solo-io/go-utils/errutils"
	"github.com/solo-io/go-utils/contextutils"
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

func New{{ .GoName }}Emitter({{ $client_declarations }}, resourceNamespaceLister resources.ResourceNamespaceLister) {{ .GoName }}Emitter {
	return New{{ .GoName }}EmitterWithEmit({{ $clients }}, resourceNamespaceLister, make(chan struct{}))
}

func New{{ .GoName }}EmitterWithEmit({{ $client_declarations }}, resourceNamespaceLister resources.ResourceNamespaceLister, emit <-chan struct{}) {{ .GoName }}Emitter {
	return &{{ lower_camel .GoName }}Emitter{
{{- range .Resources}}
		{{ lower_camel .Name }}:{{ lower_camel .Name }}Client,
{{- end}}
		resourceNamespaceLister: resourceNamespaceLister,
		forceEmit: emit,
	}
}

type {{ lower_camel .GoName }}Emitter struct {
	forceEmit <- chan struct{}
{{- range .Resources}}
	{{ lower_camel .Name }} {{ .ImportPrefix }}{{ .Name }}Client
{{- end}}
	resourceNamespaceLister resources.ResourceNamespaceLister
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

	// TODO-JAKE some of this should only be present if scoped by namespace
	errs := make(chan error)
	hasWatchedNamespaces :=  len(watchNamespaces) > 1 || (len(watchNamespaces) == 1 && watchNamespaces[0] != "")
	watchNamespacesIsEmpty := ! hasWatchedNamespaces
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
	{{ lower_camel .PluralName }}ByNamespace := sync.Map{}
{{- end }}
{{- end }}

	// watched namespaces
	for _, namespace := range watchNamespaces {
{{- range .Resources}}
{{- if (not .ClusterScoped) }}
		/* Setup namespaced watch for {{ .Name }} */
		{
			{{ lower_camel .PluralName }}, err := c.{{ lower_camel .Name }}.List(namespace, watchedNamespacesListOptions)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "initial {{ .Name }} list")
			}
			initial{{ upper_camel .Name }}List = append(initial{{ upper_camel .Name }}List, {{ lower_camel .PluralName }}...)
			{{ lower_camel .PluralName }}ByNamespace.Store(namespace, {{ lower_camel .PluralName }})
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
			for _,wns := range watchNamespaces {
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
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
			/* Setup namespaced watch for {{ upper_camel .Name }} */
			{
				{{ lower_camel .PluralName }}, err := c.{{ lower_camel .Name }}.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial {{ upper_camel .Name }} list")
				}
				initial{{ upper_camel .Name }}List = append(initial{{ upper_camel .Name }}List,{{ lower_camel .PluralName }}...)
				{{ lower_camel .PluralName }}ByNamespace.Store(namespace, {{ lower_camel .PluralName }})
			}
			{{ lower_camel .Name }}NamespacesChan, {{ lower_camel .Name }}Errs, err := c.{{ lower_camel .Name }}.Watch(namespace, opts)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "starting {{ upper_camel .Name }} watch")
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
						
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
						// TODO-JAKE we willl only need to do this once, I might be best to keep a set/map of the current
						// namespaces that are used
						if _, hit := {{ lower_camel .PluralName }}ByNamespace.Load(ns.Name); !hit {
							newNamespaces = append(newNamespaces, ns.Name)
							continue
						}
{{- end }}
{{- end }}
					}
					// TODO-JAKE I think we could get rid of this if statement if needed.
					if len(newNamespaces) > 0{
						// add a watch for all the new namespaces
						// REFACTOR
						for _, namespace := range newNamespaces {
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
							/* Setup namespaced watch for {{ upper_camel .Name }} for new namespace */
							{
								{{ lower_camel .PluralName }}, err := c.{{ lower_camel .Name }}.List(namespace, clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector})
								if err != nil {
									// INFO-JAKE not sure if we want to do something else
									// but since this is occuring in async I think it should be fine
									errs <- errors.Wrapf(err, "initial new namespace {{ upper_camel .Name }} list")
									continue
								}
								{{ lower_camel .PluralName }}ByNamespace.Store(namespace, {{ lower_camel .PluralName }})
							}
							{{ lower_camel .Name }}NamespacesChan, {{ lower_camel .Name }}Errs, err := c.{{ lower_camel .Name }}.Watch(namespace, opts)
							if err != nil {
								// TODO-JAKE if we do decide to have the namespaceErrs from the watch namespaces functionality
								// , then we could add it here namespaceErrs <- error(*) . the namespaceErrs is coming from the
								// ResourceNamespaceLister currently
								// INFO-JAKE is this what we really want to do when there is an error?
								errs <- errors.Wrapf(err, "starting new namespace {{ upper_camel .Name }} watch")
								continue
							}

							// INFO-JAKE I think this is appropriate, becasue
							// we want to watch the errors coming off the namespace
							done.Add(1)
							go func(namespace string) {
								defer done.Done()
								errutils.AggregateErrs(ctx, errs, {{ lower_camel .Name }}Errs, namespace+"-new-namespace-{{ lower_camel .PluralName }}")
							}(namespace)
{{- end }}
{{- end }}
							/* Watch for changes and update snapshot */
							// REFACTOR
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
		}()
	}
{{- range .Resources}}
{{- if .ClusterScoped }}
	// TODO-JAKE verify that this is what we should be doing with Cluster Scoped Resources
	/* Setup cluster-wide watch for {{ .Name }} */
	var err error
	currentSnapshot.{{ upper_camel .PluralName }},err = c.{{ lower_camel .Name }}.List(clients.ListOpts{Ctx: opts.Ctx, ExpressionSelector: opts.ExpressionSelector, Selector: opts.Selector})
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
				{{ lower_camel .PluralName }}ByNamespace.Store(namespace, {{ lower_camel .Name }}NamespacedList.list)
				var {{ lower_camel .Name }}List {{ .ImportPrefix }}{{ .Name }}List
				{{ lower_camel .PluralName }}ByNamespace.Range(func(key interface{}, value interface{}) bool {
					mocks := value.({{ .ImportPrefix }}{{ .Name }}List)
					{{ lower_camel .Name }}List = append({{ lower_camel .Name }}List, mocks...)
					return true
				})
				currentSnapshot.{{ upper_camel .PluralName }} = {{ lower_camel .Name }}List.Sort()
{{- end }}
{{- end }}
			}
		}
	}()
	return snapshots, errs, nil
}
`))
