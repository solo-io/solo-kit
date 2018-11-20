package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var ProjectTemplate = template.Must(template.New("p").Funcs(templates.Funcs).Parse(`
{{ . }}
`))
