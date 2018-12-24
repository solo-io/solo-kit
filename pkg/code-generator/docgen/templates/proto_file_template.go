package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/funcs"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func ProtoFileTemplate(project *model.Project) *template.Template {
	return template.Must(template.New("p").Funcs(funcs.TemplateFuncs(project)).Parse(`
{{ $File := . -}}

## Package: {{ .Package }}

## Source File: {{ .Name }} 

## Description:
{{- range .SyntaxComments.Detached }}  
    {{ remove_magic_comments (printf "%v" .) }}
{{- end }}  

## Contents:
{{ $msgLinkItem :=  "- [{{ printfptr \"%v\" .Name }}](#{{.Name}}) " }}

{{- forEachMessage .Messages $msgLinkItem }}  

{{ $msgLongInfo :=  `+"`"+`

`+"`"+` }}

{{- if gt (len .Enums) 0 }} 
- Enums:
{{- range .Enums}}
	- [{{ printfptr "%v" .Name }}](#{{.Name}})
{{- end}}
{{- end}}

---
{{range .Messages }}  
### <a name="{{ printfptr "%v" .Name }}">{{ printfptr "%v" .Name }}</a>

Description: {{ remove_magic_comments .Comments.Leading }}

` + "```" + `yaml
{{range .Fields -}}
"{{ printfptr "%v" .Name}}": {{ fieldType . }}
{{end}}
` + "```" + `

| Field | Type | Description | Default |
| ----- | ---- | ----------- |----------- | 
{{range .Fields -}}
| {{ printfptr "%v" .Name }} | {{linkForField $File . }} | {{ remove_magic_comments (nobr .Comments.Leading) }} | {{if .DefaultValue}} Default: {{.DefaultValue}}{{end}} |
{{end}}

{{- range .Enums }}  
### <a name="{{ printfptr "%v" .Name }}">{{ printfptr "%v" .Name }}</a>

Description: {{ remove_magic_comments .Comments.Leading }}

| Name | Description |
| ----- | ----------- | 
{{range .Values -}}
| {{ printfptr "%v" .Name }} | {{ remove_magic_comments (nobr .Comments.Leading) }} |
{{end}}

{{- end }}

<!-- Start of HubSpot Embed Code -->
<script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
<!-- End of HubSpot Embed Code -->
`))
}
