package templates

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/funcs"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"text/template"
)

func ProjectTemplate(project *model.Project) *template.Template {
	return template.Must(template.New("p").Funcs(funcs.TemplateFuncs(project)).Parse(`
### {{ .Name }} {{.Version}} Top Level API Objects:
{{- range .Resources}}
- [{{ .ImportPrefix }}{{ .Name }}](./{{ .Filename }}.sk.md#{{ .Name }})
{{- end}}

<!-- Start of HubSpot Embed Code -->
<script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
<!-- End of HubSpot Embed Code -->
`))
}
