package protodep

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	zglob "github.com/mattn/go-zglob"
	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/modutils"
	"github.com/spf13/afero"
)

const (
	ProtoMatchPattern   = "**/*.proto"
	SoloKitMatchPattern = "**/solo-kit.json"
)

var (
	// offer sane defaults for proto vendoring
	DefaultMatchPatterns = []string{ProtoMatchPattern, SoloKitMatchPattern}

	// matches ext.proto for solo hash gen
	ExtProtoMatcher = &GoModImport{
		Package:  "github.com/solo-io/protoc-gen-ext",
		Patterns: []string{"extproto/*.proto"},
	}

	// matches validate.proto which is needed by envoy protos
	ValidateProtoMatcher = &GoModImport{
		Package:  "github.com/envoyproxy/protoc-gen-validate",
		Patterns: []string{"validate/*.proto"},
	}

	// matches all solo-kit protos, useful for any projects using solo-kit
	SoloKitProtoMatcher = &GoModImport{
		Package:  "github.com/solo-io/solo-kit",
		Patterns: []string{"api/**/*.proto", "api/" + SoloKitMatchPattern},
	}

	// matches gogo.proto, used for gogoproto code gen.
	GogoProtoMatcher = &GoModImport{
		Package:  "github.com/gogo/protobuf",
		Patterns: []string{"gogoproto/*.proto"},
	}

	// default match options which should be used when creating a solo-kit project
	DefaultMatchOptions = []Import{
		{
			ImportType: &Import_GoMod{
				GoMod: ExtProtoMatcher,
			},
		},
		{
			ImportType: &Import_GoMod{
				GoMod: ValidateProtoMatcher,
			},
		},
		{
			ImportType: &Import_GoMod{
				GoMod: SoloKitProtoMatcher,
			},
		},
		{
			ImportType: &Import_GoMod{
				GoMod: GogoProtoMatcher,
			},
		},
	}
)

type goModOptions struct {
	MatchOptions  []*GoModImport
	LocalMatchers []string
}

// struct which represents a go module package in the module package list
type Module struct {
	ImportPath     string
	SourcePath     string
	Version        string
	SourceVersion  string
	Dir            string   // full path, $GOPATH/pkg/mod/
	VendorList     []string // files to vendor
	currentPackage bool
}

func NewGoModFactory(cwd string) (*goModFactory, error) {
	if !filepath.IsAbs(cwd) {
		absoluteDir, err := filepath.Abs(cwd)
		if err != nil {
			return nil, err
		}
		cwd = absoluteDir
	}
	fs := afero.NewOsFs()
	return &goModFactory{
		WorkingDirectory: cwd,
		fs:               fs,
		cp:               NewCopier(fs),
	}, nil
}

type goModFactory struct {
	WorkingDirectory string
	packageName      bool
	fs               afero.Fs
	cp               FileCopier
}

func (m *goModFactory) Ensure(ctx context.Context, opts *Config) error {
	var packages []*GoModImport
	for _, cfg := range opts.Imports {
		if cfg.GetGoMod() != nil {
			packages = append(packages, cfg.GetGoMod())
		}
	}
	mods, err := m.gather(goModOptions{
		MatchOptions:  packages,
		LocalMatchers: opts.GetLocal().GetPatterns(),
	})
	if err != nil {
		return err
	}

	err = m.copy(mods)
	if err != nil {
		return err
	}
	return nil
}

// gather up all packages for a given go module
// currently this function uses the cmd `go list -m all` to figure out the list of dep
func (m *goModFactory) gather(opts goModOptions) ([]*Module, error) {
	matchOptions := opts.MatchOptions
	// Ensure go.mod file exists and we're running from the project root,
	modPackageFile, err := modutils.GetCurrentModPackageFile()
	if err != nil {
		return nil, err
	}

	packageName, err := modutils.GetCurrentModPackageName(modPackageFile)
	if err != nil {
		return nil, err
	}

	modPackageReader, err := modutils.GetCurrentPackageList()
	if err != nil {
		return nil, err
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
		if _, err := m.fs.Stat(module.Dir); os.IsNotExist(err) {
			fmt.Printf("Error! %q module path does not exist, check $GOPATH/pkg/mod. "+
				"Try running go mod download\n", module.Dir)
			return nil, err
		}

		// If no match options have been supplied, match on all packages using default match patterns
		if matchOptions == nil {
			// Build list of files to module path source to project vendor folder
			vendorList, err := buildMatchList(DefaultMatchPatterns, module.Dir)
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
			vendorList, err := buildMatchList(matchOpt.Patterns, module.Dir)
			if err != nil {
				return nil, err
			}
			module.VendorList = vendorList
			if len(vendorList) > 0 {
				modules = append(modules, module)
			}
		}

	}

	localModule := &Module{
		Dir:            m.WorkingDirectory,
		ImportPath:     packageName,
		currentPackage: true,
	}
	localModule.VendorList, err = buildMatchList(opts.LocalMatchers, localModule.Dir)
	if err != nil {
		return nil, err
	}
	modules = append(modules, localModule)

	return modules, nil
}

func (m *goModFactory) copy(modules []*Module) error {
	// Copy mod vendor list files to ./vendor/
	for _, mod := range modules {
		if mod.currentPackage == true {
			for _, vendorFile := range mod.VendorList {
				localPath := strings.TrimPrefix(vendorFile, m.WorkingDirectory+"/")
				localFile := filepath.Join(m.WorkingDirectory, DefaultDepDir, mod.ImportPath, localPath)
				if _, err := m.cp.Copy(vendorFile, localFile); err != nil {
					return errors.Wrapf(err, fmt.Sprintf("Error! %s - unable to copy file %s\n",
						err.Error(), vendorFile))
				}
			}
		}
		for _, vendorFile := range mod.VendorList {
			x := strings.Index(vendorFile, mod.Dir)
			if x < 0 {
				return errors.New("Error! vendor file doesn't belong to mod, strange.")
			}

			localPath := filepath.Join(mod.ImportPath, vendorFile[len(mod.Dir):])
			localFile := filepath.Join(m.WorkingDirectory, DefaultDepDir, localPath)
			if _, err := m.cp.Copy(vendorFile, localFile); err != nil {
				return errors.Wrapf(err, fmt.Sprintf("Error! %s - unable to copy file %s\n",
					err.Error(), vendorFile))
			}
		}
	}
	return nil
}

func buildMatchList(copyPat []string, dir string) ([]string, error) {
	var vendorList []string

	for _, pat := range copyPat {
		matches, err := zglob.Glob(filepath.Join(dir, pat))
		if err != nil {
			return nil, errors.Wrapf(err, "Error! glob match failure")
		}
		// Filter out all matches which contain a vendor folder, those are leftovers from a previous run.
		// Might be worth clearing the vendor folder before every run.
		for _, match := range matches {
			vendorFolders := strings.Count(match, DefaultDepDir)
			if vendorFolders > 0 {
				continue
			}
			vendorList = append(vendorList, match)
		}
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
