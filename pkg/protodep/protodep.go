package protodep

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	zglob "github.com/mattn/go-zglob"
	"github.com/solo-io/go-utils/errors"
)

const (
	ProtoMatchPattern   = "**/*.proto"
	SoloKitMatchPattern = "**/solo-kit.json"
)

var (
	// offer sane defaults for proto vendoring
	DefaultMatchPatterns = []string{ProtoMatchPattern, SoloKitMatchPattern}

	// matches ext.proto for solo hash gen
	ExtProtoMatcher = MatchOptions{
		Package:  "github.com/solo-io/protoc-gen-ext",
		Patterns: []string{"extproto/*.proto"},
	}

	// matches validate.proto which is needed by envoy protos
	ValidateProtoMatcher = MatchOptions{
		Package:  "github.com/envoyproxy/protoc-gen-validate",
		Patterns: []string{"validate/*.proto"},
	}

	// matches all solo-kit protos, useful for any projects using solo-kit
	SoloKitProtoMatcher = MatchOptions{
		Package:  "github.com/solo-io/solo-kit",
		Patterns: []string{"api/**/*.proto"},
	}

	// matches gogo.proto, used for gogoproto code gen.
	GogoProtoMatcher = MatchOptions{
		Package:  "github.com/gogo/protobuf",
		Patterns: []string{"gogoproto/*.proto"},
	}

	// default match options which should be used when creating a solo-kit project
	DefaultMatchOptions = []MatchOptions{
		ExtProtoMatcher,
		ValidateProtoMatcher,
		SoloKitProtoMatcher,
		GogoProtoMatcher,
	}
)

type Manager interface {
	Gather(opts MatchOptions) ([]*Module, error)
	Copy([]*Module) error
}

// struct which represents how to vendor protos.
// Patters are a set of regexes which match protos for a given go package
// see examples above
type MatchOptions struct {
	Patterns []string
	Package  string
}

// struct which represents a go module package in the module package list
type Module struct {
	ImportPath    string
	SourcePath    string
	Version       string
	SourceVersion string
	Dir           string   // full path, $GOPATH/pkg/mod/
	VendorList    []string // files to vendor
}

// Expose proto dep as a prerun func for solo-kit
func PreRunProtoVendor(cwd string, matchOpts []MatchOptions) func() error {
	return func() error {
		mgr, err := NewManager(cwd)
		if err != nil {
			return err
		}
		modules, err := mgr.Gather(matchOpts)
		if err != nil {
			return err
		}
		if err := mgr.Copy(modules); err != nil {
			return err
		}
		return nil
	}
}

func NewManager(cwd string) (*manager, error) {
	if !filepath.IsAbs(cwd) {
		absoluteDir, err := filepath.Abs(cwd)
		if err != nil {
			return nil, err
		}
		cwd = absoluteDir
	}
	return &manager{
		WorkingDirectory: cwd,
	}, nil
}

type manager struct {
	WorkingDirectory string
	vendorMode       bool
}

