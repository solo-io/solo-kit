package restructured

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/funcs"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func ProtoFileTemplate(project *model.Version, docsOptions *options.DocsOptions) *template.Template {
	str := `
{{ $File := . -}}

===================================================
Package: {{ backtick }}{{ .Package }}{{ backtick }}
===================================================

{{- if gt (len .SyntaxComments.Detached) 0 }}

{{- range .SyntaxComments.Detached }}  
{{ remove_magic_comments (printf "%v" .) }}

{{- end }}


{{ end }}

{{- if gt (len .Messages) 0 }}

.. _{{ .Package }}.{{ printfptr "%v" .Name }}:


**Types:**

{{ $linkMessage :=  "- :ref:{{backtick}}message.{{ .FullName }}{{backtick}}{{- if (resourceForMessage .) }} **Top-Level Resource**{{ end }}" }}
{{ $linkEnum :=  "- [{{ printfptr \"%v\" .Name }}](#{{ printfptr \"%v\" .Name }})" }}
{{- forEachMessage $File .Messages $linkMessage $linkEnum }}  

{{ end}}

{{- if gt (len .Enums) 0 }} 

**Enums:**

{{ range .Enums}}
	- [{{ printfptr "%v" .Name }}](#{{ printfptr "%v" .Name }})

{{- end}}

{{ end}}

**Source File:** {{ githubLinkForFile "master" .Name }}

{{ $msgLongInfo :=  ` + "`" + `
{{ $Message := . -}}

.. _message.{{ .FullName }}:

{{ printfptr "%v" .Name }}
~~~~~~~~~~~~~~~~~~~~~~~~~~

{{ if gt (len .Comments.Leading) 0 }} 
{{ remove_magic_comments .Comments.Leading }}
{{ end }}

::

{{range .Fields }}
   "{{ printfptr "%v" .Name}}": {{ fieldType . }}
{{- end}}

{{range .Fields }}

.. _field.{{ .FullName }}:

{{ printfptr "%v" .Name }}
++++++++++++++++++++++++++

Type: {{linkForField (getFileForMessage $Message) . }} 

Description: {{ remove_magic_comments (nobr .Comments.Leading) }} 

{{if .DefaultValue}}Default: {{.DefaultValue}}{{end}}
{{- end}}


` + "`" + ` }}

{{ $enumLongInfo :=  ` + "`" +
		`
{{ $Enum := . -}}
---
### <a name="{{ printfptr "%v" .Name }}">{{ printfptr "%v" .Name }}</a>

{{ if gt (len .Comments.Leading) 0 }} 
{{ remove_magic_comments .Comments.Leading }}
{{- end }}

.. csv-table:: Enum Reference
   :header: "Name", "Description"
   :delim: |

{{range .Values }}
   {{backtick}}{{ printfptr "%v" .Name }}{{backtick}} | {{ remove_magic_comments (nobr .Comments.Leading) }}
{{end}}

` + "`" + ` }}

{{- forEachMessage $File .Messages $msgLongInfo $enumLongInfo }}  

{{- range .Enums }}
### <a name="{{ printfptr "%v" .Name }}">{{ printfptr "%v" .Name }}</a>

Description: {{ remove_magic_comments .Comments.Leading }}

.. csv-table:: Fields Reference
   :header: "Name", "Description"
   :delim: |

{{range .Values }}
   {{ printfptr "%v" .Name }} | {{ remove_magic_comments (nobr .Comments.Leading) }}
{{end}}


{{- end }}

.. raw:: html
   <!-- Start of HubSpot Embed Code -->
   <script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
   <!-- End of HubSpot Embed Code -->
`
	return template.Must(template.New("p").Funcs(funcs.TemplateFuncs(project, docsOptions)).Parse(str))
}
