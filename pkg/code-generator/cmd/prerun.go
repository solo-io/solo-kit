package cmd

import (
	"github.com/solo-io/solo-kit/pkg/protodep"
)

// Expose proto dep as a prerun func for solo-kit
func PreRunProtoVendor(cwd string, opts *protodep.Config) func() error {
	return func() error {
		mgr, err := protodep.NewManager(cwd)
		if err != nil {
			return err
		}
		return mgr.Gather(opts)
	}
}
