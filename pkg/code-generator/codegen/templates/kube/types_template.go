package kube

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var TypesTemplate = template.Must(template.New("kube_types").Funcs(templates.Funcs).Parse(`package {{ .ProjectConfig.Version }}


import (
	"encoding/json"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"

	api "{{ .ProjectConfig.GoPackage }}"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type metaOnly struct {
	v1.TypeMeta   {{ backtick }}json:",inline"{{ backtick }}
	v1.ObjectMeta {{ backtick }}json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"{{ backtick }}
}

{{- range .Resources}}
{{- if resourceBelongsToProject $ . }}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resourceName={{ lowercase (upper_camel .PluralName) }}
{{- if .ClusterScoped }}
// +genclient:nonNamespaced
{{- else }}
// +genclient
{{- end }}
{{- if not .HasStatus }}
// +genclient:noStatus
{{- end }}
type {{ .Name }} struct {
	v1.TypeMeta {{ backtick }}json:",inline"{{ backtick }}
	// +optional
	v1.ObjectMeta {{ backtick }}json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"{{ backtick }}

	// Spec defines the implementation of this definition.
	// +optional
	Spec api.{{ .Name }} {{ backtick }}json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"{{ backtick }}

{{- if .HasStatus }}
	Status core.NamespacedStatuses {{ backtick }}json:"status,omitempty" protobuf:"bytes,3,opt,name=status"{{ backtick }}
{{- end }}
}


func (o *{{ .Name }}) MarshalJSON() ([]byte, error) {
	spec, err := protoutils.MarshalMap(&o.Spec)
	if err != nil {
		return nil, err
	}
	delete(spec, "metadata")
{{- if .HasStatus }}
	delete(spec, "status")
{{- end }}
	asMap := map[string]interface{}{
		"metadata":   o.ObjectMeta,
		"apiVersion": o.TypeMeta.APIVersion,
		"kind":       o.TypeMeta.Kind,
{{- if .HasStatus }}
		"status": o.NamespacedStatuses,
{{- end }}
		"spec":       spec,
	}
	return json.Marshal(asMap)
}

func (o *{{ .Name }}) UnmarshalJSON(data []byte) error {
	var metaOnly metaOnly
	if err := json.Unmarshal(data, &metaOnly); err != nil {
		return err
	}
	var spec api.{{ .Name }}
	if err := protoutils.UnmarshalResource(data, &spec); err != nil {
		return err
	}
	spec.Metadata = nil
	*o = {{ .Name }}{
		ObjectMeta: metaOnly.ObjectMeta,
		TypeMeta:   metaOnly.TypeMeta,
		Spec:       spec,
	}
{{- if .HasStatus }}
	if spec.NamespacedStatuses != nil {
		o.Status = *spec.NamespacedStatuses
		o.Spec.NamespacedStatuses = nil
	}
{{- end }}

	return nil
}


// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// {{ .Name }}List is a collection of {{ .Name }}s.
type {{ .Name }}List struct {
	v1.TypeMeta {{ backtick }}json:",inline"{{ backtick }}
	// +optional
	v1.ListMeta {{ backtick }}json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"{{ backtick }}
	Items       []{{ .Name }} {{ backtick }}json:"items" protobuf:"bytes,2,rep,name=items"{{ backtick }}
}

{{- end }}
{{- end }}
`))
