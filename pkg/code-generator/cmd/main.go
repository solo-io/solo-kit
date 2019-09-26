package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/gogo/protobuf/proto"
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
	"golang.org/x/sync/errgroup"
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
	RelativeRoot string
	// compile protos found in project directories (dirs with solo-kit.json) and their subdirs
	CompileProtos bool
	// compile protos found in these directories. can also point directly to .proto files
	CustomCompileProtos []string
	GenDocs             *DocsOptions
	CustomImports       []string
	SkipDirs            []string
	// arguments for gogo_out=
	CustomGogoOutArgs []string
	// skip generated mocks
	SkipGenMocks bool
	// skip generated tests
	SkipGeneratedTests bool
}

type DescriptorWithPath struct {
	*descriptor.FileDescriptorProto

	ProtoFilePath string
}

func Generate(opts GenerateOptions) error {
	relativeRoot := opts.RelativeRoot
	compileProtos := opts.CompileProtos
	genDocs := opts.GenDocs
	customImports := opts.CustomImports
	customGogoArgs := opts.CustomGogoOutArgs
	skipDirs := opts.SkipDirs
	skipDirs = append(skipDirs, "vendor/")

	var customCompilePrefixes []string
	for _, relativePath := range opts.CustomCompileProtos {
		abs, err := filepath.Abs(relativePath)
		if err != nil {
			return err
		}
		customCompilePrefixes = append(customCompilePrefixes, abs)
	}

	absoluteRoot, err := filepath.Abs(relativeRoot)
	if err != nil {
		return err
	}

	// Creates a ProjectConfig from each of the 'solo-kit.json' files
	// found in the directory tree rooted at 'absoluteRoot'.
	projectConfigs, err := collectProjectsFromRoot(absoluteRoot, skipDirs)
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
		for _, customCompilePrefix := range customCompilePrefixes {
			if strings.HasPrefix(protoFile, customCompilePrefix) {
				return true
			}
		}
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
	descriptors, err := collectDescriptorsFromRoot(absoluteRoot, customImports, customGogoArgs, skipDirs, compileProto)
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

	projectMap, err := parser.ProcessDescriptorsFromConfigs(projectConfigs, protoDescriptors)
	if err != nil {
		return err
	}

	for _, project := range projectMap {
		code, err := codegen.GenerateFiles(project, true, opts.SkipGeneratedTests)
		if err != nil {
			return err
		}

		if err := docgen.WritePerProjectsDocs(project, genDocs, absoluteRoot); err != nil {
			return err
		}

		outDir := filepath.Join(gopathSrc(), project.ProjectConfig.GoPackage)

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
	if err := docgen.WriteCrossProjectDocs(genDocs, absoluteRoot, projectMap); err != nil {
		return err
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

func gopathSrc() string {
	return filepath.Join(os.Getenv("GOPATH"), "src")
}

func collectProjectsFromRoot(root string, skipDirs []string) ([]*model.ProjectConfig, error) {
	var projects []*model.ProjectConfig

	if err := filepath.Walk(root, func(projectFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(projectFile, model.ProjectConfigFilename) {
			return nil
		}
		for _, skip := range skipDirs {
			skipRoot, err := filepath.Abs(skip)
			if err != nil {
				return err
			}
			if strings.HasPrefix(projectFile, skipRoot) {
				log.Warnf("skipping detected project %v", projectFile)
				return nil
			}
		}

		project, err := model.LoadProjectConfig(projectFile)
		if err != nil {
			return err
		}
		projects = append(projects, &project)
		return nil
	}); err != nil {
		return nil, err
	}
	return projects, nil
}

func addDescriptorsForFile(addDescriptor func(f DescriptorWithPath), root, protoFile string, customImports, customGogoArgs []string, wantCompile func(string) bool) error {
	log.Printf("processing proto file input %v", protoFile)
	imports, err := importsForProtoFile(root, protoFile, customImports)
	if err != nil {
		return errors.Wrapf(err, "reading imports for proto file")
	}
	imports = stringutils.Unique(imports)

	// don't generate protos for non-project files
	compile := wantCompile(protoFile)

	// use a temp file to store the output from protoc, then parse it right back in
	// this is how we "wrap" protoc
	tmpFile, err := ioutil.TempFile("", "solo-kit-gen-")
	if err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if err := writeDescriptors(protoFile, tmpFile.Name(), imports, customGogoArgs, compile); err != nil {
		return errors.Wrapf(err, "writing descriptors")
	}
	desc, err := readDescriptors(tmpFile.Name())
	if err != nil {
		return errors.Wrapf(err, "reading descriptors")
	}

	for _, f := range desc.File {
		descriptorWithPath := DescriptorWithPath{FileDescriptorProto: f}
		if strings.HasSuffix(protoFile, f.GetName()) {
			descriptorWithPath.ProtoFilePath = protoFile
		}
		addDescriptor(descriptorWithPath)
	}

	return nil
}

func collectDescriptorsFromRoot(root string, customImports, customGogoArgs, skipDirs []string, wantCompile func(string) bool) ([]*DescriptorWithPath, error) {
	var descriptors []*DescriptorWithPath
	var mutex sync.Mutex
	addDescriptor := func(f DescriptorWithPath) {
		mutex.Lock()
		defer mutex.Unlock()
		// don't add the same proto twice, this avoids the issue where a dependency is imported multiple times
		// with different import paths
		for _, existing := range descriptors {
			existingCopy := proto.Clone(existing.FileDescriptorProto).(*descriptor.FileDescriptorProto)
			existingCopy.Name = f.Name
			if proto.Equal(existingCopy, f.FileDescriptorProto) {
				// if this proto file first came in as an import, but later as a solo-kit project proto,
				// ensure the original proto gets updated with the correct proto file path
				if existing.ProtoFilePath == "" {
					existing.ProtoFilePath = f.ProtoFilePath
				}
				return
			}
		}
		descriptors = append(descriptors, &f)
	}
	var g errgroup.Group
	for _, dir := range append([]string{root}, customImports...) {
		absoluteDir, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		walkErr := filepath.Walk(absoluteDir, func(protoFile string, info os.FileInfo, err error) error {
			if !strings.HasSuffix(protoFile, ".proto") {
				return nil
			}
			for _, skip := range skipDirs {
				skipRoot := filepath.Join(absoluteDir, skip)
				if strings.HasPrefix(protoFile, skipRoot) {
					log.Warnf("skipping proto %v because it is %v is a skipped directory", protoFile, skipRoot)
					return nil
				}
			}

			// parallelize parsing the descriptors as each one requires file i/o and is slow
			g.Go(func() error {
				return addDescriptorsForFile(addDescriptor, absoluteDir, protoFile, customImports, customGogoArgs, wantCompile)
			})
			return nil
		})
		if walkErr != nil {
			return nil, walkErr
		}

		// Wait for all descriptor parsing to complete.
		if err := g.Wait(); err != nil {
			return nil, err
		}
	}
	sort.SliceStable(descriptors, func(i, j int) bool {
		return descriptors[i].GetName() < descriptors[j].GetName()
	})
	return descriptors, nil
}

var protoImportStatementRegex = regexp.MustCompile(`.*import "(.*)";.*`)

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

var commonImports = []string{
	gopathSrc(),
	filepath.Join(gopathSrc(), "github.com", "solo-io", "solo-kit", "api", "external"),
}

func importsForProtoFile(absoluteRoot, protoFile string, customImports []string) ([]string, error) {
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
		dependency := filepath.Join(importPath, importedProto)
		dependencyImports, err := importsForProtoFile(absoluteRoot, dependency, customImports)
		if err != nil {
			return nil, errors.Wrapf(err, "getting imports for dependency")
		}
		importsForProto = append(importsForProto, strings.TrimSuffix(importPath, "/"))
		importsForProto = append(importsForProto, dependencyImports...)
	}

	return importsForProto, nil
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

var defaultGogoArgs = []string{
	"plugins=grpc",
	"Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor",
	"Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types",
	"Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types",
	"Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types",
	"Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types",
	"Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types",
	"Menvoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2",
}

func writeDescriptors(protoFile, toFile string, imports, gogoArgs []string, compileProtos bool) error {
	cmd := exec.Command("protoc")
	for i := range imports {
		imports[i] = "-I" + imports[i]
	}
	cmd.Args = append(cmd.Args, imports...)

	gogoArgs = append(defaultGogoArgs, gogoArgs...)

	if compileProtos {
		cmd.Args = append(cmd.Args,
			"--gogo_out="+strings.Join(gogoArgs, ",")+":"+gopathSrc())
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
		return nil, errors.Wrapf(err, "reading file")
	}
	if err := proto.Unmarshal(protoBytes, &desc); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling tmp file as descriptors")
	}
	return &desc, nil
}

func importCustomResources(imports []string) ([]model.CustomResourceConfig, error) {
	var results []model.CustomResourceConfig
	for _, imp := range imports {
		imp = filepath.Join(gopathSrc(), imp)
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
