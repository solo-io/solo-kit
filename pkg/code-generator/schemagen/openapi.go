package schemagen

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/openapi"
	"cuelang.org/go/encoding/protobuf"
	"github.com/rotisserie/eris"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

// TODO (sam-heilbron): CUE deprecates OrderedMap. We should start using ast.File and other APIs in the
// ast package to manipulate the schemas. This requires some more support from ast package
// thus leaving this as a TODO.

// Mapping from protobuf message name to OpenApi schema
type OpenApiSchemas map[string]*openapi.OrderedMap

type OpenApiSchemaGenerator interface {
	GetOpenApiSchemas(project *model.Project, timeout time.Duration) (OpenApiSchemas, error)
}

func NewCueOpenApiSchemaGenerator(importsCollector collector.Collector, protoDir, absoluteRoot string) OpenApiSchemaGenerator {
	return &cueGenerator{
		importsCollector: importsCollector,
		protoDir:         protoDir,
		absoluteRoot:     absoluteRoot,
	}
}

type cueGenerator struct {
	importsCollector collector.Collector
	protoDir         string
	absoluteRoot     string
}

func (c *cueGenerator) GetOpenApiSchemas(project *model.Project, timeout time.Duration) (OpenApiSchemas, error) {
	resultChannel := make(chan struct {
		OpenApiSchemas
		error
	}, 1)

	go func() {
		schemas, err := c.getOpenApiSchemas(project)
		resultChannel <- struct {
			OpenApiSchemas
			error
		}{OpenApiSchemas: schemas, error: err}
	}()

	select {
	case result := <-resultChannel:
		return result.OpenApiSchemas, result.error
	case <-time.After(timeout):
		// OpenApi schema generation should not take longer than the configured timeout.
		// Cue can enter an infinite loop if we define recursive protos. Also some schemas
		// take a while to generate. To protect ourselves from entering increasing schema generation
		// time, we allow a configurable upper limit.
		return nil, eris.New("Timed out while generating open api schemas for project.")
	}
}

func (c *cueGenerator) getOpenApiSchemas(project *model.Project) (OpenApiSchemas, error) {
	oapiSchemas := OpenApiSchemas{}

	// Collect all protobuf definitions including transitive dependencies.
	var imports []string
	for _, projectProto := range project.ProjectConfig.ProjectProtos {
		absoluteProjectProtoPath := filepath.Join(c.absoluteRoot, c.protoDir, projectProto)
		importsForResource, err := c.importsCollector.CollectImportsForFile(c.protoDir, absoluteProjectProtoPath)
		if err != nil {
			return nil, err
		}
		imports = append(imports, importsForResource...)
	}
	imports = stringutils.Unique(imports)

	// Parse protobuf into cuelang
	protobufExtractor := protobuf.NewExtractor(&protobuf.Config{
		Root:   c.protoDir,
		Module: project.ProjectConfig.GoPackage,
		Paths:  imports,
	})

	for _, projectProto := range project.ProjectConfig.ProjectProtos {
		if err := protobufExtractor.AddFile(projectProto, nil); err != nil {
			return nil, err
		}
	}
	instances, err := protobufExtractor.Instances()
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return oapiSchemas, nil
	}

	// Convert cuelang to openapi
	openApiGenerator := &openapi.Generator{
		// k8s structural schemas do not allow $refs, i.e. all references must be expanded
		ExpandReferences: true,
	}

	built := cue.Build(instances)
	for _, builtInstance := range built {
		// Avoid generating openapi for irrelevant proto imports.
		if !strings.HasSuffix(builtInstance.ImportPath, project.ProjectConfig.GoPackage) {
			continue
		}

		if builtInstance.Err != nil {
			return nil, err
		}
		if err = builtInstance.Value().Validate(); err != nil {
			return nil, eris.Errorf("Cue instance validation failed for %s: %+v", project.ProtoPackage, err)
		}
		schemas, err := openApiGenerator.Schemas(builtInstance)
		if err != nil {
			return nil, eris.Errorf("Cue openapi generation failed for %s: %+v", project.ProtoPackage, err)
		}

		// Iterate openapi objects to construct mapping from proto message name to openapi schema
		for _, kv := range schemas.Pairs() {
			oapiSchemas[kv.Key] = kv.Value.(*openapi.OrderedMap)
		}

		return oapiSchemas, err
	}

	return oapiSchemas, nil
}
