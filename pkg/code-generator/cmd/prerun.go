package cmd

import (
	"github.com/solo-io/solo-kit/pkg/protodep"
)

// Expose proto dep as a prerun func for solo-kit
func PreRunProtoVendor(cwd string, opts protodep.Options) func() error {
	return func() error {
		mgr, err := protodep.NewManager(cwd)
		if err != nil {
			return err
		}
		modules, err := mgr.Gather(opts)
		if err != nil {
			return err
		}
		if err := mgr.Copy(modules); err != nil {
			return err
		}
		return nil
	}
}
