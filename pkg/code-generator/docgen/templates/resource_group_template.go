package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var ResourceGroupTemplate = template.Must(template.New("rg").Funcs(templates.Funcs).Parse(`
{{ . }}
`))
