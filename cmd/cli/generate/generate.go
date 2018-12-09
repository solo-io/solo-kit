package generate

import (
	"fmt"
	"github.com/solo-io/solo-kit/cmd/cli/flags"
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/solo-io/solo-kit/cmd/cli/util"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	util.EnsureConfig(&opts.Config)
	cmd := &cobra.Command{
		Use: "generate",
		Aliases: []string{"g"},
		Short: "generate solo-kit protos",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(opts)
			return nil
		},
	}
	pflags := cmd.PersistentFlags()
	flags.ConfigFlags(pflags, &opts.Config)
	return cmd
}