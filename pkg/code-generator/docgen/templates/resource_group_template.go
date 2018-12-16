package templates

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/templateutils"
	"text/template"
)

var ResourceGroupTemplate = template.Must(template.New("rg").Funcs(templateutils.Funcs).Parse(`
{{ . }}

<!-- Start of HubSpot Embed Code -->
<script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
<!-- End of HubSpot Embed Code -->
`))
