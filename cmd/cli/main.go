package main

import (
	"github.com/solo-io/solo-kit/cmd/cli/generate"
	"github.com/solo-io/solo-kit/cmd/cli/initialize"
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/spf13/cobra"
)

func main() {
	opts := &options.Options{}
	root := RootCmd(opts)
	root.Execute()
}

// rootCmd represents the base command when called without any subcommands
func RootCmd(opts *options.Options) *cobra.Command {
	//opts := &options.Options{}
	cmd := &cobra.Command{
		Use:     "solo-kit",
		Short:   "cli for solo-kit",
		Aliases: []string{"sk"},
	}
	cmd.AddCommand(
		generate.Cmd(opts),
		initialize.Cmd(opts),
	)
	return cmd
}
