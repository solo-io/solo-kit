package kube

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var DocTemplate = template.Must(template.New("kube_doc").Funcs(templates.Funcs).Parse(`
// +k8s:deepcopy-gen=package,register

//go:generate $GOPATH/src/k8s.io/code-generator/generate-groups.sh all "{{ .ProjectConfig.GoPackage }}/kube/client" "{{ .ProjectConfig.GoPackage }}/kube/apis" {{ .ProjectConfig.Name }}:{{ .ProjectConfig.Version }}

// Package {{ .ProjectConfig.Version }} is the {{ .ProjectConfig.Version }} version of the API.
// +groupName={{ .ProjectConfig.Name }}
package {{ .ProjectConfig.Version }}

`))
