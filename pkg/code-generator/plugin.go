package code_generator

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pseudomuto/protokit"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

const (
	ParameterPrefix_OUTPUT   = "OUTPUT="
	ParameterPrefix_Generate = "GENERATE="

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

	switch {
	case strings.HasPrefix(param, ParameterPrefix_OUTPUT):
		param = strings.TrimPrefix(param, ParameterPrefix_OUTPUT)
		return p.outputProtoRequest(param, req)
	case strings.HasPrefix(param, ParameterPrefix_Generate):
		param = strings.TrimPrefix(param, ParameterPrefix_Generate)
	default:
		// default does generate, but is backwards compatible with no GENERATE= prefix
	}

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

//func foo() {
//	// merge descriptors to a single file, useful for iterating over sets of protos across multiple commands
//	log.Printf("merging descriptors to file %v", mergedRequestFilePath)
//	old := len(req.ProtoFile)
//	oldFileToGenerate := len(req.FileToGenerate)
//	if _, err := os.Stat(mergedRequestFilePath); err == nil {
//		originalRequestBytes, err := ioutil.ReadFile(p.MergeDescriptors)
//		if err != nil {
//			return nil, errors.Wrapf(err, "reading in merged descriptors file")
//		}
//		var originalRequest plugin_go.CodeGeneratorRequest
//		if err := proto.Unmarshal(originalRequestBytes, &originalRequest); err != nil {
//			return nil, errors.Wrapf(err, "parsing merged descriptors file")
//		}
//		for _, file := range originalRequest.ProtoFile {
//			// conflicts are resolved by having the last file win
//			// hopefully no files have the same name
//			var skipFile bool
//			for _, existing := range req.ProtoFile {
//				if file.GetName() == existing.GetName() {
//					//log.Fatalf("file.GetName() = %v "+
//					//	"redefined in 2 packages: %v and %v", file.GetName(), file.GetPackage(), existing.GetPackage())
//					skipFile = true
//					break
//				}
//			}
//			if skipFile {
//				continue
//			}
//			req.ProtoFile = append(req.ProtoFile, file)
//		}
//		for _, fileToGen := range originalRequest.FileToGenerate {
//			// conflicts are resolved by having the last file win
//			// hopefully no files have the same name
//			var skipFile bool
//			for _, existing := range req.FileToGenerate {
//				if fileToGen == existing {
//					//log.Fatalf("file-to-generate redefined = %v "+
//					//	"redefined in 2 requests: %v and %v", fileToGen, originalRequest.GetFileToGenerate(), req.GetFileToGenerate())
//					skipFile = true
//					break
//				}
//			}
//			if skipFile {
//				continue
//			}
//			req.FileToGenerate = append(req.FileToGenerate, fileToGen)
//		}
//	}
//
//	log.Printf("added %v ProtoFile, total: %v", len(req.ProtoFile)-old, len(req.ProtoFile))
//	log.Printf("added %v FileToGenerate, total: %v", oldFileToGenerate, req.FileToGenerate)
//
//	collectedDescriptorsBytes, err := proto.Marshal(req)
//	if err != nil {
//		return nil, errors.Wrapf(err, "failed to marshal %+v", req)
//	}
//	if err := ioutil.WriteFile(p.MergeDescriptors, collectedDescriptorsBytes, 0644); err != nil {
//		return nil, errors.Wrapf(err, "failed to write %v", param+".descriptors")
//	}
//}

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

	project, err := parser.ParseRequest(projectFilePath, req)
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
		docs, err := docgen.GenerateFiles(project, protokit.ParseCodeGenRequest(req))
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
