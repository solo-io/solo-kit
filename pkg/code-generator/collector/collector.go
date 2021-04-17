package collector

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/solo-io/solo-kit/pkg/code-generator/metrics"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/stringutils"
)

type Collector interface {
	CollectImportsForFile(root, protoFile string) ([]string, error)
}

func NewCollector(customImports, commonImports []string) *collector {
	return &collector{
		customImports:    customImports,
		commonImports:    commonImports,
		importsExtractor: NewSynchronizedImportsExtractor(),
	}
}

type collector struct {
	customImports []string
	commonImports []string

	// The collector traverses a tree of files, opening and parsing each one.
	// These are costly operations that are often duplicated (ie fileA and fileB both import fileC)
	// The importsExtractor allows us to separate *how* to fetch imports from a file
	// from *when* to fetch imports from a file. This allows us to define a caching layer
	// in the importsExtractor and the collector doesn't have to be aware of it.
	importsExtractor ImportsExtractor
}

func (c *collector) CollectImportsForFile(root, protoFile string) ([]string, error) {
	return c.synchronizedImportsForProtoFile(root, protoFile, c.customImports)
}

var protoImportStatementRegex = regexp.MustCompile(`.*import "(.*)";.*`)

func (c *collector) detectImportsForFile(file string) ([]string, error) {
	metrics.IncrementFrequency("collector-opened-files")

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
			return nil, eris.Errorf("parsing import line error: from %v found %v", line, importStatement)
		}
		protoImports = append(protoImports, importStatement[1])
	}
	return protoImports, nil
}

func (c *collector) synchronizedImportsForProtoFile(absoluteRoot, protoFile string, customImports []string) ([]string, error) {
	// Define how we will extract the imports for the proto file
	importsFetcher := func(protoFileName string) ([]string, error) {
		return c.importsForProtoFile(absoluteRoot, protoFileName, customImports)
	}

	return c.importsExtractor.FetchImportsForFile(protoFile, importsFetcher)
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
		dependencyImports, err := c.synchronizedImportsForProtoFile(absoluteRoot, dependency, customImports)
		if err != nil {
			return nil, eris.Wrapf(err, "getting imports for dependency")
		}
		importsForProto = append(importsForProto, strings.TrimSuffix(importPath, "/"))
		importsForProto = append(importsForProto, dependencyImports...)
	}

	return stringutils.Unique(importsForProto), nil
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
		return "", eris.Errorf("found no possible import paths in root directory %v for import %v",
			absoluteRoot, importedProtoFile)
	}
	if len(possibleImportPaths) != 1 {
		log.Debugf("found more than one possible import path in root directory for "+
			"import %v: %v",
			importedProtoFile, possibleImportPaths)
	}
	return possibleImportPaths[0], nil

}
