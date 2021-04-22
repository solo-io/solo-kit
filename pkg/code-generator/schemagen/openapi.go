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
	GetOpenApiSchemas(project *model.Project) (OpenApiSchemas, error)
}

func NewCueOpenApiSchemaGenerator(importsCollector collector.Collector, protoDir, absoluteRoot string) OpenApiSchemaGenerator {
	return &cueGenerator{
		importsCollector: importsCollector,
		protoDir:         protoDir,
		absoluteRoot:     absoluteRoot,
	}
}

// OpenApi schema generation should not take longer than `projectTimeout`.
// Cue can enter an infinite loop if we define recursive protos. To protect ourselves
// from entering this loop, we set an upper limit for how long schema gen should take.
const projectCycleTimeout = time.Second * 60 // TODO (sam-heilbron) for debuggigng its so high

type cueGenerator struct {
	importsCollector collector.Collector
	protoDir         string
	absoluteRoot     string
}

func (c *cueGenerator) GetOpenApiSchemas(project *model.Project) (OpenApiSchemas, error) {
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
	case <-time.After(projectCycleTimeout):
		return nil, eris.New("Timed out while generating open api schemas for project. This likely means you have created a recursive proto definition.")
	}
}

func (c *cueGenerator) getOpenApiSchemas(project *model.Project) (OpenApiSchemas, error) {
	/**
	TODO (sam-heilbron)
		- At the moment we parse projectProtos. Should we also be parsing additional imports (non-project protos)?
	*/
	oapiSchemas := OpenApiSchemas{}

	relevantProjectProtos := append([]string{}, project.ProjectConfig.ProjectProtos...)
	relevantProjectProtos = append(relevantProjectProtos, project.ProjectConfig.Imports...)

	// TEMP - for debugging
	relevantProjectProtos = []string{
		"github.com/solo-io/gloo/projects/gateway/api/v1/gateway.proto",
	}

	// Collect all protobuf definitions including transitive dependencies.
	var imports []string
	for _, projectProto := range relevantProjectProtos {
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
		Module: "github.com/solo-io/gloo", // c.absoluteRoot, // TODO - project.ProjectConfig.GoPackage?,
		Paths:  imports,
	})

	for _, projectProto := range relevantProjectProtos {
		if err := protobufExtractor.AddFile(projectProto, nil); err != nil {
			return nil, err
		}
	}
	instances, err := protobufExtractor.Instances()
	if err != nil {
		return nil, err
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
			return nil, eris.Errorf("Cue instance validation failed for %s: %+v", project.ProjectConfig.GoPackage, err)
		}
		schemas, err := openApiGenerator.Schemas(builtInstance)
		if err != nil {
			return nil, eris.Errorf("Cue openapi generation failed for %s: %+v", project.ProjectConfig.GoPackage, err)
		}

		// Iterate openapi objects to construct mapping from proto message name to openapi schema
		for _, kv := range schemas.Pairs() {
			oapiSchemas[kv.Key] = kv.Value.(*openapi.OrderedMap)
		}

		return oapiSchemas, err
	}

	return oapiSchemas, nil
}
