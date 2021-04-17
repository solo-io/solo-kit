package schemagen

import (
	"path/filepath"
	"strings"

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

type cueGenerator struct {
	importsCollector collector.Collector
	protoDir         string
	absoluteRoot     string
}

func (c *cueGenerator) GetOpenApiSchemas(project *model.Project) (OpenApiSchemas, error) {
	/**
	TODO (sam-heilbron)
		- Don't short circuit projects that are not 'gateway.solo.io'. This is to speed up debugging
			the gateway project which isn't compiling properly
		- At the moment we parse projectProtos. Should we also be parsing additional imports (non-project protos)?
	*/
	oapiSchemas := OpenApiSchemas{}

	if project.ProtoPackage != "gateway.solo.io" {
		return oapiSchemas, nil
	}

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
		Module: c.absoluteRoot, // TODO - project.ProjectConfig.GoPackage?,
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
