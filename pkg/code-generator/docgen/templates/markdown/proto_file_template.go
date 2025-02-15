package markdown

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/funcs"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func ProtoFileTemplate(project *model.Project, docsOptions *options.DocsOptions) *template.Template {
	str := `
{{ $File := . -}}

### Package: {{ backtick }}{{ .Package }}{{ backtick }}

{{- if gt (len .SyntaxComments.Detached) 0 }}

{{- range .SyntaxComments.Detached }}  
{{ remove_magic_comments (printf "%v" .) }}

{{- end }}


{{ end }}

{{- if gt (len .Messages) 0 }} 
**Types:**

{{ $linkMessage :=  "- [{{ printfptr \"%v\" .Name }}](#{{ toAnchorLink \"%v\" .Name }}) {{- if (resourceForMessage .) }} **Top-Level Resource**{{ end }}" }}
{{ $linkEnum :=  "- [{{ printfptr \"%v\" .Name }}](#{{ toAnchorLink \"%v\" .Name }})" }}
{{- forEachMessage $File .Messages $linkMessage $linkEnum }}  

{{ end}}

{{- if gt (len .Enums) 0 }} 

**Enums:**

{{ range .Enums}}
	- [{{ printfptr "%v" .Name }}](#{{ toAnchorLink "%v" .Name }})

{{- end}}

{{ end}}

**Source File: {{ githubLinkForFile "main" .Name }}**

{{ $msgLongInfo :=  ` + "`" + `
{{ $Message := . -}}
---
### {{ toHeading "%v" .Name }}

{{ if gt (len .Comments.Leading) 0 }} 
{{ remove_magic_comments .Comments.Leading }}
{{- end }}

{{backtick}}{{backtick}}{{backtick}}yaml
{{range .Fields -}}
"{{ lower_camel (printfptr "%v" .Name) }}": {{ fieldType . }}
{{end}}
{{backtick}}{{backtick}}{{backtick}}

| Field | Type | Description |
| ----- | ---- | ----------- | 
{{range .Fields -}}
{{- $description := remove_magic_comments (nobr .Comments.Leading) -}}
{{- $oneofmsg := getOneofMessage . -}}
| {{backtick}}{{ lower_camel (printfptr "%v" .Name) }}{{backtick}} | {{linkForField (getFileForMessage $Message) . }} | {{ if $description }}{{trimSuffix "." $description}}.{{ end }}{{ if $oneofmsg }} {{$oneofmsg}}{{ end }} |
{{end}}

` + "`" + ` }}


{{ $enumLongInfo :=  ` + "`" +
		`
{{ $Enum := . -}}
---
### {{ toHeading "%v" .Name }}

{{ if gt (len .Comments.Leading) 0 }} 
{{ remove_magic_comments .Comments.Leading }}
{{- end }}

| Name | Description |
| ----- | ----------- | 
{{range .Values -}}
| {{backtick}}{{ printfptr "%v" .Name }}{{backtick}} | {{ remove_magic_comments (nobr .Comments.Leading) }} |
{{end}}

` + "`" + ` }}

{{- forEachMessage $File .Messages $msgLongInfo $enumLongInfo }}  

{{- range .Enums }}  
### {{ toHeading "%v" .Name }}

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
`
	return template.Must(template.New("p").Funcs(funcs.TemplateFuncs(project, docsOptions)).Parse(str))
}
