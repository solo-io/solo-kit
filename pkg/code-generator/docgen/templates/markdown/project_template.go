package markdown

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/funcs"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func ProjectDocsRootTemplate(project *model.Project, docsOptions *options.DocsOptions) *template.Template {
	frontMatter := `
---
title: "{{ .ProjectConfig.Name }}"
weight: 5
---
`
	str := `

### API Reference for {{ .ProjectConfig.Title}}

API Version: ` + "`{{ .ProjectConfig.Name }}.{{ .ProjectConfig.Version }}`" + `

{{ .ProjectConfig.Description }}

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
	fullTemplate := str
	if docsOptions.Output == options.Hugo {
		fullTemplate = frontMatter + str
	}
	return template.Must(template.New("pf").Funcs(funcs.TemplateFuncs(project, docsOptions)).Parse(fullTemplate))
}
