package protodep

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/stringutils"
)

const (
	defaultMatchPattern = "**/*.proto"
)

type Manager interface {
	Gather(opts Options) ([]*Module, error)
	Copy([]*Module) error
}

type Options struct {
	WorkingDirectory string
	MatchPattern     string
	IncludePackages  []string
}

type Module struct {
	ImportPath    string
	SourcePath    string
	Version       string
	SourceVersion string
	Dir           string   // full path, $GOPATH/pkg/mod/
	VendorList    []string // files to vendor
}

func NewManager() *manager {
	return &manager{}
}

type manager struct{}

func (m *manager) Gather(opts Options) ([]*Module, error) {
	if !filepath.IsAbs(opts.WorkingDirectory) {
		absoluteDir, err := filepath.Abs(opts.WorkingDirectory)
		if err != nil {
			return nil, err
		}
		opts.WorkingDirectory = absoluteDir
	}
	if opts.MatchPattern == "" {
		opts.MatchPattern = defaultMatchPattern
	}
	// Ensure go.mod file exists and we're running from the project root,
	// and that ./vendor/modules.txt file exists.
	if _, err := os.Stat(filepath.Join(opts.WorkingDirectory, "go.mod")); os.IsNotExist(err) {
		fmt.Println("Whoops, cannot find `go.mod` file")
		return nil, err
	}
	modtxtPath := filepath.Join(opts.WorkingDirectory, "vendor", "modules.txt")
	if _, err := os.Stat(modtxtPath); os.IsNotExist(err) {
		fmt.Println("Whoops, cannot find vendor/modules.txt, first run `go mod vendor` and try again")
		return nil, err
	}

	// Parse/process modules.txt file of pkgs
	f, _ := os.Open(modtxtPath)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	modules := []*Module{}

	for scanner.Scan() {
		line := scanner.Text()
		// # character
		if line[0] != 35 {
			continue
		}
		s := strings.Split(line, " ")

		module := &Module{
			ImportPath: s[1],
			Version:    s[2],
		}
		if s[2] == "=>" {
			// issue https://github.com/golang/go/issues/33848 added these,
			// see comments. I think we can get away with ignoring them.
			continue
		}
		// Handle "replace" in module file if any
		if len(s) > 3 && s[3] == "=>" {
			module.SourcePath = s[4]
			module.SourceVersion = s[5]
			module.Dir = pkgModPath(module.SourcePath, module.SourceVersion)
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
		vendorList, err := buildModVendorList([]string{"**/*.proto"}, module)
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
			localFile := fmt.Sprintf("./vendor/%s", localPath)

			// if *verboseFlag {
			// 	fmt.Printf("vendoring %s\n", localPath)
			// }

			os.MkdirAll(filepath.Dir(localFile), os.ModePerm)
			if _, err := copyFile(vendorFile, localFile); err != nil {
				fmt.Printf("Error! %s - unable to copy file %s\n", err.Error(), vendorFile)
				return err
			}
		}
	}
	return nil
}

func buildModVendorList(copyPat []string, mod *Module) ([]string, error) {
	var vendorList []string

	for _, pat := range copyPat {
		matches, err := filepath.Glob(filepath.Join(mod.Dir, pat))
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
