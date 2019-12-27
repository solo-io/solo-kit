package main

import (
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/protodep"
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
		ProtoDepConfig: &protodep.Config{
			Local: &protodep.Local{
				Patterns: []string{"test/**/*.proto", "api/**/*.proto", protodep.SoloKitMatchPattern},
			},
			Imports: []*protodep.Import{
				{
					ImportType: &protodep.Import_GoMod{GoMod: protodep.ExtProtoMatcher},
				},
				{
					ImportType: &protodep.Import_GoMod{GoMod: protodep.EnvoyValidateProtoMatcher},
				},
				{
					ImportType: &protodep.Import_GoMod{GoMod: protodep.GogoProtoMatcher},
				},
				{
					ImportType: &protodep.Import_GoMod{GoMod: protodep.GoogleProtoMatcher},
				},
			},
		},
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
