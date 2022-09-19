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
// this means that Expression Selectors will have no impact on namespaces.
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
	// resourceNamespaceLister is used to watch for new namespaces when they are created.
	// It is used when Expression Selector is in the Watch Opts set in Snapshot().
	resourceNamespaceLister resources.ResourceNamespaceLister
	// namespacesWatching is the set of namespaces that we are watching. This is helpful
	// when Expression Selector is set on the Watch Opts in Snapshot().
	namespacesWatching sync.Map
	// updateNamespaces is used to perform locks and unlocks when watches on namespaces are being updated/created
	updateNamespaces sync.Mutex
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

// Snapshots will return a channel that can be used to receive snapshots of the 
// state of the resources it is watching
// when watching resources, you can set the watchNamespaces, and you can set the 
// ExpressionSelector of the WatchOpts.  Setting watchNamespaces will watch for all resources
// that are in the specified namespaces. In addition if ExpressionSelector of the WatchOpts is
// set, then all namespaces that meet the label criteria of the ExpressionSelector will
// also be watched.
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
	watchingLabeledNamespaces := ! (opts.ExpressionSelector == "")
	var done sync.WaitGroup
	ctx := opts.Ctx

	// setting up the options for both listing and watching resources in namespaces
	watchedNamespacesListOptions := clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector }
	watchedNamespacesWatchOptions := clients.WatchOpts{Ctx: opts.Ctx, Selector: opts.Selector }

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
	if hasWatchedNamespaces || ! watchingLabeledNamespaces {
		// then watch all resources on watch Namespaces

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
				defer func() {
					c.namespacesWatching.Delete(namespace)
				}()
				c.namespacesWatching.Store(namespace, true)
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
	}
	// watch all other namespaces that are labeled and fit the Expression Selector
	if opts.ExpressionSelector != "" {
		// watch resources of non-watched namespaces that fit the expression selectors
		namespaceListOptions := resources.ResourceNamespaceListOptions{
			Ctx: opts.Ctx,
			ExpressionSelector: opts.ExpressionSelector,
		}
		namespaceWatchOptions := resources.ResourceNamespaceWatchOptions{
			Ctx: opts.Ctx,
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
			c.namespacesWatching.Load(resourceNamespace)
			namespace := resourceNamespace.Name
			newlyRegisteredNamespaces[i] = namespace
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
			err = c.{{ lower_camel .Name }}.RegisterNamespace(namespace)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "there was an error registering the namespace to the {{ lower_camel .Name }}")
			}			
			/* Setup namespaced watch for {{ upper_camel .Name }} */
			{
				{{ lower_camel .PluralName }}, err := c.{{ lower_camel .Name }}.List(namespace, clients.ListOpts{Ctx: opts.Ctx})
				if err != nil {
					return nil, nil, errors.Wrapf(err, "initial {{ upper_camel .Name }} list with new namespace")
				}
				initial{{ upper_camel .Name }}List = append(initial{{ upper_camel .Name }}List,{{ lower_camel .PluralName }}...)
				{{ lower_camel .PluralName }}ByNamespace.Store(namespace, {{ lower_camel .PluralName }})
			}
			{{ lower_camel .Name }}NamespacesChan, {{ lower_camel .Name }}Errs, err := c.{{ lower_camel .Name }}.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx})
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
					select{
					case <-ctx.Done():
						return
					case err = <- errsReceiver:
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
{{- range .Resources}}
{{- if (not .ClusterScoped) }}
						{{ lower_camel .Name }}Chan <- {{ lower_camel .Name }}ListWithNamespace{list: {{ .ImportPrefix }}{{ .Name }}List{}, namespace: ns}
{{- end }}
{{- end }}
					}

					for _, namespace := range newNamespaces {
						var err error
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
						err = c.{{ lower_camel .Name }}.RegisterNamespace(namespace)
						if err != nil {
							errs <- errors.Wrapf(err, "there was an error registering the namespace to the {{ lower_camel .Name }}")
							continue
						}
						/* Setup namespaced watch for {{ upper_camel .Name }} for new namespace */
						{
							{{ lower_camel .PluralName }}, err := c.{{ lower_camel .Name }}.List(namespace, clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
							if err != nil {
								errs <- errors.Wrapf(err, "initial new namespace {{ upper_camel .Name }} list in namespace watch")
								continue
							}
							{{ lower_camel .PluralName }}ByNamespace.Store(namespace, {{ lower_camel .PluralName }})
						}
						{{ lower_camel .Name }}NamespacesChan, {{ lower_camel .Name }}Errs, err := c.{{ lower_camel .Name }}.Watch(namespace, clients.WatchOpts{Ctx: opts.Ctx, Selector: opts.Selector})
						if err != nil {
							errs <- errors.Wrapf(err, "starting new namespace {{ upper_camel .Name }} watch")
							continue
						}

						done.Add(1)
						go func(namespace string) {
							defer done.Done()
							errutils.AggregateErrs(ctx, errs, {{ lower_camel .Name }}Errs, namespace+"-new-namespace-{{ lower_camel .PluralName }}")
						}(namespace)
{{- end }}
{{- end }}
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
					if len(newNamespaces) > 0 {
						contextutils.LoggerFrom(ctx).Infof("registered the new namespace %v", newNamespaces)
					}
					c.updateNamespaces.Unlock()
				}
			}
		}()
	}
{{- range .Resources}}
{{- if .ClusterScoped }}
	/* Setup cluster-wide watch for {{ .Name }} */
	var err error
	currentSnapshot.{{ upper_camel .PluralName }},err = c.{{ lower_camel .Name }}.List(clients.ListOpts{Ctx: opts.Ctx, Selector: opts.Selector})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "initial {{ .Name }} list")
	}
	// for Cluster scoped resources, we do not use Expression Selectors
	{{ lower_camel .Name }}Chan, {{ lower_camel .Name }}Errs, err := c.{{ lower_camel .Name }}.Watch(clients.WatchOpts{
		Ctx: opts.Ctx,
		Selector: opts.Selector,
	})
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
