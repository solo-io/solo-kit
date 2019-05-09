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
type SoloKitContext struct {
	CompileProtos bool
	GenDocs       *DocsOptions
	SkipDirs      []string
	CustomImports []string
}

const (
	SkipMockGen = "SKIP_MOCK_GEN"
)

func Run(relativeRoot string, compileProtos bool, genDocs *DocsOptions, customImports, skipDirs []string) error {

	skctx := SoloKitContext{
		CompileProtos: compileProtos,
		GenDocs:       genDocs,
		SkipDirs:      skipDirs,
		CustomImports: customImports,
	}

	cmd := exec.Command("go", "env", "GOMOD")
	modBytes, err := cmd.Output()
	if err != nil {
		return err
	}
	module := strings.TrimSpace(string(modBytes))
	if module == "" {
		return RunGoPath(relativeRoot, skctx)
	}

	if len(customImports) != 0 {
		return fmt.Errorf("custom imports not supported in module mode. please vendor your protos")
	}

	return RunModules(module, relativeRoot, skctx)

}

func getModPath(module string) (string, error) {

	f, _ := os.Open(module)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	if !scanner.Scan() {
		return "", fmt.Errorf("invalid module file")
	}
	line := scanner.Text()
	parts := strings.Split(line, " ")

	return parts[len(parts)-1], nil

}

