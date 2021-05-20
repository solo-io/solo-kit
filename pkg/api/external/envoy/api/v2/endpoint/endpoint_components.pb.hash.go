// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/api/v2/endpoint/endpoint_components.proto

package endpoint

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"

	"github.com/mitchellh/hashstructure"
	safe_hasher "github.com/solo-io/protoc-gen-ext/pkg/hasher"

	core "github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/core"
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

	_ = core.HealthStatus(0)
)

// Hash function
func (m *Endpoint) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("solo.io.envoy.api.v2.endpoint.github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/endpoint.Endpoint")); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetAddress()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Address")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetAddress(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Address")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetHealthCheckConfig()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("HealthCheckConfig")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetHealthCheckConfig(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("HealthCheckConfig")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if _, err = hasher.Write([]byte(m.GetHostname())); err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *LbEndpoint) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("solo.io.envoy.api.v2.endpoint.github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/endpoint.LbEndpoint")); err != nil {
		return 0, err
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetHealthStatus())
	if err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetMetadata()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Metadata")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetMetadata(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Metadata")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetLoadBalancingWeight()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("LoadBalancingWeight")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetLoadBalancingWeight(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("LoadBalancingWeight")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	switch m.HostIdentifier.(type) {

	case *LbEndpoint_Endpoint:

		if h, ok := interface{}(m.GetEndpoint()).(safe_hasher.SafeHasher); ok {
			if _, err = hasher.Write([]byte("Endpoint")); err != nil {
				return 0, err
			}
			if _, err = h.Hash(hasher); err != nil {
				return 0, err
			}
		} else {
			if fieldValue, err := hashstructure.Hash(m.GetEndpoint(), nil); err != nil {
				return 0, err
			} else {
				if _, err = hasher.Write([]byte("Endpoint")); err != nil {
					return 0, err
				}
				if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
					return 0, err
				}
			}
		}

	case *LbEndpoint_EndpointName:

		if _, err = hasher.Write([]byte(m.GetEndpointName())); err != nil {
			return 0, err
		}

	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *LocalityLbEndpoints) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("solo.io.envoy.api.v2.endpoint.github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/endpoint.LocalityLbEndpoints")); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetLocality()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Locality")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetLocality(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Locality")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	for _, v := range m.GetLbEndpoints() {

		if h, ok := interface{}(v).(safe_hasher.SafeHasher); ok {
			if _, err = hasher.Write([]byte("")); err != nil {
				return 0, err
			}
			if _, err = h.Hash(hasher); err != nil {
				return 0, err
			}
		} else {
			if fieldValue, err := hashstructure.Hash(v, nil); err != nil {
				return 0, err
			} else {
				if _, err = hasher.Write([]byte("")); err != nil {
					return 0, err
				}
				if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
					return 0, err
				}
			}
		}

	}

	if h, ok := interface{}(m.GetLoadBalancingWeight()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("LoadBalancingWeight")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetLoadBalancingWeight(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("LoadBalancingWeight")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetPriority())
	if err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetProximity()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Proximity")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetProximity(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Proximity")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *Endpoint_HealthCheckConfig) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("solo.io.envoy.api.v2.endpoint.github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/endpoint.Endpoint_HealthCheckConfig")); err != nil {
		return 0, err
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetPortValue())
	if err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetHostname())); err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}
