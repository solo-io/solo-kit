package validate

import (
	"fmt"
	"github.com/solo-io/solo-kit/cmd/cli/options"
)

func ValidateProto(path string) error {
	return nil
}

func validateAllProtoso(opts *options.Options) error {
	if opts.Config.Input == "" {
		return fmt.Errorf("input directory cannot be empty")
	}
	if opts.Config.Root
	return nil
}