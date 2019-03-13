package main

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

//go:generate go run generate.go

func main() {
	do := &cmd.DocsOptions{
		Output: options.Hugo,
	}
	if err := cmd.Run(".", true, do, nil, nil); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
