package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/stringutils"
	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type DocsOptions = options.DocsOptions

const (
	SkipMockGen = "SKIP_MOCK_GEN"
)

func Run(relativeRoot string, compileProtos bool, genDocs *DocsOptions, customImports, skipDirs []string) error {
	return Generate(GenerateOptions{
		RelativeRoot:  relativeRoot,
		CompileProtos: compileProtos,
		GenDocs:       genDocs,
		CustomImports: customImports,
		SkipDirs:      skipDirs,
		SkipGenMocks:  os.Getenv(SkipMockGen) != "",
	})
}

type GenerateOptions struct {
	RelativeRoot  string
	CompileProtos bool
	GenDocs       *DocsOptions
	CustomImports []string
	SkipDirs      []string
	// arguments for gogo_out=
	CustomGogoOutArgs []string
	// skip generated mocks
	SkipGenMocks bool
	// skip generated tests
	SkipGeneratedTests bool
}

func Generate(opts GenerateOptions) error {
	relativeRoot := opts.RelativeRoot
	compileProtos := opts.CompileProtos
	genDocs := opts.GenDocs
	customImports := opts.CustomImports
	customGogoArgs := opts.CustomGogoOutArgs
	skipDirs := opts.SkipDirs
	skipDirs = append(skipDirs, "vendor/")
	absoluteRoot, err := filepath.Abs(relativeRoot)
	if err != nil {
		return err
	}

	collector := parser.NewCollector(absoluteRoot)

	// Creates a ProjectConfig from each of the 'solo-kit.json' files
	// found in the directory tree rooted at 'absoluteRoot'.
	projectConfigs, err := collector.CollectProjects(skipDirs)
	if err != nil {
		return err
	}

	log.Printf("collected projects: %v", func() []string {
		var names []string
		for _, project := range projectConfigs {
			names = append(names, project.Name)
		}
		sort.Strings(names)
		return names
	}())

	// whether or not to do a regular gogo-proto generate while collecting descriptors
	compileProto := func(protoFile string) bool {
		if !compileProtos {
			return false
		}
		for _, proj := range projectConfigs {
			if strings.HasPrefix(protoFile, filepath.Dir(proj.ProjectFile)) {
				return true
			}
		}
		return false
	}

	// Create a FileDescriptorProto for all the proto files under 'absoluteRoot' and each of the 'customImports' paths
	descriptors, err := collector.CollectDescriptors(customImports, customGogoArgs, skipDirs, compileProto)
	if err != nil {
		return err
	}

	log.Printf("collected descriptors: %v", func() []string {
		var names []string
		for _, desc := range descriptors {
			names = append(names, desc.GetName())
		}
		names = stringutils.Unique(names)
		sort.Strings(names)
		return names
	}())

	var protoDescriptors []*descriptor.FileDescriptorProto
	for _, projectConfig := range projectConfigs {
		importedResources, err := importCustomResources(projectConfig.Imports)
		if err != nil {
			return err
		}

		projectConfig.CustomResources = append(projectConfig.CustomResources, importedResources...)

		for _, desc := range descriptors {
			if filepath.Dir(desc.ProtoFilePath) == filepath.Dir(projectConfig.ProjectFile) {
				projectConfig.ProjectProtos = append(projectConfig.ProjectProtos, desc.GetName())
			}
			protoDescriptors = append(protoDescriptors, desc.FileDescriptorProto)
		}
	}

	for _, projectConfig := range projectConfigs {

		// Build a 'Project' object that contains a resource for each message that:
		// - is contained in the FileDescriptor and
		// - is a solo kit resource (i.e. it has a field named 'metadata')

		project, err := parser.ProcessDescriptors(projectConfig, projectConfigs, protoDescriptors)
		if err != nil {
			return err
		}

		code, err := codegen.GenerateFiles(project, true, opts.SkipGeneratedTests)
		if err != nil {
			return err
		}

		if project.ProjectConfig.DocsDir != "" && (genDocs != nil) {
			docs, err := docgen.GenerateFiles(project, genDocs)
			if err != nil {
				return err
			}

			for _, file := range docs {
				path := filepath.Join(absoluteRoot, project.ProjectConfig.DocsDir, file.Filename)
				if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
					return err
				}
				if err := ioutil.WriteFile(path, []byte(file.Content), 0644); err != nil {
					return err
				}
			}
		}

		outDir := filepath.Join(parser.GopathSrc(), project.ProjectConfig.GoPackage)

		for _, file := range code {
			path := filepath.Join(outDir, file.Filename)
			if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
				return err
			}
			if err := ioutil.WriteFile(path, []byte(file.Content), 0644); err != nil {
				return err
			}
			if out, err := exec.Command("gofmt", "-w", path).CombinedOutput(); err != nil {
				return errors.Wrapf(err, "gofmt failed: %s", out)
			}

			if out, err := exec.Command("goimports", "-w", path).CombinedOutput(); err != nil {
				return errors.Wrapf(err, "goimports failed: %s", out)
			}
		}

		// Generate mocks
		// need to run after to make sure all resources have already been written
		// Set this env var during tests so that mocks are not generated
		if !opts.SkipGenMocks {
			if err := genMocks(code, outDir, absoluteRoot); err != nil {
				return err
			}
		}
	}

	return nil
}

var (
	validMockingInterfaces = []string{
		"_client",
		"_reconciler",
		"_emitter",
		"_event_loop",
	}

	invalidMockingInterface = []string{
		"_simple_event_loop",
		"_test",
	}
)

func genMocks(code code_generator.Files, outDir, absoluteRoot string) error {
	if err := os.MkdirAll(filepath.Join(outDir, "mocks"), 0777); err != nil {
		return err
	}
	for _, file := range code {
		if out, err := genMockForFile(file, outDir, absoluteRoot); err != nil {
			return errors.Wrapf(err, "mockgen failed: %s", out)
		}

	}
	return nil
}

func genMockForFile(file code_generator.File, outDir, absoluteRoot string) ([]byte, error) {
	if containsAny(file.Filename, invalidMockingInterface) || !containsAny(file.Filename, validMockingInterfaces) {
		return nil, nil
	}
	path := filepath.Join(outDir, file.Filename)
	dest := filepath.Join(outDir, "mocks", file.Filename)
	path = strings.Replace(path, absoluteRoot, ".", 1)
	dest = strings.Replace(dest, absoluteRoot, ".", 1)
	return exec.Command("mockgen", fmt.Sprintf("-source=%s", path), fmt.Sprintf("-destination=%s", dest), "-package=mocks").CombinedOutput()
}

func containsAny(str string, slice []string) bool {
	for _, val := range slice {
		if strings.Contains(str, val) {
			return true
		}
	}
	return false
}

func importCustomResources(imports []string) ([]model.CustomResourceConfig, error) {
	var results []model.CustomResourceConfig
	for _, imp := range imports {
		imp = filepath.Join(parser.GopathSrc(), imp)
		if !strings.HasSuffix(imp, model.ProjectConfigFilename) {
			imp = filepath.Join(imp, model.ProjectConfigFilename)
		}
		byt, err := ioutil.ReadFile(imp)
		if err != nil {
			return nil, err
		}
		var projectConfig model.ProjectConfig
		err = json.Unmarshal(byt, &projectConfig)
		if err != nil {
			return nil, err
		}
		var customResources []model.CustomResourceConfig
		for _, v := range projectConfig.CustomResources {
			v.Package = projectConfig.GoPackage
			v.Imported = true
			customResources = append(customResources, v)
		}
		results = append(results, customResources...)
	}

	return results, nil
}
