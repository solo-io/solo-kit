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
	"github.com/solo-io/go-utils/stringutils"
)

const (
	ProtoMatchPattern   = "**/*.proto"
	SoloKitMatchPattern = "**/solo-kit.json"
)

var (
	DefaultMatchPatterns = []string{ProtoMatchPattern, SoloKitMatchPattern}
)

type Manager interface {
	Gather(opts Options) ([]*Module, error)
	Copy([]*Module) error
}

type Options struct {
	MatchPatterns   []string
	IncludePackages []string
}

type Module struct {
	ImportPath    string
	SourcePath    string
	Version       string
	SourceVersion string
	Dir           string   // full path, $GOPATH/pkg/mod/
	VendorList    []string // files to vendor
}

func PreRunProtoVendor(cwd string, vendorPackages, matchPatterns []string) func() error {
	return func() error {
		mgr, err := NewManager(cwd)
		if err != nil {
			return err
		}
		opts := Options{
			// TODO(make this second matcher work!)
			MatchPatterns:   matchPatterns,
			IncludePackages: vendorPackages,
		}
		if opts.MatchPatterns == nil {
			opts.MatchPatterns = DefaultMatchPatterns
		}
		modules, err := mgr.Gather(opts)
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

func (m *manager) Gather(opts Options) ([]*Module, error) {
	if opts.MatchPatterns == nil {
		opts.MatchPatterns = []string{ProtoMatchPattern}
	}
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

	scanner := bufio.NewScanner(modPackageReader)
	scanner.Split(bufio.ScanLines)

	// Clear first line as it is current package name
	scanner.Scan()

	modules := []*Module{}

	for scanner.Scan() {
		line := scanner.Text()
		// # character
		// if line[0] != 35 {
		// 	continue
		// }
		s := strings.Split(line, " ")

		module := &Module{
			ImportPath: s[1-1],
			Version:    s[2-1],
		}
		if s[2-1] == "=>" {
			// issue https://github.com/golang/go/issues/33848 added these,
			// see comments. I think we can get away with ignoring them.
			continue
		}
		// Handle "replace" in module file if any
		if len(s) > 3-1 && s[3-1] == "=>" {
			module.SourcePath = s[4-1]
			// non-local module with version
			if len(s) >= 6-1 {
				module.SourceVersion = s[5-1]
				module.Dir = pkgModPath(module.SourcePath, module.SourceVersion)
			} else {
				moduleAbsolutePath, err := filepath.Abs(module.SourcePath)
				if err != nil {
					return nil, err
				}
				module.Dir = moduleAbsolutePath
			}
		} else {
			module.Dir = pkgModPath(module.ImportPath, module.Version)
		}

		if _, err := os.Stat(module.Dir); os.IsNotExist(err) {
			fmt.Printf("Error! %q module path does not exist, check $GOPATH/pkg/mod\n", module.Dir)
			return nil, err
		}

		// only check module if is in imports list, or imports list in empty
		if len(opts.IncludePackages) != 0 &&
			!stringutils.ContainsString(module.ImportPath, opts.IncludePackages) {
			continue
		}

		// Build list of files to module path source to project vendor folder
		vendorList, err := buildModVendorList(opts.MatchPatterns, module)
		if err != nil {
			return nil, err
		}
		module.VendorList = vendorList
		if len(vendorList) > 0 {
			modules = append(modules, module)
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
