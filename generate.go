package main

import (
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
)

//go:generate go run generate.go

func main() {

	log.Printf("starting generate")
	if err := cmd.Generate(cmd.GenerateOptions{
		RelativeRoot:  ".",
		CompileProtos: true,
		SkipGenMocks:  true,
		GenDocs: &cmd.DocsOptions{
			Output: options.Hugo,
		},
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
