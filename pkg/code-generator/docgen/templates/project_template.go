package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var ProjectTemplate = template.Must(template.New("p").Funcs(templates.Funcs).Parse(`
### {{ .Name }} {{.Version}} Top Level API Objects:
{{- range .Resources}}
- [{{ .ImportPrefix }}{{ .Name }}](./{{ .Filename }}.sk.md#{{ .Name }})
{{- end}}

`))
