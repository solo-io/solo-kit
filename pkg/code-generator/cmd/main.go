package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/stringutils"
	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
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
	// Root of files to be compiled (will default to "." if not set)
	RelativeRoot string
	// // Root of package, necessary to find vendor (will default to $(go env GOMOD) if not set)
	// ProjectRoot string
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
	/*
		Represents the go package which this package would have been in the GOPATH
		This allows it to be able to maintain compatility with the old solo-kit

		default: current github.com/solo-io/<current-folder>
		for example: github.com/solo-io/solo-it
	*/
	PackageName string
}

type Runner struct {
	Opts             GenerateOptions
	RelativeRoot     string
	DescriptorOutDir string
	BaseDir          string
	AbsoluteRoot     string
}

func Generate(opts GenerateOptions) error {
	opts.SkipDirs = append(opts.SkipDirs, "vendor/")

	workingRootRelative := opts.RelativeRoot
	if workingRootRelative == "" {
		workingRootRelative = "."
	}

	cmd := exec.Command("go", "env", "GOMOD")
	modBytes, err := cmd.Output()
	if err != nil {
		return err
	}
	modFileString := strings.TrimSpace(string(modBytes))
	modPackageName, err := getModPackageName(modFileString)
	if err != nil {
		return err
	}
	modPathString := filepath.Dir(modFileString)

	if opts.PackageName == "" {
		opts.PackageName = modPackageName
	}

	descriptorOutDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.Remove(descriptorOutDir)

	// copy over our protos to right path
	r := Runner{
		RelativeRoot:     workingRootRelative,
		Opts:             opts,
		BaseDir:          modPathString,
		DescriptorOutDir: descriptorOutDir,
	}

	// copy out generated code
	err = r.Run()
	if err != nil {
		return err
	}

	if err := filepath.Walk(filepath.Join(descriptorOutDir, r.Opts.PackageName), func(pbgoFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !(strings.HasSuffix(pbgoFile, ".pb.go") || strings.HasSuffix(pbgoFile, ".pb.hash.go")) {
			return nil
		}

		dest := strings.TrimPrefix(pbgoFile, filepath.Join(descriptorOutDir, r.Opts.PackageName))
		dest = strings.TrimPrefix(dest, "/")
		// dest = filepath.Join(relativeRoot, dest)
		dir, _ := filepath.Split(dest)
		os.MkdirAll(dir, 0755)

		// copy
		srcFile, err := os.Open(pbgoFile)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		log.Printf("copying %v -> %v", pbgoFile, dest)
		_, err = io.Copy(dstFile, srcFile)
		return err

	}); err != nil {
		return err
	}

	return nil
}

