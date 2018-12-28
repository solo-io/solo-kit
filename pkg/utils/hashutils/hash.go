package hashutils

import (
	"github.com/mitchellh/hashstructure"
)

// hash one or more values
// order matters
func HashAll(values ...interface{}) uint64 {
	h, err := hashstructure.Hash(values, nil)
	if err != nil {
		panic(err)
	}
	return h
}
