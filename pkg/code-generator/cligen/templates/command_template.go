package templates

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/templateutils"
	"text/template"
)

var CommandTemplate = template.Must(template.New("p").Funcs(templateutils.Funcs).Parse(`

import (
{{.Imports}}
	"github.com/spf13/cobra"
)

func {{.Name}}Cmd(opts *options.Options) *cobra.Command {
	//opts := &options.Options{}
	cmd := &cobra.Command{
		Use:     "{{.Use}}",
		Short:   "{{.Short}}",
		Long: 	 "{{.Long}}"
        Run: func(cmd *cobra.Command, args []string) {
			// TODO: Add application logic here
			return nil
		},
	}

	cmd.AddCommand(
	{{range $key, $value := .Resources}} 
		{{$value.Name}}.Cmd(opts),
	{{end}}
	)
	return cmd
}

`))