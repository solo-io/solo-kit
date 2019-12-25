package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/stringutils"
	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/utils/modutils"
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

type RunFunc func() error

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
		for example: github.com/solo-io/solo-kit
	*/
	PackageName string

	PreRunFuncs []RunFunc
}

type Runner struct {
	Opts GenerateOptions
	// Relative root of solo-kit gen. Will be used as the root of all generation
	RelativeRoot string
	// Location to output all proto code gen, defaults to a temp dir
	DescriptorOutDir string
	// root of the go mod package
	BaseDir string
	// common import directories in which solo-kit should look for protos in the current package
	CommonImports []string
}

func Generate(opts GenerateOptions) error {
	for _, preRun := range opts.PreRunFuncs {
		if err := preRun(); err != nil {
			return err
		}
	}

	// opts.SkipDirs = append(opts.SkipDirs, "vendor/")

	workingRootRelative := opts.RelativeRoot
	if workingRootRelative == "" {
		workingRootRelative = "."
	}

	modBytes, err := modutils.GetCurrentModPackageFile()
	modFileString := strings.TrimSpace(string(modBytes))
	modPackageName, err := modutils.GetCurrentModPackageName(modFileString)
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
	commonImports, err := getCommonImports()
	if err != nil {
		return err
	}

	// copy over our protos to right path
	r := Runner{
		RelativeRoot:     workingRootRelative,
		Opts:             opts,
		BaseDir:          modPathString,
		DescriptorOutDir: descriptorOutDir,
		CommonImports:    commonImports,
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
	projectConfigRoot := filepath.Join(r.BaseDir, "vendor", r.Opts.PackageName)
	projectConfigs, err := r.collectProjectsFromRoot(projectConfigRoot, r.Opts.SkipDirs)
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

	descriptorCollector := collector.NewCollector(r.Opts.CustomImports, r.CommonImports,
		r.Opts.CustomGogoOutArgs, r.DescriptorOutDir, compileProto)

	descriptors, err := descriptorCollector.CollectDescriptorsFromRoot(filepath.Join(r.BaseDir, "vendor"), r.Opts.SkipDirs)
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

			switch {
			case strings.HasSuffix(file.Filename, ".sh"):
				if out, err := exec.Command("chmod", "+x", filepath.Join(workingRootAbsolute, path)).CombinedOutput(); err != nil {
					return errors.Wrapf(err, "chmod failed: %s", out)
				}

			case strings.HasSuffix(file.Filename, ".go"):
				if out, err := exec.Command("gofmt", "-w", path).CombinedOutput(); err != nil {
					return errors.Wrapf(err, "gofmt failed: %s", out)
				}

				if out, err := exec.Command("goimports", "-w", path).CombinedOutput(); err != nil {
					return errors.Wrapf(err, "goimports failed: %s", out)
				}
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

func getCommonImports() ([]string, error) {
	var result []string
	modPackageFile, err := modutils.GetCurrentModPackageFile()
	if err != nil {
		return nil, err
	}
	modPackageDir := filepath.Dir(modPackageFile)
	for _, v := range commonImportStrings {
		result = append(result, filepath.Join(modPackageDir, v))
	}
	return result, nil
}

var commonImportStrings = []string{
	"vendor",
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
