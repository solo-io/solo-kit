package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/templateutils"
)

var ResourceGroupTemplate = template.Must(template.New("rg").Funcs(templateutils.Funcs).Parse(`
{{ . }}

<!-- Start of HubSpot Embed Code -->
<script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
<!-- End of HubSpot Embed Code -->
`))
