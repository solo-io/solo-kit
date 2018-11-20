package main

import (
	"log"
	"os"

	"github.com/pseudomuto/protokit"
	"github.com/solo-io/solo-kit/pkg/code-generator"
)

func main() {
	outputDescriptors := os.Getenv("OUTPUT_DESCRIPTORS") == "1"
	plugin := &code_generator.Plugin{OutputDescriptors: outputDescriptors}
	// use this to debug without running protoc
	if descriptorsFile := os.Getenv("USE_DESCRIPTORS"); descriptorsFile != "" {
		// descriptorsFile e.g.: "projects/supergloo/api/v1/project.json.descriptors"
		f, err := os.Open(descriptorsFile)
		if err != nil {
			log.Fatal(err)
		}
		if err := protokit.RunPluginWithIO(plugin, f, os.Stdout); err != nil {
			log.Fatal(err)
		}
	}
	if err := protokit.RunPlugin(plugin); err != nil {
		log.Fatal(err)
	}
}
