package templates

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/templateutils"
	"text/template"
)

var MainTemplate = template.Must(template.New("p").Funcs(templateutils.Funcs).Parse(`
package main

{{if .Imports }}
import (
{{range .Imports}}	"{{.}}"
{{end}})
{{end}}


func main() {
	opts := &options.Options{}
	root := cmd.RootCmd(opts)
	root.Execute()
}

`))