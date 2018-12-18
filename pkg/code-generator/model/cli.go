package model

import (
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

type CliConfig struct {
	Name    string
	Path    string
	Version string
}

type CliProject struct {
	CliConfig
	Resources      []*Resource
	ResourceGroups []*ResourceGroup
}

type CliFile struct {
	Filename    string
	PackageName string
	Resources   []*Resource

	imports []string
}

func (file *CliFile) AddImport(imports ...string) error {
	for _, v := range imports {
		fullImportPath, err := filepath.Abs(v)
		if err != nil {
			return err
		}
		importPath := strings.Split(fullImportPath, "go/src/")[1]
		file.imports = append(file.imports, fmt.Sprintf("%s", importPath))
	}
	return nil
}

func (file *CliFile) Imports() []string {
	return file.imports
}

func (file *CliFile) StrImports() string {
	return strings.Join(file.imports, "\n")
}

type CliResourceFile struct {
	CliFile
	IsRoot   bool
	Cmd      *cobra.Command
	CmdName  string
	Resource *Resource
}
