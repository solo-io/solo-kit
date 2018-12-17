package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

func Gopath() string {
	return os.Getenv("GOPATH")
}

var commonImports = []string{
	"-I" + Gopath() + "/src",
	"-I" + Gopath() + "/src/github.com/solo-io/solo-kit/api/external",
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}

func run() error {
	relativeRoot := flag.String("r", "../solo-projects/projects/gloo", "path to project absoluteRoot")
	compileProtos := flag.Bool("gogo", true, "compile normal gogo protos")
	flag.Parse()

	absoluteRoot, err := filepath.Abs(*relativeRoot)
	if err != nil {
		return err
	}

	var projectDirs []string

	// discover all project.json
	if err := filepath.Walk(absoluteRoot, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, "project.json") {
			return nil
		}
		projectDirs = append(projectDirs, filepath.Dir(path))
		return nil
	}); err != nil {
		return err
	}

	for _, inDir := range projectDirs {
		outDir := strings.Replace(inDir, "api", "pkg/api", -1)

		imports := append(commonImports, "-I"+inDir)

		tmpFile, err := ioutil.TempFile("", "solo-kit-gen-")
		if err != nil {
			return err
		}
		if err := tmpFile.Close(); err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		var descriptors []*descriptor.FileDescriptorSet

		if err := filepath.Walk(inDir, func(path string, info os.FileInfo, err error) error {
			if !strings.HasSuffix(path, ".proto") {
				return nil
			}
			if err := writeDescriptors(path, tmpFile.Name(), imports, *compileProtos); err != nil {
				return err
			}
			desc, err := readDescriptors(tmpFile.Name())
			if err != nil {
				return err
			}
			descriptors = append(descriptors, desc)
			return nil
		}); err != nil {
			return err
		}

		projectConfig, err := model.LoadProjectConfig(inDir + "/project.json")
		if err != nil {
			return err
		}

		project, err := parser.ProcessDescriptors(projectConfig, descriptors)
		if err != nil {
			return err
		}

		code, err := codegen.GenerateFiles(project, true)
		if err != nil {
			return err
		}

		if project.DocsDir != "" {
			docs, err := docgen.GenerateFiles(project)
			if err != nil {
				return err
			}

			for _, file := range docs {
				path := filepath.Join(absoluteRoot, project.DocsDir, file.Filename)
				if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
					return err
				}
				if err := ioutil.WriteFile(path, []byte(file.Content), 0644); err != nil {
					return err
				}
			}
		}

		for _, file := range code {
			path := filepath.Join(outDir, file.Filename)
			if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
				return err
			}
			if err := ioutil.WriteFile(path, []byte(file.Content), 0644); err != nil {
				return err
			}
			if err := exec.Command("gofmt", "-w", path).Run(); err != nil {
				return err
			}

			if err := exec.Command("goimports", "-w", path).Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeDescriptors(protoFile, toFile string, imports []string, compileProtos bool) error {
	cmd := exec.Command("protoc")
	cmd.Args = append(cmd.Args, imports...)

	if compileProtos {
		cmd.Args = append(cmd.Args,
			"--gogo_out="+
				"Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,"+
				"Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types"+
				":"+Gopath()+"/src/")
	}

	cmd.Args = append(cmd.Args, "-o"+toFile, "--include_imports", "--include_source_info",
		protoFile)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "%v failed: %s", cmd.Args, out)
	}
	return nil
}

func readDescriptors(fromFile string) (*descriptor.FileDescriptorSet, error) {
	var desc descriptor.FileDescriptorSet
	protoBytes, err := ioutil.ReadFile(fromFile)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(protoBytes, &desc); err != nil {
		return nil, err
	}
	return &desc, nil
}
