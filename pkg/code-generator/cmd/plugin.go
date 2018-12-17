package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

const (
	coreApiRoot = "github.com/solo-io/solo-kit/api/"
)

// plugin is an implementation of protokit.Plugin
type Plugin struct {
	OutputRequest bool
}

func (p *Plugin) Generate(req *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	res, err := p.generate(req)
	if err != nil {
		return &plugin_go.CodeGeneratorResponse{
			Error: proto.String(err.Error()),
		}, err
	}
	return res, nil
}

func (p *Plugin) generate(req *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	log.DefaultOut = &bytes.Buffer{}
	if os.Getenv("DEBUG") == "1" {
		log.DefaultOut = os.Stderr
	}

	log.Printf("received request files: %v | params: %v", req.FileToGenerate, req.GetParameter())
	param := req.GetParameter()

	return p.generateCode(param, req)
}

func (p *Plugin) outputProtoRequest(outputPath string, req *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	collectedDescriptorsBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal %+v", req)
	}

	if outputPath == "" {
		outputPath = "codegen_request.pb"
	}
	return &plugin_go.CodeGeneratorResponse{
		File: []*plugin_go.CodeGeneratorResponse_File{
			{
				Name:    proto.String(outputPath),
				Content: proto.String(string(collectedDescriptorsBytes)),
			},
		},
	}, nil
}

func (p *Plugin) generateCode(projectFilePath string, req *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	if projectFilePath == "" {
		return nil, errors.Errorf("must provide parameter file via --solo-kit_opt=PARAM. " +
			"typically this should be the path to your ${PWD}/project.json")
	}

	// if OutputRequest==true save request as a descriptors file
	if p.OutputRequest {
		collectedDescriptorsBytes, err := proto.Marshal(req)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal %+v", req)
		}
		if err := ioutil.WriteFile(projectFilePath+".descriptors", collectedDescriptorsBytes, 0644); err != nil {
			return nil, errors.Wrapf(err, "failed to write %v", projectFilePath+".descriptors")
		}
	}

	// add core protos to proto request
	// this allows us to gennerate docs for core protos
	for _, file := range req.ProtoFile {
	addDepToGenFiles:
		for _, dep := range file.Dependency {
			for _, genFile := range req.FileToGenerate {
				if genFile == dep {
					continue addDepToGenFiles
				}
			}
			// TODO: make configurable? core root
			if !strings.HasPrefix(dep, coreApiRoot) {
				continue
			}
			req.FileToGenerate = append(req.FileToGenerate, dep)
		}
	}

	projectFile, err := model.LoadProjectConfig(projectFilePath)
	if err != nil {
		return nil, err
	}

	project, err := parser.ParseRequest(projectFile, req)
	if err != nil {
		return nil, err
	}

	code, err := codegen.GenerateFiles(project, true)
	if err != nil {
		return nil, err
	}

	log.Printf("%v", project)
	log.Printf("%v", code)

	resp := new(plugin_go.CodeGeneratorResponse)

	for _, file := range code {
		resp.File = append(resp.File, &plugin_go.CodeGeneratorResponse_File{
			Name:    proto.String(file.Filename),
			Content: proto.String(file.Content),
		})
	}

	if project.DocsDir != "" {
		docs, err := docgen.GenerateFiles(project)
		if err != nil {
			return nil, err
		}

		for _, file := range docs {
			log.Printf("doc: %v", file.Filename)
			resp.File = append(resp.File, &plugin_go.CodeGeneratorResponse_File{
				Name:    proto.String(filepath.Join(project.DocsDir, file.Filename)),
				Content: proto.String(file.Content),
			})
		}
		log.Printf("%v", docs)
	}

	return resp, nil
}
