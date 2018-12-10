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
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use: "generate",
		Aliases: []string{"g"},
		Short: "generate solo-kit protos",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generate(cmd, args, opts)
		},
	}
	pflags := cmd.PersistentFlags()
	flags.ConfigFlags(pflags, opts)
	return cmd
}

func generate(cmd *cobra.Command, args []string, opts *options.Options) error {
	err := util.EnsureConfigFile(opts)
	if err != nil {
		return err
	}
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
	protoc := buildGogoProtoCommand(&opts.Config)
	protoc.Stderr = errorHandler
	err = protoc.Run()
	if err != nil {
		return errorHandler
	}
	errorHandler.Flush()
	protoc = buildSKProtoCommand(&opts.Config)
	protoc.Stderr = errorHandler
	err = protoc.Run()
	if err != nil {
		return err
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

	// Expand env variables
	for i, dir := range cfg.GogoImports {
		cfg.GogoImports[i] = translatePath(cfg.Root, dir)
	}

	for i, dir := range cfg.SoloKitImports {
		cfg.SoloKitImports[i] = os.ExpandEnv(dir)
	}
	return nil
}

func translatePath(root, pathStr string) string {
	//Check with ENV vars
	dir := os.ExpandEnv(pathStr)
	_, err := ioutil.ReadDir(dir)
	if err == nil {
		return path.Clean(dir)
	}
	// Check relative to Root
	dir, err = filepath.Abs(pathStr)
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

func buildGogoProtoCommand(cfg *options.Config) *exec.Cmd {
	args := buildGoGoImports(cfg.GogoImports)
	inputProtos := fmt.Sprintf("%s/*.proto", cfg.Input)
	args = append(args, inputProtos)
	args = append(args, util.GOGO_FLAG)
	cmd := exec.Command("protoc", args...)
	fmt.Println(cmd.Args)
	return cmd
}



func buildSKProtoCommand(cfg *options.Config) *exec.Cmd {
	args := buildGoGoImports(cfg.GogoImports)
	args = append(args, util.SOLO_KIT_FLAG(cfg))
	inputProtos := make([]string, len(cfg.SoloKitImports) + 2)
	for _,v := range cfg.SoloKitImports {
		inputProtos = append(inputProtos, v)
	}
	args = append(args,  inputProtos...)
	cmd := exec.Command("protoc", args...)
	fmt.Println(cmd.Args)
	return cmd
}