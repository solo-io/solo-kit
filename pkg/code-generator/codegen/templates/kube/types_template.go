package kube

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var ResourceTemplate = template.Must(template.New("kube_resource").Funcs(templates.Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}


import (
	api "{{ .Project.ProjectConfig.GoPackage }}"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
{{- if $.ClusterScoped }}
// +genclient:nonNamespaced
{{- else }}
// +genclient
{{- end }}
{{ if $.HasStatus -}}
// +genclient:noStatus
{{- end }}
type {{ .Name }} struct {
	v1.TypeMeta {{ backtick }}json:",inline"{{ backtick }}
	// +optional
	v1.ObjectMeta {{ backtick }}json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"{{ backtick }}

	// Spec defines the implementation of this definition.
	// +optional
	Spec api.{{ .Name }} {{ backtick }}json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"{{ backtick }}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// {{ .Name }}List is a collection of {{ .Name }}s.
type {{ .Name }}List struct {
	v1.TypeMeta {{ backtick }}json:",inline"{{ backtick }}
	// +optional
	v1.ListMeta {{ backtick }}json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"{{ backtick }}
	Items       []{{ .Name }} {{ backtick }}json:"items" protobuf:"bytes,2,rep,name=items"{{ backtick }}
}


/*



import (
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

		{{- if $.SkipHashingAnnotations }}
		meta.Annotations = nil
		{{- end }}
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
	metaCopy.Generation = 0
	// investigate zeroing out owner refs as well
	{{- if $.SkipHashingAnnotations }}
	metaCopy.Annotations = nil
	{{- end }}
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


func (r *{{ .Name }}) GroupVersionKind() schema.GroupVersionKind {
	return {{ .Name }}GVK
}

type {{ .Name }}List []*{{ .Name }}

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

func init() {
	if err := crd.AddCrd({{ .Name }}Crd); err != nil {
		log.Fatalf("could not add crd to global registry")
	}
}

{{- end}}

var (
	{{ .Name }}GVK = schema.GroupVersionKind{
		Version: "{{ .Project.ProjectConfig.Version }}",
		Group: "{{ $crdGroupName }}",
		Kind: "{{ .Name }}",
	}
)

*/
`))
