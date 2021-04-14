package schemagen

import (
	"log"

	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/openapi"
	"cuelang.org/go/encoding/protobuf"
	"github.com/rotisserie/eris"
	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

//go:generate mockgen -destination mocks/mock_oapi_schema_generator.go -package mocks -source open_api.go OpenApiSchemaGenerator

// TODO: CUE deprecates OrderedMap. We should start using ast.File and other APIs in the
// ast package to manipulate the schemas. This requires some more support from ast package
// thus leaving this as a TODO.

// Mapping from protobuf message name to OpenApi schema
type OpenApiSchemas map[string]*openapi.OrderedMap

type OpenApiSchemaGenerator interface {
	GetOpenApiSchemas(project model.Project, protoDir string) (OpenApiSchemas, error)
}

func NewCueOpenApiSchemaGenerator() OpenApiSchemaGenerator {
	return &cueGenerator{}
}

type cueGenerator struct {
}

func (c *cueGenerator) GetOpenApiSchemas(project model.Project, protoDir string) (OpenApiSchemas, error) {
	oapiSchemas := OpenApiSchemas{}

	if protoDir == "" {
		protoDir = anyvendor.DefaultDepDir
	}

	// Collect all protobuf definitions including transitive dependencies.
	var imports []string
	coll := collector.NewCollector([]string{protoDir}, nil)
	for _, fileDescriptor := range project.DescriptorsWithPath {
		importsForFileDescriptor, err := coll.CollectImportsForFile(protoDir, fileDescriptor.ProtoFilePath)
		if err != nil {
			return nil, err
		}
		imports = append(imports, importsForFileDescriptor...)
	}
	imports = stringutils.Unique(imports)

	// Parse protobuf into cuelang
	cfg := &protobuf.Config{
		Root:   protoDir,
		Module: project.ProjectConfig.GoPackage,
		Paths:  imports,
	}

	ext := protobuf.NewExtractor(cfg)
	for _, fileDescriptor := range project.DescriptorsWithPath {
		if err := ext.AddFile(fileDescriptor.ProtoFilePath, nil); err != nil {
			// TODO - I'm currently getting a `name "Any" not found` error
			// https://github.com/cuelang/cue/blob/3bdfa5d10b6bc232241eb84d7a0b7761b806bac1/encoding/protobuf/parse.go#L295
			return nil, err
		}
	}

	instances, err := ext.Instances()
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return oapiSchemas, nil
	}

	// Convert cuelang to openapi
	generator := &openapi.Generator{
		// k8s structural schemas do not allow $refs, i.e. all references must be expanded
		ExpandReferences: true,
	}

	built := cue.Build(instances)

	for _, builtInstance := range built {
		// Avoid generating openapi for irrelevant proto imports.
		log.Printf("BUILD INSTANCE import path: %s", builtInstance.ImportPath)

		/*
		   if !strings.HasSuffix(builtInstance.ImportPath, grp.Group+"/"+grp.Version) {
		       continue
		   }
		*/

		if builtInstance.Err != nil {
			return nil, err
		}
		if err = builtInstance.Value().Validate(); err != nil {
			return nil, eris.Errorf("Cue instance validation failed for %s: %+v", project.ProjectConfig.Name, err)
		}
		schemas, err := generator.Schemas(builtInstance)
		if err != nil {
			return nil, eris.Errorf("Cue openapi generation failed for %s: %+v", project.ProjectConfig.Name, err)
		}

		// Iterate openapi objects to construct mapping from proto message name to openapi schema
		for _, kv := range schemas.Pairs() {
			oapiSchemas[kv.Key] = kv.Value.(*openapi.OrderedMap)
		}

		return oapiSchemas, err
	}
	return nil, nil

}
