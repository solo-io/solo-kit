package templates

import (
	"text/template"
)

var ResourceTemplate = template.Must(template.New("resource").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}

import (
	"sort"

{{- if $.IsCustom }}
	{{ $.CustomImportPrefix }} "{{ $.CustomResource.Package }}"
{{- end }}

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/hashutils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func New{{ .Name }}(namespace, name string) *{{ .Name }} {
	{{ lowercase .Name }} := &{{ .Name }}{}
{{- if $.IsCustom }}
	{{ lowercase .Name }}.{{ $.Name }}.SetMetadata(core.Metadata{
{{- else }}
	{{ lowercase .Name }}.SetMetadata(core.Metadata{
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

func (r *{{ .Name }}) Hash() uint64 {
	clone := r.{{ .Name }}.Clone()

	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})

	return hashutils.HashAll(clone)
}

{{- else }}

func (r *{{ .Name }}) SetMetadata(meta core.Metadata) {
	r.Metadata = meta
}

{{- if $.HasStatus }}

func (r *{{ .Name }}) SetStatus(status core.Status) {
	r.Status = status
}
{{- end }}

func (r *{{ .Name }}) Hash() uint64 {
	metaCopy := r.GetMetadata()
	metaCopy.ResourceVersion = ""
	return hashutils.HashAll(
		metaCopy,
{{- range .Fields }}
	{{- if not ( or (eq .Name "metadata") (eq .Name "status") .IsOneof .SkipHashing ) }}
		r.{{ upper_camel .Name }},
	{{- end }}
{{- end}}
{{- range .Oneofs }}
		r.{{ upper_camel .Name }},
{{- end}}
	)
}

{{- end }}

type {{ .Name }}List []*{{ .Name }}
type {{ upper_camel .PluralName }}ByNamespace map[string]{{ .Name }}List

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list {{ .Name }}List) Find(namespace, name string) (*{{ .Name }}, error) {
	for _, {{ lower_camel .Name }} := range list {
		if {{ lower_camel .Name }}.GetMetadata().Name == name {
			if namespace == "" || {{ lower_camel .Name }}.GetMetadata().Namespace == namespace {
				return {{ lower_camel .Name }}, nil
			}
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

func (list {{ .Name }}List) AsInterfaces() []interface{}{
	var asInterfaces []interface{}
	list.Each(func(element *{{ .Name }}) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

func (byNamespace {{ upper_camel .PluralName }}ByNamespace) Add({{ lower_camel .Name }} ... *{{ .Name }}) {
	for _, item := range {{ lower_camel .Name }} {
		byNamespace[item.GetMetadata().Namespace] = append(byNamespace[item.GetMetadata().Namespace], item)
	}
}

func (byNamespace {{ upper_camel .PluralName }}ByNamespace) Clear(namespace string) {
	delete(byNamespace, namespace)
}

func (byNamespace {{ upper_camel .PluralName }}ByNamespace) List() {{ .Name }}List {
	var list {{ .Name }}List
	for _, {{ lower_camel .Name }}List := range byNamespace {
		list = append(list, {{ lower_camel .Name }}List...)
	}
	return list.Sort()
}

func (byNamespace {{ upper_camel .PluralName }}ByNamespace) Clone() {{ upper_camel .PluralName }}ByNamespace {
	cloned := make({{ upper_camel .PluralName }}ByNamespace)
	for ns, list := range byNamespace {
		cloned[ns] = list.Clone()
	}
	return cloned
}

{{- if not $.IsCustom }}

var _ resources.Resource = &{{ .Name }}{}

// Kubernetes Adapter for {{ .Name }}

func (o *{{ .Name }}) GetObjectKind() schema.ObjectKind {
	t := {{ .Name }}Crd.TypeMeta()
	return &t
}

func (o *{{ .Name }}) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*{{ .Name }})
}

{{- $crdGroupName := .Project.ProtoPackage }}
{{- if ne .Project.ProjectConfig.CrdGroupOverride "" }}
{{- $crdGroupName = .Project.ProjectConfig.CrdGroupOverride }}
{{- end}}

var {{ .Name }}Crd = crd.NewCrd("{{ $crdGroupName }}",
	"{{ lowercase (upper_camel .PluralName) }}",
	"{{ $crdGroupName }}",
	"{{ .Project.ProjectConfig.Version }}",
	"{{ .Name }}",
	"{{ .ShortName }}",
	{{ .ClusterScoped }},
	&{{ .Name }}{})

{{- end}}
`))
