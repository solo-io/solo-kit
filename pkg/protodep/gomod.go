package protodep

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
	DefaultMatchOptions = []*anyvendor.Import{
		{
			ImportType: &anyvendor.Import_GoMod{
				GoMod: ExtProtoMatcher,
			},
		},
		{
			ImportType: &anyvendor.Import_GoMod{
				GoMod: EnvoyValidateProtoMatcher,
			},
		},
		{
			ImportType: &anyvendor.Import_GoMod{
				GoMod: SoloKitProtoMatcher,
			},
		},
		{
			ImportType: &anyvendor.Import_GoMod{
				GoMod: GogoProtoMatcher,
			},
		},
	}
)
