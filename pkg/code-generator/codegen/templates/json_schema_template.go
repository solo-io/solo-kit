package templates

import (
	"text/template"
)

var JsonSchemaTemplate = template.Must(template.New("json_schema").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}

var {{ .Name }}JsonSchema = ` + "`" + `
{{ .JsonSchema }}
` + "`" + `
`))
