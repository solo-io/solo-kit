package templates

import (
	"text/template"
)

var ResourceGroupSnapshotTemplate = template.Must(template.New("resource_group_snapshot").Funcs(Funcs).Parse(
	`package {{ .Project.ProjectConfig.Version }}

{{/* creating a variable that lets us understand how many resources are hashable input resources. */}}

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"log"

	{{ .Imports }}
	"k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/rotisserie/eris"
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
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	fields = append(fields, zap.Uint64("{{ lower_camel .PluralName }}", {{ upper_camel .PluralName }}Hash ))
{{- end}}
	snapshotHash, err := s.Hash(hasher)
	if err != nil {
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	return append(fields, zap.Uint64("snapshotHash",  snapshotHash))
}

func (s *{{ .GoName }}Snapshot) GetResourcesList(resource resources.Resource) (resources.ResourceList, error) {
	switch resource.(type) {
{{- range .Resources }}
	case *{{ .ImportPrefix }}{{ .Name }}:
		return s.{{ upper_camel .PluralName }}.AsResources(), nil
{{- end }}
	default:
		return resources.ResourceList{}, eris.New("did not contain the input resource type returning empty list")
	}
}

func (s *{{ .GoName }}Snapshot) RemoveFromResourceList(resource resources.Resource) error {
	refKey := resource.GetMetadata().Ref().Key()
	switch resource.(type) {
{{- range .Resources }}
	case *{{ .ImportPrefix }}{{ .Name }}:
		{{/* no need to sort because it is already sorted */}}
		for i, res := range s.{{ upper_camel .PluralName }} {
			if refKey == res.GetMetadata().Ref().Key() {
				s.{{ upper_camel .PluralName }} = append(s.{{ upper_camel .PluralName }}[:i], s.{{ upper_camel .PluralName }}[i+1:]...)
				break
			}
		}
		return nil	
{{- end }}
	default:
		return eris.Errorf("did not remove the reousource because its type does not exist [%T]", resource)
	}
}

func (s *{{ .GoName }}Snapshot) UpsertToResourceList(resource resources.Resource) error {
	refKey := resource.GetMetadata().Ref().Key() 
	switch typed := resource.(type) {
{{- range .Resources }}
	case *{{ .ImportPrefix }}{{ .Name }}:
		updated := false
		for i, res := range s.{{ upper_camel .PluralName }} {
			if refKey == res.GetMetadata().Ref().Key() {
				s.{{ upper_camel .PluralName }}[i] = typed
				updated = true
			}
		}
		if !updated {
			s.{{ upper_camel .PluralName }} = append(s.{{ upper_camel .PluralName }}, typed)
		}
		s.{{ upper_camel .PluralName }}.Sort()
		return nil	
{{- end }}
	default:
		return eris.Errorf("did not add/replace the resource type because it does not exist %T", resource)
	}
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
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
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

var {{.GoName }}GvkToHashableResource = map[schema.GroupVersionKind]func() resources.HashableResource {
{{- range .Resources}}
	{{ .ImportPrefix }}{{ .Name }}GVK: {{ .ImportPrefix }}New{{ .Name }}HashableResource,
{{- end }}	
}

`))
