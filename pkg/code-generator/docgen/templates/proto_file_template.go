package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

var ProtoFileTemplate = template.Must(template.New("resource").Funcs(templates.Funcs).Parse(`
## Package:
{{ .Package }}

## Source File:
{{ .Name }} 

## Description:
{{- range .SyntaxComments.Detached }}  
{{ printf "%v" . }}
{{- end }}  

## Contents:
- Messages:
{{- range .Messages }}  
	- [{{ printfptr "%v" .Name }}](#{{.Name}})
{{- end }}

{{- if gt (len .Enums) 0 }} 
- Enums:
{{- range .Enums}}
	- [{{ printfptr "%v" .Name }}](#{{.Name}})
{{- end}}
{{- end}}

---
{{range .Messages }}  
### <a name="{{ printfptr "%v" .Name }}">{{ printfptr "%v" .Name }}</a>

Description: {{ .Comments.Leading }}

`+"```"+`yaml
{{range .Fields -}}
"{{ printfptr "%v" .Name}}": {{ fieldType . }}
{{end}}
`+"```"+`

{{- end }}

`))
