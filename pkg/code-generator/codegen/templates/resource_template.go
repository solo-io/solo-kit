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
{{- if $.HasStatus }}
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"
{{- end }}
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// Compile-time assertion
{{- if $.HasStatus -}}
{{- if $.IsCustom }}
	_ resources.CustomInputResource = new({{ .Name }})
{{- else }}
	_ resources.InputResource = new({{ .Name }})
{{- end }}
{{- else }}
	_ resources.Resource = new({{ .Name }})
{{- end }}
)

func New{{ .Name }}HashableResource() resources.HashableResource {
	return new({{ .Name }})
}

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

	{{- if $.SpecHasHash }}

	_, err := hasher.Write([]byte(r.{{ .Name }}.Namespace))
	if err != nil {
		return 0, err
	}
	_, err = hasher.Write([]byte(r.{{ .Name }}.Name))
	if err != nil {
		return 0, err
	}
	_, err = hasher.Write([]byte(r.{{ .Name }}.UID))
	if err != nil {
		return 0, err
	}

	{
		var result uint64
		innerHash := fnv.New64()
		for k, v := range r.Labels {
			innerHash.Reset()

			if _, err = innerHash.Write([]byte(v)); err != nil {
				return 0, err
			}

			if _, err = innerHash.Write([]byte(k)); err != nil {
				return 0, err
			}

			result = result ^ innerHash.Sum64()
		}
		err = binary.Write(hasher, binary.LittleEndian, result)
		if err != nil {
			return 0, err
		}
	}
	{{- if not $.SkipHashingAnnotations }}
	{
		var result uint64
		innerHash := fnv.New64()
		for k, v := range r.Annotations {
			innerHash.Reset()

			if _, err = innerHash.Write([]byte(v)); err != nil {
				return 0, err
			}

			if _, err = innerHash.Write([]byte(k)); err != nil {
				return 0, err
			}

			result = result ^ innerHash.Sum64()
		}
		err = binary.Write(hasher, binary.LittleEndian, result)
		if err != nil {
			return 0, err
		}
	}
	{{- end }}
	
	_, err = r.{{ .Name }}.Spec.Hash(hasher)
	if err != nil {
		return 0, err
	}

	{{- else }}
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


	{{- end }}
	return hasher.Sum64(), nil
}

{{- else }}

func (r *{{ .Name }}) SetMetadata(meta *core.Metadata) {
	r.Metadata = meta
}

{{- if $.HasStatus }}

// Deprecated
func (r *{{ .Name }}) SetStatus(status *core.Status) {
	statusutils.SetSingleStatusInNamespacedStatuses(r, status)
}

// Deprecated
func (r *{{ .Name }}) GetStatus() *core.Status {
	if r != nil {
		return statusutils.GetSingleStatusInNamespacedStatuses(r)
	}
	return nil
}

func (r *{{ .Name }}) SetNamespacedStatuses(namespacedStatuses *core.NamespacedStatuses) {
	r.NamespacedStatuses = namespacedStatuses
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
