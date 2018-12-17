package generate

import (
	"github.com/solo-io/solo-kit/cmd/cli/flags"
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/solo-io/solo-kit/cmd/cli/util"
	"github.com/solo-io/solo-kit/cmd/solo-kit-gen"
	"github.com/spf13/cobra"
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
			return solo_kit_gen.Generate(cmd, args, opts)
		},
	}
	pflags := cmd.PersistentFlags()
	flags.GenerateFlags(pflags, opts)
	return cmd
}
