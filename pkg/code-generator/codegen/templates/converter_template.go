package templates

import (
	"text/template"
)

var ConverterTemplate = template.Must(template.New("converter").Funcs(Funcs).Parse(`package {{ .ConversionGoPackageShort }}

import (
	"errors"

	"github.com/solo-io/go-utils/versionutils/kubeapi"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"

	{{- range .Conversions }}
	{{- range .Projects }}
	{{ .Version }} "{{ .GoPackage }}"
	{{- end }}
	{{- end }}
)

{{- range .Conversions }}
{{ $resourceName := .Name }}

type {{ upper_camel $resourceName }}UpConverter interface {
	{{- range .Projects }}
	{{- if .NextVersion }}
	From{{ upper_camel .Version }}To{{ upper_camel .NextVersion }}(src *{{ .Version }}.{{ upper_camel $resourceName }}) *{{ .NextVersion }}.{{ upper_camel $resourceName }}
	{{- end }}
	{{- end }}
}

type {{ upper_camel $resourceName }}DownConverter interface {
	{{- range .Projects }}
	{{- if .PreviousVersion }}
	From{{ upper_camel .Version }}To{{ upper_camel .PreviousVersion }}(src *{{ .Version }}.{{ upper_camel $resourceName }}) *{{ .PreviousVersion }}.{{ upper_camel $resourceName }}
	{{- end }}
	{{- end }}
}

type {{ lower_camel $resourceName }}Converter struct {
	upConverter   {{ upper_camel $resourceName }}UpConverter
	downConverter {{ upper_camel $resourceName }}DownConverter
}

func New{{ upper_camel $resourceName }}Converter(u {{ upper_camel $resourceName }}UpConverter, d {{ upper_camel $resourceName }}DownConverter) crd.Converter {
	return &{{ lower_camel $resourceName }}Converter{
		upConverter:   u,
		downConverter: d,
	}
}

func (c *{{ lower_camel $resourceName }}Converter) Convert(src, dst crd.SoloKitCrd) error {
	srcVersion, err := kubeapi.ParseVersion(src.GetObjectKind().GroupVersionKind().Version)
	if err != nil {
		return err
	}
	dstVersion, err := kubeapi.ParseVersion(dst.GetObjectKind().GroupVersionKind().Version)
	if err != nil {
		return err
	}

	if srcVersion.GreaterThan(dstVersion) {
		return c.convertDown(src, dst)
	} else if srcVersion.LessThan(dstVersion) {
		return c.convertUp(src, dst)
	}
	return crd.Copy(src, dst)
}

func (c *{{ lower_camel $resourceName }}Converter) convertUp(src, dst crd.SoloKitCrd) error {
	if src.GetObjectKind().GroupVersionKind().Version == dst.GetObjectKind().GroupVersionKind().Version {
		return crd.Copy(src, dst)
	}

	switch t := src.(type) {
	{{- range .Projects }}
	{{- if .NextVersion }}
	case *{{ .Version }}.{{ upper_camel $resourceName }}:
		return c.convertUp(c.upConverter.From{{ upper_camel .Version }}To{{ upper_camel .NextVersion }}(t), dst)
	{{- end }}
	{{- end }}
	}
	return errors.New("unrecognized source type, this should never happen")
}

func (c *{{ lower_camel $resourceName }}Converter) convertDown(src, dst crd.SoloKitCrd) error {
	if src.GetObjectKind().GroupVersionKind().Version == dst.GetObjectKind().GroupVersionKind().Version {
		return crd.Copy(src, dst)
	}

	switch t := src.(type) {
	{{- range .Projects }}
	{{- if .PreviousVersion }}
	case *{{ .Version }}.{{ upper_camel $resourceName }}:
		return c.convertDown(c.downConverter.From{{ upper_camel .Version }}To{{ upper_camel .PreviousVersion }}(t), dst)
	{{- end }}
	{{- end }}
	}
	return errors.New("unrecognized source type, this should never happen")
}

{{- end }}
`))
