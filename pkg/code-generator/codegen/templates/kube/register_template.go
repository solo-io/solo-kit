package kube

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var RegisterTemplate = template.Must(template.New("kube_doc").Funcs(templates.Funcs).Parse(`package {{ .ProjectConfig.Version }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// Package-wide variables from generator "register".
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "{{ .ProjectConfig.Version }}"}
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

const (
	// Package-wide consts from generator "register".
	GroupName = "{{ .ProjectConfig.Name }}"
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

{{ $project := . }}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
{{- range .Resources }}
{{- if resourceBelongsToProject $project . }}
		&{{ .Name }}{},
		&{{ .Name }}List{},
{{- end }}
{{- end }}
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

`))
