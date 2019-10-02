package kube

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var DocTemplate = template.Must(template.New("kube_doc").Funcs(templates.Funcs).Parse(`
// +k8s:deepcopy-gen=package,register

// Package {{ .ProjectConfig.Version }} is the {{ .ProjectConfig.Version }} version of the API.
// +groupName={{ .ProjectConfig.Name }}
package {{ .ProjectConfig.Version }}

`))
