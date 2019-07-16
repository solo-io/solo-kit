package templates

import (
	"text/template"
)

var ConverterTestTemplate = template.Must(template.New("converter_test").Funcs(Funcs).Parse(`package {{ .ConversionGoPackageShort }}_test

{{ $short_package := .ConversionGoPackageShort }}

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"{{ .ConversionGoPackage }}"

	{{- range .Conversions }}
	{{- range .Projects }}
	{{ .Version }} "{{ .GoPackage }}"
	{{- end }}
	{{- end }}
)

var converter crd.Converter

{{- range .Conversions }}
{{ $resource_name := .Name }}

var _ = Describe("{{ upper_camel $resource_name }}Converter", func() {
	BeforeEach(func() {
		converter = {{ $short_package }}.New{{ upper_camel $resource_name }}Converter({{ lower_camel $resource_name }}UpConverter{}, {{ lower_camel $resource_name }}DownConverter{})
	})

	Describe("Convert", func() {
		It("works for noop conversions", func() {
			src := &{{ (index .Projects 0).Version }}.{{ upper_camel $resource_name }}{Metadata: core.Metadata{Name: "test"}}
			dst := &{{ (index .Projects 0).Version }}.{{ upper_camel $resource_name }}{}
			err := converter.Convert(src, dst)
			Expect(err).NotTo(HaveOccurred())
			Expect(dst.GetMetadata().Name).To(Equal("test"))
		})

		It("converts all the way up", func() {
			src := &{{ (index .Projects 0).Version }}.{{ upper_camel $resource_name }}{}
			dst := &{{ (index .Projects (add_int (len .Projects) -1)).Version }}.{{ upper_camel $resource_name }}{}
			err := converter.Convert(src, dst)
			Expect(err).NotTo(HaveOccurred())
			Expect(dst.GetMetadata().Name).To(Equal("{{ (index .Projects (add_int (len .Projects) -1)).Version }}"))
		})
		
		It("converts all the way down", func() {
			src := &{{ (index .Projects (add_int (len .Projects) -1)).Version }}.{{ upper_camel $resource_name }}{}
			dst := &{{ (index .Projects 0).Version }}.{{ upper_camel $resource_name }}{}
			err := converter.Convert(src, dst)
			Expect(err).NotTo(HaveOccurred())
			Expect(dst.GetMetadata().Name).To(Equal("{{ (index .Projects 0).Version }}"))
		})
	})
})

type {{ lower_camel $resource_name }}UpConverter struct{}
{{- range .Projects }}
{{- if .NextVersion }}
func ({{ lower_camel $resource_name }}UpConverter) From{{ upper_camel .Version }}To{{ upper_camel .NextVersion }}(src *{{ .Version }}.{{ upper_camel $resource_name }}) *{{ .NextVersion }}.{{ upper_camel $resource_name }} {
	return &{{ .NextVersion }}.{{ upper_camel $resource_name }}{Metadata: core.Metadata{Name: "{{ .NextVersion }}"}}
}
{{- end }}
{{- end }}

type {{ lower_camel $resource_name }}DownConverter struct{}
{{- range .Projects }}
{{- if .PreviousVersion }}
func ({{ lower_camel $resource_name }}DownConverter) From{{ upper_camel .Version }}To{{ upper_camel .PreviousVersion }}(src *{{ .Version }}.{{ upper_camel $resource_name }}) *{{ .PreviousVersion }}.{{ upper_camel $resource_name }} {
	return &{{ .PreviousVersion }}.{{ upper_camel $resource_name }}{Metadata: core.Metadata{Name: "{{ .PreviousVersion }}"}}
}
{{- end }}
{{- end }}

{{- end }}
`))
