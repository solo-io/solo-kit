package main

import (
	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
)

//go:generate go run generate.go
//go:generate ./api/v1/generate.sh

func main() {

	log.Printf("starting generate")
	if err := cmd.Generate(cmd.GenerateOptions{
		RelativeRoot:       ".",
		CompileProtos:      true,
		SkipGenMocks:       true,
		SkipGeneratedTests: true,
		ProtoDepConfig: &anyvendor.Config{
			Local: &anyvendor.Local{
				Patterns: []string{"test/**/*.proto", "api/**/*.proto", sk_anyvendor.SoloKitMatchPattern},
			},
			Imports: []*anyvendor.Import{
				{
					ImportType: &anyvendor.Import_GoMod{GoMod: sk_anyvendor.ExtProtoMatcher},
				},
				{
					ImportType: &anyvendor.Import_GoMod{GoMod: sk_anyvendor.EnvoyValidateProtoMatcher},
				},
				{
					ImportType: &anyvendor.Import_GoMod{GoMod: sk_anyvendor.GogoProtoMatcher},
				},
			},
		},
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
