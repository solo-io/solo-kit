package generate

import (
	"fmt"
	"github.com/solo-io/solo-kit/cmd/cli/flags"
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/solo-io/solo-kit/cmd/cli/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use: "generate",
		Aliases: []string{"g"},
		Short: "generate solo-kit protos",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return util.EnsureConfigFile(opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return generate(cmd, args, opts)
		},
	}
	pflags := cmd.PersistentFlags()
	flags.ConfigFlags(pflags, opts)
	return cmd
}

func generate(cmd *cobra.Command, args []string, opts *options.Options) error {

	if opts.ConfigFile != "" {
		err := util.ReadConfigFile(opts)
		if err != nil {
			return err
		}
	}

	if err := VerifyDirectories(&opts.Config); err != nil {
		return err
	}

	errorHandler := &util.ErrorWriter{}
	protoc := buildProtoCommand(&opts.Config)
	protoc.Stderr = errorHandler
	err := protoc.Run()
	if err != nil {
		return errorHandler
	}


	return nil
}



func VerifyDirectories(cfg *options.Config) error {
	cfg.Root = os.ExpandEnv(cfg.Root)
	if _, err := ioutil.ReadDir(cfg.Root); err != nil {
		return err
	}
	err := os.Setenv("ROOT", cfg.Root)
	if err != nil {
		return err
	}
	err = os.Chdir(cfg.Root)
	if err != nil {
		return err
	}
	cfg.Input = translatePath(cfg.Root, cfg.Input)
	if _, err := ioutil.ReadDir(cfg.Input); err != nil {
		return err
	}
	cfg.Output = translatePath(cfg.Root, cfg.Output)
	if _, err := ioutil.ReadDir(cfg.Output); err != nil {
		return err
	}


	//Handle Globs
	// TODO(EItanya): Handle possible globs everywhere
	var expandedSoloKitImports []string
	for _, dir := range cfg.Imports {
		expandedDir := os.ExpandEnv(dir)
		globbedFiles, err := filepath.Glob(expandedDir)
		if err == nil {
			expandedSoloKitImports = append(expandedSoloKitImports, globbedFiles...)
		} else {
			expandedSoloKitImports = append(expandedSoloKitImports, expandedDir)
		}
	}
	cfg.Imports = expandedSoloKitImports

	return nil
}

func translatePath(root, pathStr string) string {
	if strings.Contains(pathStr, "$") {
		//Check with ENV vars
		dir := os.ExpandEnv(pathStr)
		_, err := ioutil.ReadDir(dir)
		if err == nil {
			return path.Clean(dir)
		}
	}

	// Check relative to Root
	dir, err := filepath.Abs(pathStr)
	if err == nil {
		return dir
	}
	return pathStr
}

func buildGoGoImports(imports []string) []string {
	imps := make([]string, len(imports))
	for i, v := range imports {
		imps[i] = fmt.Sprintf("-I=%s", v)
	}
	return imps
}

func buildProtoCommand(cfg *options.Config) *exec.Cmd {
	args := buildGoGoImports(cfg.Imports)
	args = append(args, util.GOGO_FLAG)
	args = append(args, util.SOLO_KIT_FLAG(cfg))
	cmd := exec.Command("protoc", args...)
	fmt.Println(cmd.Args)
	return cmd
}


