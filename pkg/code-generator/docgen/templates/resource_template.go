package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var ResourceTemplate = template.Must(template.New("resource").Funcs(templates.Funcs).Parse(`
{{ . }}
`))
