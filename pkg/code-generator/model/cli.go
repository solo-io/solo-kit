package model

import (
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

type CliConfig struct {
	Name string
	Path string
	// String representation of go package name for current project
	// e.g. "github.com/solo-io/solo-kit/..."
	ImportDir string
}

type CliFile struct {
	Filename    string
	Imports     []string
	PackageName string
	Resources   []*Resource
}

func (file *CliFile) AddImport(imports ...string) error {
	for _, v := range imports {
		fullImportPath, err := filepath.Abs(v)
		if err != nil {
			return err
		}
		importPath := strings.Split(fullImportPath, "go/src/")[1]
		file.Imports = append(file.Imports, fmt.Sprintf("%s", importPath))
	}
	return nil
}

func (file *CliFile) StrImports() string {
	return strings.Join(file.Imports, "\n")
}

type CliResourceFile struct {
	CliFile
	IsRoot   bool
	Cmd      *cobra.Command
	Resource *Resource
}