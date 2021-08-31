package templates

import (
	"text/template"
)

var ResourceTemplate = template.Must(template.New("resource").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"log"
	"os"
	"sort"

{{- if $.IsCustom }}
	{{ $.CustomImportPrefix }} "{{ $.CustomResource.Package }}"
{{- end }}

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/go-utils/hashutils"
{{- if not $.IsCustom }}
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"k8s.io/apimachinery/pkg/runtime"
{{- end }}
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func New{{ .Name }}(namespace, name string) *{{ .Name }} {
	{{ lowercase .Name }} := &{{ .Name }}{}
{{- if $.IsCustom }}
	{{ lowercase .Name }}.{{ $.Name }}.SetMetadata(&core.Metadata{
{{- else }}
	{{ lowercase .Name }}.SetMetadata(&core.Metadata{
{{- end }}
		Name:      name,
		Namespace: namespace,
	})
	return {{ lowercase .Name }}
}

{{- if $.IsCustom }}

// require custom resource to implement Clone() as well as resources.Resource interface

type Cloneable{{ $.Name }} interface {
	resources.Resource
	Clone() *{{ $.CustomImportPrefix}}.{{ $.Name }}
}

var _ Cloneable{{ $.Name }} = &{{ $.CustomImportPrefix}}.{{ $.Name }}{}

type {{ $.Name }} struct {
	{{ $.CustomImportPrefix}}.{{ $.Name }}
}

func (r *{{ .Name }}) Clone() resources.Resource {
	return &{{ .Name }}{ {{ .Name }}: *r.{{ .Name }}.Clone() }
}

func (r *{{ .Name }}) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	clone := r.{{ .Name }}.Clone()
	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
		{{- if $.SkipHashingAnnotations }}
		meta.Annotations = nil
		{{- end }}
	})
	err := binary.Write(hasher, binary.LittleEndian, hashutils.HashAll(clone))
	if err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

{{- else }}

func (r *{{ .Name }}) SetMetadata(meta *core.Metadata) {
	r.Metadata = meta
}

{{- if $.HasStatus }}

// Deprecated
func (r *{{ .Name }}) SetStatus(status *core.Status) {
	r.SetStatusForNamespace(status)
}

// Deprecated
func (r *{{ .Name }}) GetStatus() *core.Status {
	if r != nil {
		s, _ := r.GetStatusForNamespace()
		return s
	}
	return nil
}

func (r *{{ .Name }}) SetNamespacedStatuses(statuses *core.NamespacedStatuses) {
	r.NamespacedStatuses = statuses
}

// SetStatusForNamespace inserts the specified status into the NamespacedStatuses.Statuses map for
// the current namespace (as specified by POD_NAMESPACE env var).  If the resource does not yet
// have a NamespacedStatuses, one will be created.
// Note: POD_NAMESPACE environment variable must be set for this function to behave as expected.
// If unset, a podNamespaceErr is returned.
func (r *{{ .Name }}) SetStatusForNamespace(status *core.Status) error {
	podNamespace := os.Getenv(envutils.PodNamespaceEnvName)
	if podNamespace == "" {
		return errors.NewPodNamespaceErr()
	}
	if r.GetNamespacedStatuses() == nil {
		r.SetNamespacedStatuses(&core.NamespacedStatuses{})
	}
	if r.GetNamespacedStatuses().GetStatuses() == nil {
		r.GetNamespacedStatuses().Statuses = make(map[string]*core.Status)
	}
	r.GetNamespacedStatuses().GetStatuses()[podNamespace] = status
	return nil
}

// GetStatusForNamespace returns the status stored in the NamespacedStatuses.Statuses map for the
// controller specified by the POD_NAMESPACE env var, or nil if no status exists for that
// controller.
// Note: POD_NAMESPACE environment variable must be set for this function to behave as expected.
// If unset, a podNamespaceErr is returned.
func (r *{{ .Name }}) GetStatusForNamespace() (*core.Status, error) {
	podNamespace := os.Getenv(envutils.PodNamespaceEnvName)
	if podNamespace == "" {
		return nil, errors.NewPodNamespaceErr()
	}
	if r.GetNamespacedStatuses() == nil {
		return nil, nil
	}
	if r.GetNamespacedStatuses().GetStatuses() == nil {
		return nil, nil
	}
	return r.GetNamespacedStatuses().GetStatuses()[podNamespace], nil
}

{{- end }}

{{- end }}

func (r *{{ .Name }}) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *{{ .Name }}) GroupVersionKind() schema.GroupVersionKind {
	return {{ .Name }}GVK
}

