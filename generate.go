package main

import (
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
)

//go:generate go run generate.go
//go:generate ./api/v1/generate.sh
//go:generate ./pkg/api/v1/apiserver/generate.sh

func main() {
	log.Printf("starting generate")
	if err := cmd.Generate(cmd.GenerateOptions{
		RelativeRoot:       ".",
		CompileProtos:      true,
		SkipGenMocks:       true,
		SkipGeneratedTests: true,
		ExternalImports: &sk_anyvendor.Imports{
			Local: []string{"test/**/*.proto", "api/**/*.proto", sk_anyvendor.SoloKitMatchPattern},
			External: map[string][]string{
				sk_anyvendor.ExtProtoMatcher.Package:           sk_anyvendor.ExtProtoMatcher.Patterns,
				sk_anyvendor.EnvoyValidateProtoMatcher.Package: sk_anyvendor.EnvoyValidateProtoMatcher.Patterns,
			},
		},
		ValidationSchemaOptions: &schemagen.ValidationSchemaOptions{
			// Path to where test CRDs are stored
			CrdDirectory:   "test/mocks/crds",
			JsonSchemaTool: "protoc",
			MessagesWithEmptySchema: []string{
				"core.solo.io.Status",
			},
		},
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
