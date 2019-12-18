package templates

import (
	"text/template"
)

var ResourceGroupSnapshotTemplate = template.Must(template.New("resource_group_snapshot").Funcs(Funcs).Parse(
	`package {{ .Project.ProjectConfig.Version }}

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"log"

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

func (s {{ .GoName }}Snapshot) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
{{- range .Resources}}
	if _, err := s.hash{{ upper_camel .PluralName }}(hasher); err != nil {
		return 0, err
	}
{{- end}}
	return hasher.Sum64(), nil
}

{{- $ResourceGroup := . }}
{{- range .Resources }}

func (s {{ $ResourceGroup.GoName }}Snapshot) hash{{ upper_camel .PluralName }}(hasher hash.Hash64) (uint64, error) {
	{{- if .SkipHashingAnnotations }}
	clonedList := s.{{ upper_camel .PluralName }}.Clone()
	for _, v := range clonedList {
		v.Metadata.Annotations = nil
	}
	return hashutils.HashAllSafe(hasher, clonedList.AsInterfaces()...)
	{{- else }}
	return hashutils.HashAllSafe(hasher, s.{{ upper_camel .PluralName }}.AsInterfaces()...)
	{{- end }}
}
{{- end}}

func (s {{ .GoName }}Snapshot) HashFields() []zap.Field {
	var fields []zap.Field
	hasher := fnv.New64()
{{- range .Resources}}
	{{ upper_camel .PluralName }}Hash, err := s.hash{{ upper_camel .PluralName }}(hasher)
	if err != nil {
		log.Println(err)
	}
	fields = append(fields, zap.Uint64("{{ lower_camel .PluralName }}", {{ upper_camel .PluralName }}Hash ))
{{- end}}
	snapshotHash, _ := s.Hash(hasher)
	return append(fields, zap.Uint64("snapshotHash",  snapshotHash))
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
	snapshotHash, err := s.Hash(nil)
	if err != nil {
		log.Println(err)
	}
	return {{ .GoName }}SnapshotStringer{
		Version: snapshotHash,
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
