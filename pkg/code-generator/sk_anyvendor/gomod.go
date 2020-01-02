package sk_anyvendor

import (
	"github.com/solo-io/anyvendor/anyvendor"
)

const (
	SoloKitMatchPattern = "**/solo-kit.json"
)

var (
	// offer sane defaults for proto vendoring
	DefaultMatchPatterns = []string{anyvendor.ProtoMatchPattern, SoloKitMatchPattern}

	// matches ext.proto for solo hash gen
	ExtProtoMatcher = &anyvendor.GoModImport{
		Package:  "github.com/solo-io/protoc-gen-ext",
		Patterns: []string{"extproto/*.proto"},
	}

	// matches validate.proto which is needed by envoy protos
	EnvoyValidateProtoMatcher = &anyvendor.GoModImport{
		Package:  "github.com/envoyproxy/protoc-gen-validate",
		Patterns: []string{"validate/*.proto"},
	}

	// matches all solo-kit protos, useful for any projects using solo-kit
	SoloKitProtoMatcher = &anyvendor.GoModImport{
		Package:  "github.com/solo-io/solo-kit",
		Patterns: []string{"api/**/*.proto", "api/" + SoloKitMatchPattern},
	}

	// matches gogo.proto, used for gogoproto code gen.
	GogoProtoMatcher = &anyvendor.GoModImport{
		Package:  "github.com/gogo/protobuf",
		Patterns: []string{"gogoproto/*.proto"},
	}

	// default match options which should be used when creating a solo-kit project
	DefaultExternalMatchOptions = map[string][]string{
		ExtProtoMatcher.Package:           ExtProtoMatcher.Patterns,
		EnvoyValidateProtoMatcher.Package: EnvoyValidateProtoMatcher.Patterns,
		SoloKitProtoMatcher.Package:       SoloKitProtoMatcher.Patterns,
		GogoProtoMatcher.Package:          GogoProtoMatcher.Patterns,
	}
)

func CreateDefaultMatchOptions(local []string) *Imports {
	return &Imports{
		Local:    local,
		External: DefaultExternalMatchOptions,
	}
}

type Imports struct {
	Local    []string
	External map[string][]string
}

func (i *Imports) ConvertToAnvendorConfig() *anyvendor.Config {
	result := &anyvendor.Config{}
	var imports []*anyvendor.GoModImport
	for pkg, patterns := range i.External {
		imports = append(imports, &anyvendor.GoModImport{
			Patterns: patterns,
			Package:  pkg,
		})
	}
	result.Local = &anyvendor.Local{
		Patterns: i.Local,
	}
	return result
}
