package schemagen

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type protocGenerator struct {
	// The path to the tmp directory used to write schemas to
	tmpSchemaOutputDir string
	project            *model.Project
	importsCollector   collector.Collector
	protocExecutor     collector.ProtocExecutor
}

func NewProtocGenerator(project *model.Project, importsCollector collector.Collector, outputDir string) *protocGenerator {
	return &protocGenerator{
		tmpSchemaOutputDir: outputDir,
		project:            project,
		importsCollector:   importsCollector,
		protocExecutor: &collector.OpenApiProtocExecutor{
			OutputDir: outputDir,
			Project:   project,
		},
	}
}

func (p *protocGenerator) Generate() error {
	absoluteRoot, err := filepath.Abs(anyvendor.DefaultDepDir)
	if err != nil {
		return err
	}

	// 1. Generate the openApiSchemas for the project, writing them to a temp directory (tmpSchemaOutputDir)
	for _, projectProto := range p.project.ProjectConfig.ProjectProtos {
		if err := p.generateSchemasForProjectProto(absoluteRoot, projectProto); err != nil {
			return err
		}
	}

	// 2. Walk the tmpSchemaOutputDir and apply the validation schemas to the inputted CRDs
	return p.processGeneratedSchemas()
}

func (p *protocGenerator) generateSchemasForProjectProto(root, projectProtoFile string) error {
	imports, err := p.importsCollector.CollectImportsForFile(root, filepath.Join(root, projectProtoFile))
	if err != nil {
		return errors.Wrapf(err, "collecting imports for proto file")
	}

	// we don't use the output of protoc so use a temp file
	tmpFile, err := ioutil.TempFile("", "schema-gen-")
	if err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if err := p.protocExecutor.Execute(projectProtoFile, tmpFile.Name(), imports); err != nil {
		return errors.Wrapf(err, "executing protoc")
	}

	return nil
}

func (p *protocGenerator) processGeneratedSchemas() error {
	return filepath.Walk(p.tmpSchemaOutputDir, func(schemaFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(schemaFile, ".yaml") {
			return nil
		}

		log.Printf("Generated Schema File: %s", schemaFile)

		// TODO (sam-heilbron) Do something with the generated schemas

		return nil
	})
}