type {{ .Name }}List []*{{ .Name }}

func (list {{ .Name }}List) Find(namespace, name string) (*{{ .Name }}, error) {
	for _, {{ lower_camel .Name }} := range list {
		if {{ lower_camel .Name }}.GetMetadata().Name == name && {{ lower_camel .Name }}.GetMetadata().Namespace == namespace {
			return {{ lower_camel .Name }}, nil
		}
	}
	return nil, errors.Errorf("list did not find {{ lower_camel .Name }} %v.%v", namespace, name)
}

func (list {{ .Name }}List) AsResources() resources.ResourceList {
	var ress resources.ResourceList 
	for _, {{ lower_camel .Name }} := range list {
		ress = append(ress, {{ lower_camel .Name }})
	}
	return ress
}

{{ if $.HasStatus -}}
func (list {{ .Name }}List) AsInputResources() resources.InputResourceList {
	var ress resources.InputResourceList
	for _, {{ lower_camel .Name }} := range list {
		ress = append(ress, {{ lower_camel .Name }})
	}
	return ress
}
{{- end}}

func (list {{ .Name }}List) Names() []string {
	var names []string
	for _, {{ lower_camel .Name }} := range list {
		names = append(names, {{ lower_camel .Name }}.GetMetadata().Name)
	}
	return names
}

func (list {{ .Name }}List) NamespacesDotNames() []string {
	var names []string
	for _, {{ lower_camel .Name }} := range list {
		names = append(names, {{ lower_camel .Name }}.GetMetadata().Namespace + "." + {{ lower_camel .Name }}.GetMetadata().Name)
	}
	return names
}

func (list {{ .Name }}List) Sort() {{ .Name }}List {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list {{ .Name }}List) Clone() {{ .Name }}List {
	var {{ lower_camel .Name }}List {{ .Name }}List
	for _, {{ lower_camel .Name }} := range list {
		{{ lower_camel .Name }}List = append({{ lower_camel .Name }}List, resources.Clone({{ lower_camel .Name }}).(*{{ .Name }}))
	}
	return {{ lower_camel .Name }}List 
}

func (list {{ .Name }}List) Each(f func(element *{{ .Name }})) {
	for _, {{ lower_camel .Name }} := range list {
		f({{ lower_camel .Name }})
	}
}

func (list {{ .Name }}List) EachResource(f func(element resources.Resource)) {
	for _, {{ lower_camel .Name }} := range list {
		f({{ lower_camel .Name }})
	}
}

func (list {{ .Name }}List) AsInterfaces() []interface{}{
	var asInterfaces []interface{}
	list.Each(func(element *{{ .Name }}) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

{{- $crdGroupName := .Project.ProtoPackage }}
{{- if ne .Project.ProjectConfig.CrdGroupOverride "" }}
{{- $crdGroupName = .Project.ProjectConfig.CrdGroupOverride }}
{{- end}}

{{- if not $.IsCustom }}

// Kubernetes Adapter for {{ .Name }}

func (o *{{ .Name }}) GetObjectKind() schema.ObjectKind {
	t := {{ .Name }}Crd.TypeMeta()
	return &t
}

func (o *{{ .Name }}) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*{{ .Name }})
}

func (o *{{ .Name }}) DeepCopyInto(out *{{ .Name }}) {
	clone := resources.Clone(o).(*{{ .Name }})
	*out = *clone
}

var (
	{{ .Name }}Crd = crd.NewCrd(
		"{{ lowercase (upper_camel .PluralName) }}",
		{{ .Name }}GVK.Group,
		{{ .Name }}GVK.Version,
		{{ .Name }}GVK.Kind,
		"{{ .ShortName }}",
		{{ .ClusterScoped }},
		&{{ .Name }}{})
)

{{- end}}

var (
	{{ .Name }}GVK = schema.GroupVersionKind{
		Version: "{{ .Project.ProjectConfig.Version }}",
		Group: "{{ $crdGroupName }}",
		Kind: "{{ .Name }}",
	}
)
`))