func (r *Runner) Run() error {
	workingRootAbsolute, err := filepath.Abs(r.RelativeRoot)
	if err != nil {
		return err
	}
	// Creates a ProjectConfig from each of the 'solo-kit.json' files
	// found in the directory tree rooted at 'workingRootAbsolute'.
	projectConfigs, err := r.collectProjectsFromRoot(workingRootAbsolute, r.Opts.SkipDirs)
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

	var customCompilePrefixes []string
	for _, relativePath := range r.Opts.CustomCompileProtos {
		abs, err := filepath.Abs(relativePath)
		if err != nil {
			return err
		}
		customCompilePrefixes = append(customCompilePrefixes, abs)
	}

	// whether or not to do a regular gogo-proto generate while collecting descriptors
	compileProto := func(protoFile string) bool {
		for _, customCompilePrefix := range customCompilePrefixes {
			if strings.HasPrefix(protoFile, customCompilePrefix) {
				return true
			}
		}
		if !r.Opts.CompileProtos {
			return false
		}
		for _, proj := range projectConfigs {
			if strings.HasPrefix(protoFile, filepath.Dir(proj.ProjectFile)) {
				return true
			}
		}
		return false
	}

	// Create a FileDescriptorProto for all the proto files under 'workingRootAbsolute' and each of the 'customImports' paths
	descriptors, err := r.collectDescriptorsFromRoot(workingRootAbsolute, r.Opts.CustomImports,
		r.Opts.CustomGogoOutArgs, r.Opts.SkipDirs, compileProto)
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
		importedResources, err := r.importCustomResources(projectConfig.Imports)
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
		code, err := codegen.GenerateFiles(project, true, r.Opts.SkipGeneratedTests, project.ProjectConfig.GenKubeTypes)
		if err != nil {
			return err
		}

		if err := docgen.WritePerProjectsDocs(project, r.Opts.GenDocs, workingRootAbsolute); err != nil {
			return err
		}

		split := strings.SplitAfterN(project.ProjectConfig.GoPackage, "/", filepathValidLength)
		if len(split) < filepathValidLength {
			return errors.Errorf("projectConfig.GoPackage is not valid, %s", project.ProjectConfig.GoPackage)
		}
		outDir := split[filepathValidLength-1]

		for _, file := range code {
			path := filepath.Join(outDir, file.Filename)
			if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
				return err
			}
			if err := ioutil.WriteFile(path, []byte(file.Content), 0644); err != nil {
				return err
			}
			if !strings.HasSuffix(file.Filename, ".go") {
				continue
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
		if !r.Opts.SkipGenMocks {
			if err := genMocks(code, outDir, workingRootAbsolute); err != nil {
				return err
			}
		}
	}
	if err := docgen.WriteCrossProjectDocs(r.Opts.GenDocs, workingRootAbsolute, projectMap); err != nil {
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

func (r *Runner) collectProjectsFromRoot(root string, skipDirs []string) ([]*model.ProjectConfig, error) {
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

func (r *Runner) addDescriptorsForFile(addDescriptor func(f model.DescriptorWithPath), root, protoFile string, customImports, customGogoArgs []string, wantCompile func(string) bool) error {
	log.Printf("processing proto file input %v", protoFile)
	imports, err := r.importsForProtoFile(root, protoFile, customImports)
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

	if err := r.writeDescriptors(protoFile, tmpFile.Name(), imports, customGogoArgs, compile); err != nil {
		return errors.Wrapf(err, "writing descriptors")
	}
	desc, err := readDescriptors(tmpFile.Name())
	if err != nil {
		return errors.Wrapf(err, "reading descriptors")
	}

	for _, f := range desc.File {
		descriptorWithPath := model.DescriptorWithPath{FileDescriptorProto: f}
		if strings.HasSuffix(protoFile, f.GetName()) {
			descriptorWithPath.ProtoFilePath = protoFile
		}
		addDescriptor(descriptorWithPath)
	}

	return nil
}

func (r *Runner) collectDescriptorsFromRoot(root string, customImports, customGogoArgs, skipDirs []string, wantCompile func(string) bool) ([]*model.DescriptorWithPath, error) {
	var descriptors []*model.DescriptorWithPath
	var mutex sync.Mutex
	addDescriptor := func(f model.DescriptorWithPath) {
		mutex.Lock()
		defer mutex.Unlock()
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
				return r.addDescriptorsForFile(addDescriptor, absoluteDir, protoFile, customImports, customGogoArgs, wantCompile)
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

	// don't add the same proto twice, this avoids the issue where a dependency is imported multiple times
	// with different import paths
	return parser.FilterDuplicateDescriptors(descriptors), nil
}

var protoImportStatementRegex = regexp.MustCompile(`.*import "(.*)";.*`)

func (r *Runner) detectImportsForFile(file string) ([]string, error) {
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

func getCommonImports() ([]string, error) {
	var result []string
	for _, v := range commonImportStrings {
		if abs, err := filepath.Abs(v); err != nil {
			return nil, err
		} else {
			result = append(result, abs)
		}
	}
	return result, nil
}

var commonImportStrings = []string{
	".",
	"./api",
}

func (r *Runner) importsForProtoFile(absoluteRoot, protoFile string, customImports []string) ([]string, error) {
	importStatements, err := r.detectImportsForFile(protoFile)
	if err != nil {
		return nil, err
	}
	commonImports, err := getCommonImports()
	if err != nil {
		return nil, err
	}
	importsForProto := append([]string{}, commonImports...)
	for _, importedProto := range importStatements {
		importPath, err := r.findImportRelativeToRoot(absoluteRoot, importedProto, customImports, importsForProto)
		if err != nil {
			return nil, err
		}
		dependency := filepath.Join(importPath, importedProto)
		dependencyImports, err := r.importsForProtoFile(absoluteRoot, dependency, customImports)
		if err != nil {
			return nil, errors.Wrapf(err, "getting imports for dependency")
		}
		importsForProto = append(importsForProto, strings.TrimSuffix(importPath, "/"))
		importsForProto = append(importsForProto, dependencyImports...)
	}

	return importsForProto, nil
}

func (r *Runner) findImportRelativeToRoot(absoluteRoot, importedProtoFile string, customImports, existingImports []string) (string, error) {
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
	"Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types",
	"Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types",
	"Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types",
	"Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types",
	"Menvoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2",
}

func (r *Runner) writeDescriptors(protoFile, toFile string, imports, gogoArgs []string, compileProtos bool) error {
	cmd := exec.Command("protoc")
	for i := range imports {
		imports[i] = "-I" + imports[i]
	}
	cmd.Args = append(cmd.Args, imports...)

	gogoArgs = append(defaultGogoArgs, gogoArgs...)

	if compileProtos {
		cmd.Args = append(cmd.Args,
			"--gogo_out="+strings.Join(gogoArgs, ",")+":"+r.DescriptorOutDir,
			"--ext_out="+strings.Join(gogoArgs, ",")+":"+r.DescriptorOutDir,
		)
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

const (
	filepathValidLength      = 4
	filepathWithVendorLength = filepathValidLength + 1
)

func (r *Runner) importCustomResources(imports []string) ([]model.CustomResourceConfig, error) {
	var results []model.CustomResourceConfig
	for _, imp := range imports {
		imp = filepath.Join("vendor", imp)
		if !strings.HasSuffix(imp, model.ProjectConfigFilename) {
			imp = filepath.Join(imp, model.ProjectConfigFilename)
		}
		byt, err := ioutil.ReadFile(imp)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			/*
				used to split file name up if check in vendor fails.
				for example: vendor/github.com/solo-io/solo-kit/api/external/kubernetes/solo-kit.json
				will become: [vendor/, github.com/, solo-io/, solo-kit/, api/external/kubernetes/solo-kit.json]
				and the final member is the local path
			*/
			split := strings.SplitAfterN(imp, "/", filepathWithVendorLength)
			if len(split) < filepathWithVendorLength {
				return nil, errors.Errorf("filepath is not valid, %s", imp)
			}
			byt, err = ioutil.ReadFile(split[filepathWithVendorLength-1])
			if err != nil {
				return nil, err
			}
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

func getModPackageName(module string) (string, error) {
	f, err := os.Open(module)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	if !scanner.Scan() {
		return "", fmt.Errorf("invalid module file")
	}
	line := scanner.Text()
	parts := strings.Split(line, " ")

	modPath := parts[len(parts)-1]
	if modPath == "/dev/null" || modPath == "" {
		return "", errors.New("solo-kit must be run from within go.mod repo")
	}

	return parts[len(parts)-1], nil
}
