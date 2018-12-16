package templates

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/templateutils"
	"text/template"
)

var CommandTemplate = template.Must(template.New("p").Funcs(templateutils.Funcs).Parse(`
package {{if .CliFile.PackageName}}{{.CliFile.PackageName}}{{else}}{{lowercase .Resource.Name}}{{end}}

import (
{{range .CliFile.Imports}}	"{{lowercase .}}"
{{end}}
	"github.com/spf13/cobra"
)

func {{if .IsRoot}}Root{{else}}{{.CliFile.PackageName}}{{end}}Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "{{.Cmd.Use}}",
		Short:   "{{.Cmd.Short}}",
		Long: 	 "{{.Cmd.Long}}",
        RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Add application logic here
			return nil
		},
	}
{{if .IsRoot}}
	cmd.AddCommand(
		{{range .Resources}}{{lowercase .Name}}.Cmd(opts),
		{{end}}
	)
{{end}}
	return cmd
}

`))



