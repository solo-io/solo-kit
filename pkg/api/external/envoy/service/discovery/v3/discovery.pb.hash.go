// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/service/discovery/v3/discovery.proto

package v3

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"

	"github.com/mitchellh/hashstructure"
	safe_hasher "github.com/solo-io/protoc-gen-ext/pkg/hasher"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = new(hash.Hash64)
	_ = fnv.New64
	_ = hashstructure.Hash
	_ = new(safe_hasher.SafeHasher)
)

// Hash function
func (m *DiscoveryRequest) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("envoy.service.discovery.v3.github.com/solo-io/solo-kit/pkg/api/external/envoy/service/discovery/v3.DiscoveryRequest")); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetVersionInfo())); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetNode()).(safe_hasher.SafeHasher); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetNode(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	for _, v := range m.GetResourceNames() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	if _, err = hasher.Write([]byte(m.GetTypeUrl())); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetResponseNonce())); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetErrorDetail()).(safe_hasher.SafeHasher); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetErrorDetail(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *DiscoveryResponse) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("envoy.service.discovery.v3.github.com/solo-io/solo-kit/pkg/api/external/envoy/service/discovery/v3.DiscoveryResponse")); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetVersionInfo())); err != nil {
		return 0, err
	}

	for _, v := range m.GetResources() {

		if h, ok := interface{}(v).(safe_hasher.SafeHasher); ok {
			if _, err = h.Hash(hasher); err != nil {
				return 0, err
			}
		} else {
			if val, err := hashstructure.Hash(v, nil); err != nil {
				return 0, err
			} else {
				if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
					return 0, err
				}
			}
		}

	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetCanary())
	if err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetTypeUrl())); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetNonce())); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetControlPlane()).(safe_hasher.SafeHasher); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetControlPlane(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *DeltaDiscoveryRequest) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("envoy.service.discovery.v3.github.com/solo-io/solo-kit/pkg/api/external/envoy/service/discovery/v3.DeltaDiscoveryRequest")); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetNode()).(safe_hasher.SafeHasher); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetNode(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	if _, err = hasher.Write([]byte(m.GetTypeUrl())); err != nil {
		return 0, err
	}

	for _, v := range m.GetResourceNamesSubscribe() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	for _, v := range m.GetResourceNamesUnsubscribe() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	{
		var result uint64
		innerHash := fnv.New64()
		for k, v := range m.GetInitialResourceVersions() {
			innerHash.Reset()

			if _, err = innerHash.Write([]byte(v)); err != nil {
				return 0, err
			}

			if _, err = innerHash.Write([]byte(k)); err != nil {
				return 0, err
			}

			result = result ^ innerHash.Sum64()
		}
		err = binary.Write(hasher, binary.LittleEndian, result)
		if err != nil {
			return 0, err
		}

	}

	if _, err = hasher.Write([]byte(m.GetResponseNonce())); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetErrorDetail()).(safe_hasher.SafeHasher); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetErrorDetail(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *DeltaDiscoveryResponse) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("envoy.service.discovery.v3.github.com/solo-io/solo-kit/pkg/api/external/envoy/service/discovery/v3.DeltaDiscoveryResponse")); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetSystemVersionInfo())); err != nil {
		return 0, err
	}

	for _, v := range m.GetResources() {

		if h, ok := interface{}(v).(safe_hasher.SafeHasher); ok {
			if _, err = h.Hash(hasher); err != nil {
				return 0, err
			}
		} else {
			if val, err := hashstructure.Hash(v, nil); err != nil {
				return 0, err
			} else {
				if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
					return 0, err
				}
			}
		}

	}

	if _, err = hasher.Write([]byte(m.GetTypeUrl())); err != nil {
		return 0, err
	}

	for _, v := range m.GetRemovedResources() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	if _, err = hasher.Write([]byte(m.GetNonce())); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetControlPlane()).(safe_hasher.SafeHasher); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetControlPlane(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *Resource) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("envoy.service.discovery.v3.github.com/solo-io/solo-kit/pkg/api/external/envoy/service/discovery/v3.Resource")); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetName())); err != nil {
		return 0, err
	}

	for _, v := range m.GetAliases() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	if _, err = hasher.Write([]byte(m.GetVersion())); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetResource()).(safe_hasher.SafeHasher); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetResource(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetTtl()).(safe_hasher.SafeHasher); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetTtl(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetCacheControl()).(safe_hasher.SafeHasher); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetCacheControl(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *Resource_CacheControl) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("envoy.service.discovery.v3.github.com/solo-io/solo-kit/pkg/api/external/envoy/service/discovery/v3.Resource_CacheControl")); err != nil {
		return 0, err
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetDoNotCache())
	if err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}
