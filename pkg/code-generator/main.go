package code_generator

import (
	"log"
	"os"

	"github.com/pseudomuto/protokit"
)

func Main() {
	outputDescriptors := os.Getenv("OUTPUT_DESCRIPTORS") == "1"
	mergeOutputDescriptors := os.Getenv("OUTPUT_MERGED_DESCRIPTORS_FILE")
	plugin := &Plugin{OutputDescriptors: outputDescriptors, MergeDescriptors: mergeOutputDescriptors}
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
