package parser

import (
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
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"golang.org/x/sync/errgroup"
)

type DescriptorWithPath struct {
	*descriptor.FileDescriptorProto

	ProtoFilePath string
}

type Collector interface {
	CollectProjects(skipDirs []string) ([]*model.ProjectConfig, error)
	CollectDescriptors(customImports, customGogoArgs, skipDirs []string, wantCompile func(string) bool) ([]*DescriptorWithPath, error)
}

type collector struct {
	root string
}

func NewCollector(root string) *collector {
	return &collector{root: root}
}

func (c *collector) CollectProjects(skipDirs []string) ([]*model.ProjectConfig, error) {
	var projects []*model.ProjectConfig

	if err := filepath.Walk(c.root, func(projectFile string, info os.FileInfo, err error) error {
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

func (c *collector) CollectDescriptors(customImports, customGogoArgs, skipDirs []string, wantCompile func(string) bool) ([]*DescriptorWithPath, error) {
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
	for _, dir := range append([]string{c.root}, customImports...) {
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
			"--gogo_out="+strings.Join(gogoArgs, ",")+":"+GopathSrc())
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

func GopathSrc() string {
	return filepath.Join(os.Getenv("GOPATH"), "src")
}

var commonImports = []string{
	GopathSrc(),
	filepath.Join(GopathSrc(), "github.com", "solo-io", "solo-kit", "api", "external"),
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