// gather up all packages for a given go module
// currently this function uses the cmd `go list -m all` to figure out the list of dep
func (m *manager) Gather(matchOptions []MatchOptions) ([]*Module, error) {
	// Ensure go.mod file exists and we're running from the project root,
	if _, err := os.Stat(filepath.Join(m.WorkingDirectory, "go.mod")); os.IsNotExist(err) {
		fmt.Println("Whoops, cannot find `go.mod` file")
		return nil, err
	}

	modPackageReader := &bytes.Buffer{}
	packageListCmd := exec.Command("go", "list", "-m", "all")
	packageListCmd.Stdout = modPackageReader
	packageListCmd.Stderr = modPackageReader
	err := packageListCmd.Run()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list packages for current mod package: %s",
			modPackageReader.String())
	}

	// split list of pacakges from cmd by line
	scanner := bufio.NewScanner(modPackageReader)
	scanner.Split(bufio.ScanLines)

	// Clear first line as it is current package name
	scanner.Scan()

	modules := []*Module{}

	for scanner.Scan() {
		line := scanner.Text()
		s := strings.Split(line, " ")

		/*
			the packages come in 3 varities
			1. helm.sh/helm/v3 v3.0.0
			2. k8s.io/api v0.0.0-20191121015604-11707872ac1c => k8s.io/api v0.0.0-20191004120104-195af9ec3521
			3. k8s.io/api v0.0.0-20191121015604-11707872ac1c => /path/to/local

			All three variants share the same first 2 members
		*/
		module := &Module{
			ImportPath: s[0],
			Version:    s[1],
		}
		if s[1] == "=>" {
			// issue https://github.com/golang/go/issues/33848 added these,
			// see comments. I think we can get away with ignoring them.
			continue
		}
		// Handle "replace" in module file if any
		if len(s) > 2 && s[2] == "=>" {
			module.SourcePath = s[3]
			// non-local module with version
			if len(s) >= 5 {
				// see case 2 above
				module.SourceVersion = s[4]
				module.Dir = pkgModPath(module.SourcePath, module.SourceVersion)
			} else {
				// see case 3 above
				moduleAbsolutePath, err := filepath.Abs(module.SourcePath)
				if err != nil {
					return nil, err
				}
				module.Dir = moduleAbsolutePath
			}
		} else {
			module.Dir = pkgModPath(module.ImportPath, module.Version)
		}

		// make sure module exists
		if _, err := os.Stat(module.Dir); os.IsNotExist(err) {
			fmt.Printf("Error! %q module path does not exist, check $GOPATH/pkg/mod. "+
				"Try running go mod download\n", module.Dir)
			return nil, err
		}

		// If no match options have been supplied, match on all packages using default match patterns
		if matchOptions == nil {
			// Build list of files to module path source to project vendor folder
			vendorList, err := buildModVendorList(DefaultMatchPatterns, module)
			if err != nil {
				return nil, err
			}
			module.VendorList = vendorList
			if len(vendorList) > 0 {
				modules = append(modules, module)
			}
			continue
		}

		for _, matchOpt := range matchOptions {
			// only check module if is in imports list, or imports list in empty
			if len(matchOpt.Package) != 0 &&
				!strings.Contains(module.ImportPath, matchOpt.Package) {
				continue
			}
			// Build list of files to module path source to project vendor folder
			vendorList, err := buildModVendorList(matchOpt.Patterns, module)
			if err != nil {
				return nil, err
			}
			module.VendorList = vendorList
			if len(vendorList) > 0 {
				modules = append(modules, module)
			}
		}

	}

	return modules, nil
}

func (m *manager) Copy(modules []*Module) error {
	// Copy mod vendor list files to ./vendor/
	for _, mod := range modules {
		for _, vendorFile := range mod.VendorList {
			x := strings.Index(vendorFile, mod.Dir)
			if x < 0 {
				return errors.New("Error! vendor file doesn't belong to mod, strange.")
			}

			localPath := filepath.Join(mod.ImportPath, vendorFile[len(mod.Dir):])
			localFile := filepath.Join(m.WorkingDirectory, "vendor", localPath)

			log.Printf("copying %v -> %v", vendorFile, localFile)

			if err := os.MkdirAll(filepath.Dir(localFile), os.ModePerm); err != nil {
				return err
			}
			if _, err := copyFile(vendorFile, localFile); err != nil {
				return errors.Wrapf(err, fmt.Sprintf("Error! %s - unable to copy file %s\n",
					err.Error(), vendorFile))
			}
		}
	}
	return nil
}

func buildModVendorList(copyPat []string, mod *Module) ([]string, error) {
	var vendorList []string

	for _, pat := range copyPat {
		matches, err := zglob.Glob(filepath.Join(mod.Dir, pat))
		if err != nil {
			return nil, errors.Wrapf(err, "Error! glob match failure")
		}
		vendorList = append(vendorList, matches...)
	}

	return vendorList, nil
}

func pkgModPath(importPath, version string) string {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		// the default GOPATH for go v1.11
		goPath = filepath.Join(os.Getenv("HOME"), "go")
	}

	var normPath string

	for _, char := range importPath {
		if unicode.IsUpper(char) {
			normPath += "!" + string(unicode.ToLower(char))
		} else {
			normPath += string(char)
		}
	}

	return filepath.Join(goPath, "pkg", "mod", fmt.Sprintf("%s@%s", normPath, version))
}

func copyFile(src, dst string) (int64, error) {
	srcStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !srcStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}
