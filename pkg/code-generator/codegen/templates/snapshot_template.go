package templates

import (
	"text/template"
)

var ResourceGroupSnapshotTemplate = template.Must(template.New("resource_group_snapshot").Funcs(Funcs).Parse(
	`package {{ .ApiGroup.ResourceGroupGoPackageShort }}

import (
	"fmt"

	{{ .Imports }}
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/go-utils/hashutils"
	"go.uber.org/zap"
)

type {{ .GoName }}Snapshot struct {
{{- range .Resources}}
	{{ upper_camel .PluralName }} {{ .ImportPrefix }}{{ .Name }}List
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
	return hashutils.HashAll(
{{- range .Resources}}
		s.hash{{ upper_camel .PluralName }}(),
{{- end}}
	)
}

{{- $ResourceGroup := . }}
{{- range .Resources }}

func (s {{ $ResourceGroup.GoName }}Snapshot) hash{{ upper_camel .PluralName }}() uint64 {
	return hashutils.HashAll(s.{{ upper_camel .PluralName }}.AsInterfaces()...)
}
{{- end}}

func (s {{ .GoName }}Snapshot) HashFields() []zap.Field {
	var fields []zap.Field

{{- range .Resources}}
	fields = append(fields, zap.Uint64("{{ lower_camel .PluralName }}", s.hash{{ upper_camel .PluralName }}() ))
{{- end}}

	return append(fields, zap.Uint64("snapshotHash",  s.Hash()))
}

type {{ .GoName }}SnapshotStringer struct {
	Version              uint64
{{- range .Resources}}
	{{ upper_camel .PluralName }} []string
{{- end}}
}

func (ss {{ .GoName }}SnapshotStringer) String() string {
	s := fmt.Sprintf("{{ .GoName }}Snapshot %v\n", ss.Version)
{{- range .Resources}}

	s += fmt.Sprintf("  {{ upper_camel .PluralName }} %v\n", len(ss.{{ upper_camel .PluralName }}))
	for _, name := range ss.{{ upper_camel .PluralName }} {
		s += fmt.Sprintf("    %v\n", name)
	}
{{- end}}

	return s
}

func (s {{ .GoName }}Snapshot) Stringer() {{ .GoName }}SnapshotStringer {
	return {{ .GoName }}SnapshotStringer{
		Version: s.Hash(),
{{- range .Resources}}
{{- if .ClusterScoped }}
		{{ upper_camel .PluralName }}: s.{{ upper_camel .PluralName }}.Names(),
{{- else }}
		{{ upper_camel .PluralName }}: s.{{ upper_camel .PluralName }}.NamespacesDotNames(),
{{- end }}
{{- end}}
	}
}
`))