func RunModules(module string, relativeRoot string, skctx SoloKitContext) error {
	// vendor our protos

	modulePath, err := getModPath(module)
	if err != nil {
		return err
	}

	directory, _ := filepath.Split(module)

	absoluteVendor, err := filepath.Abs(filepath.Join(directory, "vendor"))
	if err != nil {
		return err
	}

	projecteRoot := filepath.Join(absoluteVendor, modulePath, relativeRoot)

	// copy over our protos to right path
	r := Runner{
		RelativeRoot:     relativeRoot,
		SoloKitContext:   skctx,
		BaseOutDir:       absoluteVendor,
		DescriptorOutDir: absoluteVendor,
		CommonImports: []string{
			absoluteVendor,
		},
		AbsoluteRoot: absoluteVendor,
		ProjectRoot:  projecteRoot,
	}
	// copy out generated code
	err = r.Run()
	if err != nil {
		return err
	}
	// copy protos back

	if err := filepath.Walk(projecteRoot, func(pbgoFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(pbgoFile, ".pb.go") {
			return nil
		}

		dest := strings.TrimPrefix(pbgoFile, projecteRoot)
		dest = strings.TrimPrefix(dest, "/")
		dest = filepath.Join(relativeRoot, dest)
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

func RunGoPath(relativeRoot string, skctx SoloKitContext) error {

	absoluteRoot, err := filepath.Abs(relativeRoot)
	if err != nil {
		return err
	}
	gopathSrc := filepath.Join(os.Getenv("GOPATH"), "src")

	r := Runner{
		RelativeRoot:     relativeRoot,
		SoloKitContext:   skctx,
		DescriptorOutDir: gopathSrc,
		BaseOutDir:       gopathSrc,
		CommonImports: []string{
			gopathSrc,
			filepath.Join(gopathSrc, "github.com", "solo-io", "solo-kit", "api", "external"),
		},
		AbsoluteRoot: absoluteRoot,
		ProjectRoot:  absoluteRoot,
	}

	r.SkipDirs = append(r.SkipDirs, "vendor/")
	return r.Run()
}

type Runner struct {
	SoloKitContext
	CommonImports    []string
	RelativeRoot     string
	DescriptorOutDir string
	BaseOutDir       string
	AbsoluteRoot     string
	ProjectRoot      string
}

func (r *Runner) Run() error {

	// Creates a ProjectConfig from each of the 'solo-kit.json' files
	// found in the directory tree rooted at 'absoluteRoot'.
	projectConfigs, err := r.collectProjectsFromRoot(r.ProjectRoot, r.SkipDirs)
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
		if !r.CompileProtos {
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
	descriptors, err := r.collectDescriptorsFromRoot(compileProto)
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

	for _, projectConfig := range projectConfigs {
		importedResources, err := r.importCustomResources(projectConfig.Imports)
		if err != nil {
			return err
		}

		projectConfig.CustomResources = append(projectConfig.CustomResources, importedResources...)

		// Build a 'Project' object that contains a resource for each message that:
		// - is contained in the FileDescriptor and
		// - is a solo kit resource (i.e. it has a field named 'metadata')
		project, err := parser.ProcessDescriptors(projectConfig, descriptors)
		if err != nil {
			return err
		}

		code, err := codegen.GenerateFiles(project, true)
		if err != nil {
			return err
		}

		if project.ProjectConfig.DocsDir != "" && (r.GenDocs != nil) {
			docs, err := docgen.GenerateFiles(project, r.GenDocs)
			if err != nil {
				return err
			}

			for _, file := range docs {
				path := filepath.Join(r.ProjectRoot, project.ProjectConfig.DocsDir, file.Filename)
				if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
					return err
				}
				if err := ioutil.WriteFile(path, []byte(file.Content), 0644); err != nil {
					return err
				}
			}
		}

		outDir := filepath.Join(r.BaseOutDir, project.ProjectConfig.GoPackage)

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
		if os.Getenv(SkipMockGen) != "1" {
			if err := genMocks(code, outDir, r.AbsoluteRoot); err != nil {
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
	if strings.Contains(file.Filename, "_test") || !containsAny(file.Filename, validMockingInterfaces) {
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

func (r *Runner) collectProjectsFromRoot(root string, skipDirs []string) ([]model.ProjectConfig, error) {
	var projects []model.ProjectConfig

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
		projects = append(projects, project)
		return nil
	}); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *Runner) addDescriptorsForFile(addDescriptor func(f *descriptor.FileDescriptorProto), root, protoFile string, wantCompile func(string) bool) error {

	// TODO(yuval-k): if in go mod mode, copy proto into right place in vendor folder
	// vendor/removePrefix(protoFiles,root)

	log.Printf("processing proto file input %v", protoFile)
	imports, err := r.importsForProtoFile(protoFile)
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

	if err := r.writeDescriptors(protoFile, tmpFile.Name(), imports, compile); err != nil {
		return errors.Wrapf(err, "writing descriptors")
	}
	desc, err := r.readDescriptors(tmpFile.Name())
	if err != nil {
		return errors.Wrapf(err, "reading descriptors")
	}

	for _, f := range desc.File {
		addDescriptor(f)
	}

	return nil
}

func (r *Runner) collectDescriptorsFromRoot(wantCompile func(string) bool) ([]*descriptor.FileDescriptorProto, error) {
	var descriptors []*descriptor.FileDescriptorProto
	var mutex sync.Mutex
	addDescriptor := func(f *descriptor.FileDescriptorProto) {
		mutex.Lock()
		defer mutex.Unlock()
		// don't add the same proto twice, this avoids the issue where a dependency is imported multiple times
		// with different import paths
		for _, existing := range descriptors {
			if existing.GetName() == f.GetName() {
				return
			}
			existingCopy := proto.Clone(existing).(*descriptor.FileDescriptorProto)
			existingCopy.Name = f.Name
			if proto.Equal(existingCopy, f) {
				return
			}
		}
		descriptors = append(descriptors, f)
	}
	var g errgroup.Group
	for _, dir := range append([]string{r.ProjectRoot}, r.CustomImports...) {
		absoluteDir, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		walkErr := filepath.Walk(absoluteDir, func(protoFile string, info os.FileInfo, err error) error {
			if !strings.HasSuffix(protoFile, ".proto") {
				return nil
			}
			for _, skip := range r.SkipDirs {
				skipRoot := filepath.Join(absoluteDir, skip)
				if strings.HasPrefix(protoFile, skipRoot) {
					log.Warnf("skipping proto %v because it is %v is a skipped directory", protoFile, skipRoot)
					return nil
				}
			}

			// parallelize parsing the descriptors as each one requires file i/o and is slow
			g.Go(func() error {
				return r.addDescriptorsForFile(addDescriptor, absoluteDir, protoFile, wantCompile)
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

func (r *Runner) importsForProtoFile(protoFile string) ([]string, error) {
	importStatements, err := r.detectImportsForFile(protoFile)
	if err != nil {
		return nil, err
	}
	importsForProto := append([]string{}, r.CommonImports...)
	for _, importedProto := range importStatements {
		importPath, err := r.findImportRelativeToRoot(importedProto, importsForProto)
		if err != nil {
			return nil, err
		}
		dependency := filepath.Join(importPath, importedProto)
		dependencyImports, err := r.importsForProtoFile(dependency)
		if err != nil {
			return nil, errors.Wrapf(err, "getting imports for dependency")
		}
		importsForProto = append(importsForProto, strings.TrimSuffix(importPath, "/"))
		importsForProto = append(importsForProto, dependencyImports...)
	}

	return importsForProto, nil
}

func (r *Runner) findImportRelativeToRoot(importedProtoFile string, existingImports []string) (string, error) {
	// if the file is already imported, point to that import
	for _, importPath := range existingImports {
		if _, err := os.Stat(filepath.Join(importPath, importedProtoFile)); err == nil {
			return importPath, nil
		}
	}
	rootsToTry := []string{r.AbsoluteRoot}

	for _, customImport := range r.CustomImports {
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
			r.AbsoluteRoot, importedProtoFile)
	}
	if len(possibleImportPaths) != 1 {
		log.Warnf("found more than one possible import path in root directory for "+
			"import %v: %v",
			importedProtoFile, possibleImportPaths)
	}
	return possibleImportPaths[0], nil

}

func (r *Runner) writeDescriptors(protoFile, toFile string, imports []string, compileProtos bool) error {
	cmd := exec.Command("protoc")
	for i := range imports {
		imports[i] = "-I" + imports[i]
	}
	cmd.Args = append(cmd.Args, imports...)

	if compileProtos {
		cmd.Args = append(cmd.Args,
			"--gogo_out=plugins=grpc,"+
				"Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor,"+
				"Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,"+
				"Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,"+
				"Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,"+
				"Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,"+
				"Menvoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2"+
				":"+r.DescriptorOutDir)
	}

	cmd.Args = append(cmd.Args, "-o"+toFile, "--include_imports", "--include_source_info",
		protoFile)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "%v failed: %s", cmd.Args, out)
	}
	return nil
}

func (r *Runner) readDescriptors(fromFile string) (*descriptor.FileDescriptorSet, error) {
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

func (r *Runner) importCustomResources(imports []string) ([]model.CustomResourceConfig, error) {
	var results []model.CustomResourceConfig
	for _, imp := range imports {
		imp = filepath.Join(r.BaseOutDir, imp)
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
