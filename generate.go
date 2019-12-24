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
		PreRunFuncs: []cmd.RunFunc{
			protodep.PreRunProtoVendor(".",
				[]protodep.MatchOptions{
					protodep.ExtProtoMatcher,
					protodep.ValidateProtoMatcher,
				},
			),
		},
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
