package collector

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/metrics"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
	"github.com/solo-io/solo-kit/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func init() {
	// initalize metrics during init phase.
	// can be safely called multiple times
	metrics.NewAggregator()
}

type ProtoCompiler interface {
	CompileDescriptorsFromRoot(root string, skipDirs []string) ([]*model.DescriptorWithPath, error)
}

func NewProtoCompiler(collector Collector, executor ProtocExecutor) *protoCompiler {
	return &protoCompiler{
		importsCollector: collector,
		protocExecutor:   executor,
	}
}

type protoCompiler struct {
	importsCollector Collector
	protocExecutor   ProtocExecutor
}

func (c *protoCompiler) CompileDescriptorsFromRoot(root string, skipDirs []string) ([]*model.DescriptorWithPath, error) {
	defer metrics.MeasureElapsed("proto-compiler", time.Now())
	log.Printf("Compiling proto descriptors from root:  %s", root)

	var descriptors []*model.DescriptorWithPath
	var mutex sync.Mutex
	addDescriptor := func(f model.DescriptorWithPath) {
		mutex.Lock()
		defer mutex.Unlock()
		descriptors = append(descriptors, &f)
	}
	var (
		g            errgroup.Group
		sem              chan struct{}
		limitConcurrency bool
	)
	if s := os.Getenv("MAX_CONCURRENT_PROTOCS"); s != "" {
		maxProtocs, err := strconv.Atoi(s)
		if err != nil {
			return nil, eris.Wrapf(err, "invalid value for MAX_CONCURRENT_PROTOCS: %s", s)
		}
		sem = make(chan struct{}, maxProtocs)
		limitConcurrency = true
	}

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
				if limitConcurrency {
					sem <- struct{}{}
				}
				defer func() {
					if limitConcurrency {
						<-sem
					}
				}()
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
	log.Debugf("processing proto file input %v", protoFile)

	imports, err := c.importsCollector.CollectImportsForFile(root, protoFile)
	if err != nil {
		return errors.Wrapf(err, "reading imports for proto file")
	}

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

	if err := c.protocExecutor.Execute(protoFile, tmpFile.Name(), imports); err != nil {
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
