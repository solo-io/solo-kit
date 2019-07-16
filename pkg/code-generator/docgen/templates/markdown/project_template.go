package markdown

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/funcs"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func ProjectDocsRootTemplate(project *model.Version, docsOptions *options.DocsOptions) *template.Template {
	str := `

### API Reference for {{ .VersionConfig.ApiGroup.SoloKitProject.Title}}

API Version: ` + "`{{ .VersionConfig.ApiGroup.Name }}.{{ .VersionConfig.Version }}`" + `

{{ .VersionConfig.ApiGroup.SoloKitProject.Description }}

### API Resources:
{{- range .Resources}}
{{- if (not .SkipDocsGen) }}
- {{linkForResource . }}
{{- end}}
{{- end}}

<!-- Start of HubSpot Embed Code -->
<script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
<!-- End of HubSpot Embed Code -->
`
	return template.Must(template.New("markdown_project").Funcs(funcs.TemplateFuncs(project, docsOptions)).Parse(str))
}
