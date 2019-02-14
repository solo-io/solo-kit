package restructured

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/funcs"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func ProjectDocsRootTemplate(project *model.Project, docsOptions *options.DocsOptions) *template.Template {
	return template.Must(template.New("pf").Funcs(funcs.TemplateFuncs(project, docsOptions)).Parse(`
===========================================
API Reference for {{ .ProjectConfig.Title}}
===========================================
.. _{{ .ProjectConfig.Title}}:

API Version: ` + "`{{ .ProjectConfig.Name }}.{{ .ProjectConfig.Version }}`" + `

{{ .ProjectConfig.Description }}

API Resources:
~~~~~~~~~~~~~~

.. toctree::
   :maxdepth: 2

{{ range .Resources}}
{{- if (not .SkipDocsGen) }}
   .Name <{{ .Filename }}.sk>
{{- end}}
{{- end}}

.. raw:: html
   <!-- Start of HubSpot Embed Code -->
   <script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
   <!-- End of HubSpot Embed Code -->
`))
}
