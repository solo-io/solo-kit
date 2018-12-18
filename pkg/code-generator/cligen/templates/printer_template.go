package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/templateutils"
)

var PrinterTemplate = template.Must(template.New("p").Funcs(templateutils.Funcs).Parse(`
package  {{lowercase .Resource.Name}}

import (
	"io"
	"os"
{{range .Imports}}	"{{lowercase .}}"
{{end}}
	"github.com/olekukonko/tablewriter"
	"github.com/solo-io/solo-kit/pkg/utils/cliutils"
)

func PrintTable(list *v1.{{.Resource.Name}}List, output string, template string) error {
	err := cliutils.PrintList(output, template, list,
		func(data interface{}, w io.Writer) error {
			{{lowercase .Resource.Name}}Table(data.(*v1.{{.Resource.Name}}List), w)
			return nil
		},
		os.Stdout)
	return err
}

func {{lowercase .Resource.Name}}Table(list *v1.{{.Resource.Name}}List, w io.Writer) {
	table := tablewriter.NewWriter(w)
	headers := []string{"", "name"}
	table.SetHeader(headers)
	table.SetBorder(false)
	//for i, v := range *list {
		//table.Append(transform(v, i+1))
	//}
	table.Render()
}

`))
