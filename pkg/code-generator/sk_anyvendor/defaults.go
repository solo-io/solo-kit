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
		Package: "github.com/solo-io/protoc-gen-ext",
		Patterns: []string{
			"extproto/*.proto",
			"external/google/**/*.proto",
		},
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

	// default match options which should be used when creating a solo-kit project
	DefaultExternalMatchOptions = map[string][]string{
		ExtProtoMatcher.Package:           ExtProtoMatcher.Patterns,
		EnvoyValidateProtoMatcher.Package: EnvoyValidateProtoMatcher.Patterns,
		SoloKitProtoMatcher.Package:       SoloKitProtoMatcher.Patterns,
	}
)

func CreateDefaultMatchOptions(local []string) *Imports {
	return &Imports{
		Local:    local,
		External: DefaultExternalMatchOptions,
	}
}

// this type represents a solo-kit abstraction of the anyvendor API for vendoring non-go files.
type Imports struct {
	// files which should be gathered from the local repo
	Local []string
	// files which should be gathered from other go.mod repos
	External map[string][]string
}

func (i *Imports) ToAnyvendorConfig() *anyvendor.Config {
	var imports []*anyvendor.Import
	for pkg, patterns := range i.External {
		imports = append(imports, &anyvendor.Import{
			ImportType: &anyvendor.Import_GoMod{
				GoMod: &anyvendor.GoModImport{
					Patterns: patterns,
					Package:  pkg,
				},
			},
		})
	}
	return &anyvendor.Config{
		Local: &anyvendor.Local{
			Patterns: i.Local,
		},
		Imports: imports,
	}
}
