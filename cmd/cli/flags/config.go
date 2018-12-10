package flags

import (
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/spf13/pflag"
)

const (
	input = "input"
	input_default = "api/v1/"
	output = "output"
	output_default = "pkg/api/v1"
	docs = "docs"
	docs_default = "doc/docs"
	config = "config"
	gogo_imports = "gogo-imports"
	sk_imports = "solo-kit-imports"
)

func ConfigFlags(flags *pflag.FlagSet, opts *options.Options) {
	flags.StringVarP(&opts.Config.Input, input, "i", input_default, "input protos")
	flags.StringVarP(&opts.Config.Output, output, "o", output_default, "output directory")
	flags.StringVarP(&opts.Config.Docs, docs, "d", docs_default, "output directory for docs, if different from normal.")

	flags.StringVar(&opts.ConfigFile, config, "", "config file (default is $PROJECT_ROOT/solo-kit.yaml)")

	flags.StringSliceVar(&opts.Config.GogoImports, gogo_imports, []string{}, "List of external directories with protos for gogo")
	flags.StringSliceVar(&opts.Config.SoloKitImports, sk_imports, []string{}, "List of external proto files for solo-kit")

}