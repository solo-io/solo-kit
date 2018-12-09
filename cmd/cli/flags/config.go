package flags

import (
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/spf13/pflag"
)

const (
	input = "input"
	input_default = "/api/v1/"
	output = "output"
	output_default = "pkg/api/v1"
)

func ConfigFlags(flags *pflag.FlagSet, cfg *options.Config) {
	flags.StringVarP(&cfg.Input, input, "i", input_default, "input protos")

	flags.StringVarP(&cfg.Output, output, "o", output_default, "output directory")
	flags.StringVar(&cfg.Dir, "config", "", "config file (default is $PROJECT_ROOT/solo-kit.yaml)")
}