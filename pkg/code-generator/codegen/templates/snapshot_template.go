package templates

import (
	"text/template"
)

var ResourceGroupSnapshotTemplate = template.Must(template.New("resource_group_snapshot").Funcs(Funcs).Parse(
	`package {{ .Project.Version }}

import (
	{{ .Imports }}
	"github.com/mitchellh/hashstructure"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"go.uber.org/zap"
)

type {{ .GoName }}Snapshot struct {
{{- range .Resources}}
	{{ upper_camel .PluralName }} {{ .ImportPrefix }}{{ upper_camel .PluralName }}ByNamespace
{{- end}}
}

func (s {{ .GoName }}Snapshot) Clone() {{ .GoName }}Snapshot {
	return {{ .GoName }}Snapshot{
{{- range .Resources}}
		{{ upper_camel .PluralName }}: s.{{ upper_camel .PluralName }}.Clone(),
{{- end}}
	}
}

func (s {{ .GoName }}Snapshot) Hash() uint64 {
	return s.hashStruct([]uint64{
{{- range .Resources}}
		s.hash{{ upper_camel .PluralName }}(),
{{- end}}
	})
}

{{- $ResourceGroup := . }}
{{- range .Resources }}

func (s {{ $ResourceGroup.GoName }}Snapshot) hash{{ upper_camel .PluralName }}() uint64 {
	var hashes []uint64
	for _, res := range s.{{ upper_camel .PluralName }}.List() {
		hashes = append(hashes, resources.HashResource(res))
	}
	return s.hashStruct(hashes)
}
{{- end}}

func (s {{ .GoName }}Snapshot) HashFields() []zap.Field {
	var fields []zap.Field

{{- range .Resources}}
	fields = append(fields, zap.Uint64("{{ lower_camel .PluralName }}", s.hash{{ upper_camel .PluralName }}() ))
{{- end}}

	return append(fields, zap.Uint64("snapshotHash",  s.Hash()))
}
 
func (s {{ .GoName }}Snapshot) hashStruct(v interface{}) uint64 {
	h, err := hashstructure.Hash(v, nil)
	 if err != nil {
		 panic(err)
	 }
	 return h
}


`))
