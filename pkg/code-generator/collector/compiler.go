package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/code-generator/metrics"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type ProtoCompiler interface {
	CompileDescriptorsFromRoot(root string, skipDirs []string) ([]*model.DescriptorWithPath, error)
}

func NewProtoCompiler(
	collector Collector,
	customImports, commonImports, customGoArgs, customPlugins []string,
	descriptorOutDir string, wantCompile func(string) bool) *protoCompiler {
	return &protoCompiler{
		collector:        collector,
		descriptorOutDir: descriptorOutDir,
		customImports:    customImports,
		commonImports:    commonImports,
		customGoArgs:     customGoArgs,
		wantCompile:      wantCompile,
		customPlugins:    customPlugins,
	}
}

type protoCompiler struct {
	collector        Collector
	descriptorOutDir string
	customImports    []string
	commonImports    []string
	customGoArgs     []string
	wantCompile      func(string) bool
	customPlugins    []string
}

func (c *protoCompiler) CompileDescriptorsFromRoot(root string, skipDirs []string) ([]*model.DescriptorWithPath, error) {
	defer metrics.MeasureElapsed("proto-compiler", time.Now())

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
func (c *protoCompiler) addDescriptorsForFile(addDescriptor func(f model.DescriptorWithPath), root, protoFile string) error {
	log.Printf("processing proto file input %v", protoFile)
	imports, err := c.collector.CollectImportsForFile(root, protoFile)
	if err != nil {
		return errors.Wrapf(err, "reading imports for proto file")
	}

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

var defaultGoArgs = []string{
	"plugins=grpc",
	"Mgithub.com/solo-io/solo-kit/api/external/envoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2",
}

func (c *protoCompiler) writeDescriptors(protoFile, toFile string, imports []string, compileProtos bool) error {
	cmd := exec.Command("protoc")
	var cmdImports []string
	for _, i := range imports {
		cmdImports = append(cmdImports, fmt.Sprintf("-I%s", i))
	}
	cmd.Args = append(cmd.Args, cmdImports...)
	goArgs := append(defaultGoArgs, c.customGoArgs...)

	if compileProtos {
		cmd.Args = append(cmd.Args,
			"--go_out="+strings.Join(goArgs, ",")+":"+c.descriptorOutDir,
			"--ext_out="+strings.Join(goArgs, ",")+":"+c.descriptorOutDir,
		)

		for _, plugin := range c.customPlugins {
			cmd.Args = append(cmd.Args,
				"--"+plugin+"_out="+strings.Join(goArgs, ",")+":"+c.descriptorOutDir,
			)
		}
	}

	cmd.Args = append(
		cmd.Args,
		"-o", toFile,
		"--include_imports",
		"--include_source_info",
		protoFile)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "%v failed: %s", cmd.Args, out)
	}
	return nil
}
