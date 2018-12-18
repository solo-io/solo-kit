package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/pkg/utils/stringutils"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	if i == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var customImports arrayFlags

func run() error {
	relativeRoot := flag.String("r", "", "path to project absoluteRoot")
	compileProtos := flag.Bool("gogo", true, "compile normal gogo protos")
	flag.Var(&customImports, "i", "import additional directories as proto roots")
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

		tmpFile, err := ioutil.TempFile("", "solo-kit-gen-")
		if err != nil {
			return err
		}
		if err := tmpFile.Close(); err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		var descriptors []*descriptor.FileDescriptorSet

		var projectGoPackage string
		if err := filepath.Walk(inDir, func(protoFile string, info os.FileInfo, err error) error {
			if !strings.HasSuffix(protoFile, ".proto") {
				return nil
			}
			goPkg, err := detectGoPackageForFile(protoFile)
			if err != nil {
				return err
			}
			if projectGoPackage == "" {
				projectGoPackage = goPkg
			} else {
				if projectGoPackage != goPkg {
					log.Warnf("file %v does not contain "+
						"expected go_package %v (%v instead)", protoFile, projectGoPackage, goPkg)
				}
			}

			imports, err := importsForProtoFile(absoluteRoot, protoFile)
			if err != nil {
				return err
			}
			imports = append([]string{inDir}, imports...)

			if err := writeDescriptors(protoFile, tmpFile.Name(), imports, *compileProtos); err != nil {
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

		outDir := filepath.Join(GopathSrc(), projectGoPackage)

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

func GopathSrc() string {
	return filepath.Join(os.Getenv("GOPATH"), "src")
}

var commonImports = []string{
	GopathSrc(),
	GopathSrc() + "/github.com/solo-io/solo-kit/api/external",
}

var protoImportStatementRegex = regexp.MustCompile(`.*import "(.*)";.*`)

var goPackageStatementRegex = regexp.MustCompile(`option go_package = "(.*)";`)

func detectImportsForFile(file string) ([]string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	var protoImports []string
	for _, line := range lines {
		importStatement := protoImportStatementRegex.FindStringSubmatch(line)
		if len(importStatement) == 0 {
			continue
		}
		if len(importStatement) != 2 {
			return nil, errors.Errorf("parsing import line error: from %v found %v", line, importStatement)
		}
		protoImports = append(protoImports, importStatement[1])
	}
	return protoImports, nil
}

func detectGoPackageForFile(file string) (string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		goPackage := goPackageStatementRegex.FindStringSubmatch(line)
		if len(goPackage) == 0 {
			continue
		}
		if len(goPackage) != 2 {
			return "", errors.Errorf("parsing go_package error: from %v found %v", line, goPackage)
		}
		return goPackage[1], nil
	}
	return "", errors.Errorf("no go_package statement found in file %v", file)
}

func findImportRelativeToRoot(absoluteRoot, importedProtoFile string, customImports, existingImports []string) (string, error) {
	// if the file is already imported, point to that import
	for _, importPath := range existingImports {
		if _, err := os.Stat(filepath.Join(importPath, importedProtoFile)); err == nil {
			return importPath, nil
		}
	}
	rootsToTry := []string{absoluteRoot}

	for _, customImport := range customImports {
		absoluteCustomImport, err := filepath.Abs(customImport)
		if err != nil {
			return "", err
		}
		rootsToTry = append(rootsToTry, absoluteCustomImport)
	}

	var possibleImportPaths []string
	for _, root := range rootsToTry {
		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, importedProtoFile) {
				importPath := strings.TrimSuffix(path, importedProtoFile)
				possibleImportPaths = append(possibleImportPaths, importPath)
			}
			return nil
		}); err != nil {
			return "", err
		}
	}
	if len(possibleImportPaths) == 0 {
		return "", errors.Errorf("found no possible import paths in root directory %v for import %v",
			absoluteRoot, importedProtoFile)
	}
	if len(possibleImportPaths) != 1 {
		log.Warnf("found more than one possible import path in root directory for "+
			"import %v: %v",
			importedProtoFile, possibleImportPaths)
	}
	return possibleImportPaths[0], nil

}

func importsForProtoFile(absoluteRoot, protoFile string) ([]string, error) {
	importStatements, err := detectImportsForFile(protoFile)
	if err != nil {
		return nil, err
	}
	importsForProto := append([]string{}, commonImports...)
	for _, importedProto := range importStatements {
		importPath, err := findImportRelativeToRoot(absoluteRoot, importedProto, customImports, importsForProto)
		if err != nil {
			return nil, err
		}
		importsForProto = append(importsForProto, strings.TrimSuffix(importPath, "/"))
	}
	importsForProto = stringutils.Unique(importsForProto)

	return importsForProto, nil
}

func writeDescriptors(protoFile, toFile string, imports []string, compileProtos bool) error {
	cmd := exec.Command("protoc")
	for i := range imports {
		imports[i] = "-I" + imports[i]
	}
	cmd.Args = append(cmd.Args, imports...)

	if compileProtos {
		cmd.Args = append(cmd.Args,
			"--gogo_out="+
				"Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,"+
				"Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types"+
				":"+GopathSrc())
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
