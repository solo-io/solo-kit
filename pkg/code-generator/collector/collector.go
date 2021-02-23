package collector

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Collector interface {
	CollectDescriptorsFromRoot(root string, skipDirs []string) ([]*model.DescriptorWithPath, error)
}

func NewCollector(customImports, commonImports, customGogoArgs, customPlugins []string,
	descriptorOutDir string, wantCompile func(string) bool) *collector {
	return &collector{
		descriptorOutDir: descriptorOutDir,
		customImports:    customImports,
		commonImports:    commonImports,
		customGogoArgs:   customGogoArgs,
		wantCompile:      wantCompile,
		customPlugins:    customPlugins,
	}
}

type collector struct {
	descriptorOutDir string
	customImports    []string
	commonImports    []string
	customGogoArgs   []string
	wantCompile      func(string) bool
	customPlugins    []string
}

func (c *collector) CollectDescriptorsFromRoot(root string, skipDirs []string) ([]*model.DescriptorWithPath, error) {
	var descriptors []*model.DescriptorWithPath
	var mutex sync.Mutex
	addDescriptor := func(f model.DescriptorWithPath) {
		mutex.Lock()
		defer mutex.Unlock()
		descriptors = append(descriptors, &f)
	}
	var g errgroup.Group
	for _, dir := range append([]string{root}) {
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
				return c.addDescriptorsForFile(addDescriptor, absoluteDir, protoFile)
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
func (c *collector) addDescriptorsForFile(addDescriptor func(f model.DescriptorWithPath), root, protoFile string) error {
	log.Printf("processing proto file input %v", protoFile)
	imports, err := c.importsForProtoFile(root, protoFile, c.customImports)
	if err != nil {
		return errors.Wrapf(err, "reading imports for proto file")
	}
	imports = stringutils.Unique(imports)

	// don't generate protos for non-project files
	compile := c.wantCompile(protoFile)

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

	if err := c.writeDescriptors(protoFile, tmpFile.Name(), imports, compile); err != nil {
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

var protoImportStatementRegex = regexp.MustCompile(`.*import "(.*)";.*`)

func (c *collector) detectImportsForFile(file string) ([]string, error) {
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

func (c *collector) importsForProtoFile(absoluteRoot, protoFile string, customImports []string) ([]string, error) {
	importStatements, err := c.detectImportsForFile(protoFile)
	if err != nil {
		return nil, err
	}
	importsForProto := append([]string{}, c.commonImports...)
	for _, importedProto := range importStatements {
		importPath, err := c.findImportRelativeToRoot(absoluteRoot, importedProto, customImports, importsForProto)
		if err != nil {
			return nil, err
		}
		dependency := filepath.Join(importPath, importedProto)
		dependencyImports, err := c.importsForProtoFile(absoluteRoot, dependency, customImports)
		if err != nil {
			return nil, errors.Wrapf(err, "getting imports for dependency")
		}
		importsForProto = append(importsForProto, strings.TrimSuffix(importPath, "/"))
		importsForProto = append(importsForProto, dependencyImports...)
	}

	return importsForProto, nil
}

func (c *collector) findImportRelativeToRoot(absoluteRoot, importedProtoFile string, customImports, existingImports []string) (string, error) {
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
		// Try the more specific custom imports first, rather than trying all of vendor
		rootsToTry = append([]string{absoluteCustomImport}, rootsToTry...)
	}

	// Sort by length, so longer (more specific paths are attempted first)
	sort.Slice(rootsToTry, func(i, j int) bool {
		elementsJ := strings.Split(rootsToTry[j], string(os.PathSeparator))
		elementsI := strings.Split(rootsToTry[i], string(os.PathSeparator))
		return len(elementsI) > len(elementsJ)
	})

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
		// if found break
		if len(possibleImportPaths) > 0 {
			break
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
	"Mgithub.com/solo-io/solo-kit/api/external/envoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2",
	"Mgithub.com/solo-io/solo-kit/api/external/envoy/service/discovery/v3/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3",
}

func (c *collector) writeDescriptors(protoFile, toFile string, imports []string, compileProtos bool) error {
	cmd := exec.Command("protoc")
	for i := range imports {
		imports[i] = "-I" + imports[i]
	}
	cmd.Args = append(cmd.Args, imports...)
	gogoArgs := append(defaultGogoArgs, c.customGogoArgs...)

	if compileProtos {
		cmd.Args = append(cmd.Args,
			"--go_out="+strings.Join(gogoArgs, ",")+":"+c.descriptorOutDir,
			"--ext_out="+strings.Join(gogoArgs, ",")+":"+c.descriptorOutDir,
		)

		for _, plugin := range c.customPlugins {
			cmd.Args = append(cmd.Args,
				"--"+plugin+"_out="+strings.Join(gogoArgs, ",")+":"+c.descriptorOutDir,
			)
		}
	}

	cmd.Args = append(cmd.Args, "-o"+toFile, "--include_imports", "--include_source_info",
		protoFile)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "%v failed: %s", cmd.Args, out)
	}
	return nil
}
