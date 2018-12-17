package templates

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/templateutils"
	"text/template"
)

var ClientHelperTemplate = template.Must(template.New("p").Funcs(templateutils.Funcs).Parse(`
package {{.PackageName}}

import (
{{range .Imports}}	"{{lowercase .}}"
{{end}}
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

{{range .Resources}}func Must{{.Name}}Client() v1.{{.Name}}Client {
	client, err := {{.Name}}Client()
	if err != nil {
		log.Fatalf("failed to create {{lowercase .Name}} client: %v", err)
	}
	return client
}

func {{.Name}}Client() (v1.{{.Name}}Client, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	cache := kube.NewKubeCache()
	{{lowercase .Name}}Client, err := v1.New{{.Name}}Client(&factory.KubeResourceClientFactory{
		Crd:         v1.{{.Name}}Crd,
		Cfg:         cfg,
		SharedCache: cache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating {{lowercase .Name}}s client")
	}
	if err := {{lowercase .Name}}Client.Register(); err != nil {
		return nil, err
	}
	return {{lowercase .Name}}Client, nil
}

{{end}}

`))
