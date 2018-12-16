package validate

import (
	"fmt"
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/solo-io/solo-kit/cmd/cli/util"
	"github.com/spf13/cobra"
	"path/filepath"
)

var validFileTypes = []string{"proto", "protos"}

func Cmd(opts *options.Options) *cobra.Command{
	cmd := &cobra.Command{
		Use: "validate",
		Aliases: []string{"v"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Validate(opts, args)
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return util.EnsureConfigFile(opts)
		},
	}
	return cmd
}

func Validate(opts *options.Options, args []string) error {
	if len(args) > 0 {
		err := parseArgs(args, opts.Vaidate.All)
		if err != nil {
			return err
		}
	} else {
		//err :=
	}

	return nil
}

func parseArgs(args []string, all bool) error {
	if len(args) == 0 {
		return fmt.Errorf("this command requires at least one arg")
	}
	for _, v := range args {
		ext := filepath.Ext(v)
		if ext == "" {
			// Check if contained in valid plugins
		} else {
			// Check if file exists and has valid file extension
		}
	}
	return nil
}

