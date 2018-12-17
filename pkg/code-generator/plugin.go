package code_generator

import (
	"bytes"
	"github.com/solo-io/solo-kit/pkg/code-generator/cligen"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pseudomuto/protokit"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

// plugin is an implementation of protokit.Plugin
type Plugin struct {
	OutputDescriptors bool
	// merge descriptors to a single file
	MergeDescriptors string
}

func (p *Plugin) Generate(req *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	log.DefaultOut = &bytes.Buffer{}
	if os.Getenv("DEBUG") == "1" {
		log.DefaultOut = os.Stderr
	}

	log.Printf("received request files: %v | params: %v", req.FileToGenerate, req.GetParameter())
	paramString := req.GetParameter()
	params := strings.Split(paramString, ",")
	if len(params) < 1 {
		return nil, errors.Errorf("must provide project file via --solo-kit_opt=${PROJECT_DIR}/project.json[,{DOCS_DIR}]")
	}
	projectFile := params[0]
	var docsDir string
	if len(params) > 1 {
		docsDir = params[1]
	}
	var cliDir string
	if len(params) > 2 {
		cliDir = params[2]
	}

	if projectFile == "" {
		return nil, errors.Errorf(`must provide project file via --solo-kit_out=${PWD}/project.json:${OUT}`)
	}

	// if OutputDescriptors==true save request as a descriptors file
	if p.OutputDescriptors {
		collectedDescriptorsBytes, err := proto.Marshal(req)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal %+v", req)
		}
		if err := ioutil.WriteFile(projectFile+".descriptors", collectedDescriptorsBytes, 0644); err != nil {
			return nil, errors.Wrapf(err, "failed to write %v", projectFile+".descriptors")
		}
	}

	// merge descriptors to a single file, useful for iterating over sets of protos across multiple commands
	if p.MergeDescriptors != "" {
		log.Printf("merging descriptors to file %v", p.MergeDescriptors)
		old := len(req.ProtoFile)
		oldFileToGenerate := len(req.FileToGenerate)
		if _, err := os.Stat(p.MergeDescriptors); err == nil {
			originalRequestBytes, err := ioutil.ReadFile(p.MergeDescriptors)
			if err != nil {
				return nil, errors.Wrapf(err, "reading in merged descriptors file")
			}
			var originalRequest plugin_go.CodeGeneratorRequest
			if err := proto.Unmarshal(originalRequestBytes, &originalRequest); err != nil {
				return nil, errors.Wrapf(err, "parsing merged descriptors file")
			}
			for _, file := range originalRequest.ProtoFile {
				// conflicts are resolved by having the last file win
				// hopefully no files have the same name
				var skipFile bool
				for _, existing := range req.ProtoFile {
					if file.GetName() == existing.GetName() {
						//log.Fatalf("file.GetName() = %v "+
						//	"redefined in 2 packages: %v and %v", file.GetName(), file.GetPackage(), existing.GetPackage())
						skipFile = true
						break
					}
				}
				if skipFile {
					continue
				}
				req.ProtoFile = append(req.ProtoFile, file)
			}
			for _, fileToGen := range originalRequest.FileToGenerate {
				// conflicts are resolved by having the last file win
				// hopefully no files have the same name
				var skipFile bool
				for _, existing := range req.FileToGenerate {
					if fileToGen == existing {
						//log.Fatalf("file-to-generate redefined = %v "+
						//	"redefined in 2 requests: %v and %v", fileToGen, originalRequest.GetFileToGenerate(), req.GetFileToGenerate())
						skipFile = true
						break
					}
				}
				if skipFile {
					continue
				}
				req.FileToGenerate = append(req.FileToGenerate, fileToGen)
			}
		}

		log.Printf("added %v ProtoFile, total: %v", len(req.ProtoFile) - old, len(req.ProtoFile))
		log.Printf("added %v FileToGenerate, total: %v", oldFileToGenerate, req.FileToGenerate)

		collectedDescriptorsBytes, err := proto.Marshal(req)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal %+v", req)
		}
		if err := ioutil.WriteFile(p.MergeDescriptors, collectedDescriptorsBytes, 0644); err != nil {
			return nil, errors.Wrapf(err, "failed to write %v", projectFile+".descriptors")
		}
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

	if docsDir != "" {
		docs, err := docgen.GenerateFiles(project, protokit.ParseCodeGenRequest(req))
		if err != nil {
			return nil, err
		}

		for _, file := range docs {
			resp.File = append(resp.File, &plugin_go.CodeGeneratorResponse_File{
				Name:    proto.String(filepath.Join(docsDir, file.Filename)),
				Content: proto.String(file.Content),
			})
		}
	}

	if cliDir != "" {
		cli, err := cligen.GenerateFiles(project, cliDir, true)
		if err != nil {
			return nil, err
		}

		for _, file := range cli {
			resp.File = append(resp.File, &plugin_go.CodeGeneratorResponse_File{
				Name:    proto.String(filepath.Join(cliDir, file.Filename)),
				Content: proto.String(file.Content),
			})
		}
		gofmt := exec.Cmd{
			Stderr: os.Stderr,
			Stdout: os.Stdout,
			Path: "gofmt",
			Args: []string{"-w", cliDir},
		}
		err = gofmt.Run()
		if err != nil {
			log.Fatalf("%s", err)
			return nil, err
		}
		goimports := exec.Cmd{
			Stderr: os.Stderr,
			Stdout: os.Stdout,
			Path: "goimports",
			Args: []string{"-w", cliDir},
		}
		err = goimports.Run()
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}


