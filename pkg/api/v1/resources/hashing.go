package resources

import (
	"github.com/mitchellh/hashstructure"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

// Hashers are resources which have a custom hashing function defined.
// this is typically done by placing an `extensions.go` file in the generated proto directory
// with a custom implementation for that resource
// Hashing is used by the snapshot emitter
type Hasher interface {
	Resource
	Hash() uint64
}

type HashOpts struct {
	IgnoreResourceStatus  bool
	IgnoreResourceVersion bool
}

var defaultHashOpts = HashOpts{
	IgnoreResourceStatus:  true,
	IgnoreResourceVersion: true,
}

func HashResource(resource Resource, hashOpts ...HashOpts) uint64 {
	if hasher, ok := resource.(Hasher); ok {
		return hasher.Hash()
	}
	opts := defaultHashOpts
	if len(hashOpts) > 0 {
		opts = hashOpts[0]
	}
	if opts.IgnoreResourceVersion {
		UpdateMetadata(resource, func(meta *core.Metadata) {
			meta.ResourceVersion = ""
		})
	}
	if opts.IgnoreResourceStatus {
		if inputRes, ok := resource.(InputResource); ok {
			inputRes.SetStatus(core.Status{})
		}
	}
	h, err := hashstructure.Hash(resource, nil)
	if err != nil {
		panic("resource failed to hash: " + err.Error())
	}
	return h
}
