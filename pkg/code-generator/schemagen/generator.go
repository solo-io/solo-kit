package schemagen

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func GenerateOpenApiValidationSchemas(project *model.Project, importsCollector collector.Collector) error {
	if !project.ProjectConfig.GenKubeValidationSchemas {
		log.Printf("Project %s not configured to generate validation schema", project.String())
		return nil
	}

	// Use a tmp directory as the output of schemas
	// The schemas will then be matched with the appropriate CRD
	tmpOutputDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.Remove(tmpOutputDir)

	// Use a directory that is specific to this project
	// This ensures that when we traverse the outputDir, we only traverse project specific schemas
	projectOutputDir := filepath.Join(tmpOutputDir, project.String())
	_ = os.MkdirAll(projectOutputDir, os.ModePerm)

	generator := NewProtocGenerator(project, importsCollector, projectOutputDir)
	return generator.Generate()
}
