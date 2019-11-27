package main

import (
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
)

//go:generate go run generate.go
//go:generate ./api/v1/generate.sh

func main() {

	log.Printf("starting generate")
	if err := cmd.Generate(cmd.GenerateOptions{
		RelativeRoot:  ".",
		CompileProtos: true,
		SkipGenMocks:  true,
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
