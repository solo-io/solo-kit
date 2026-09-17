// Copyright 2018 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/solo-io/go-utils/contextutils"
)

var (
	// Compile-time assertion
	_ Snapshot = new(GenericSnapshot)
)

type TypedResources map[string]Resources

type GenericSnapshot struct {
	typedResources TypedResources
}

// Combine snapshots with distinct types to one.
func (s *GenericSnapshot) Combine(a *GenericSnapshot) (*GenericSnapshot, error) {
	if s.typedResources == nil {
		return a, nil
	} else if a.typedResources == nil {
		return s, nil
	}
	combined := TypedResources{}
	for k, v := range s.typedResources {
		combined[k] = v
	}
	for k, v := range a.typedResources {
		if _, ok := combined[k]; ok {
			return nil, errors.New("overlapping types found")
		}
		combined[k] = v
	}
	return NewGenericSnapshot(combined), nil
}

// Combine snapshots with distinct types to one.
func (s *GenericSnapshot) Merge(newSnap *GenericSnapshot) (*GenericSnapshot, error) {
	if s.typedResources == nil {
		return newSnap, nil
	}
	combined := TypedResources{}
	for k, v := range s.typedResources {
		combined[k] = v
	}
	for k, v := range newSnap.typedResources {
		combined[k] = v
	}
	return NewGenericSnapshot(combined), nil
}

// NewSnapshot creates a snapshot from response types and a version.
func NewGenericSnapshot(resources TypedResources) *GenericSnapshot {
	return &GenericSnapshot{
		typedResources: resources,
	}
}
func NewEasyGenericSnapshot(version string, resources ...[]Resource) *GenericSnapshot {
	t := TypedResources{}

	for _, resources := range resources {
		for _, resource := range resources {
			r := t[resource.Self().Type]
			if r.Items == nil {
				r.Items = make(map[string]Resource)
				r.Version = version
			}
			r.Items[resource.Self().Name] = resource
			t[resource.Self().Type] = r
		}
	}

	return &GenericSnapshot{
		typedResources: t,
	}
}

func (s *GenericSnapshot) Consistent() error {
	if s == nil {
		return nil
	}

	var required []XdsResourceReference

	for _, resources := range s.typedResources {
		for _, resource := range resources.Items {
			required = append(required, resource.References()...)
		}
	}

	for _, ref := range required {
		if resources, ok := s.typedResources[ref.Type]; ok {
			if _, ok := resources.Items[ref.Name]; !ok {
				return fmt.Errorf("required resource name not in snapshot: %s %s", ref.Type, ref.Name)
			}
		} else {
			return fmt.Errorf("required resource type not in snapshot: %s %s", ref.Type, ref.Name)
		}
	}

	return nil
}

func (s *GenericSnapshot) MakeConsistent() {
	// this is fine since generic snapshots are only used by extauth/ratelimit extensions syncers; and those don't
	// have dependent resources. this will not be called anywhere
	contextutils.LoggerFrom(context.TODO()).DPanicf("it is an error to call make consistent on a generic snapshot")
	if s == nil {
		return
	}
}

// GetResources selects snapshot resources by type.
func (s *GenericSnapshot) GetResources(typ string) Resources {
	if s == nil {
		return Resources{}
	}

	return s.typedResources[typ]
}

// Clone deep-copies the snapshot, carrying each resource's identity -- name,
// type and references -- across to the copy.
//
// The identity has to be carried rather than recomputed. A GenericSnapshot
// holds arbitrary Resource implementations, and those name themselves from
// whatever field of the wrapped proto suits them, so nothing here can work out
// what a copied proto should be called. Asking the original is the only option
// that keeps a copy indistinguishable from what it was copied from.
func (s *GenericSnapshot) Clone() Snapshot {
	if s == nil {
		return &GenericSnapshot{}
	}
	typedResourcesCopy := make(TypedResources, len(s.typedResources))
	for typeName, resources := range s.typedResources {
		resourcesCopy := Resources{
			Version: resources.Version,
			Items:   make(map[string]Resource, len(resources.Items)),
		}
		for k, v := range resources.Items {
			resourcesCopy.Items[k] = &clonedResource{
				ProtoMessage: proto.Clone(v.ResourceProto()),
				self:         v.Self(),
				references:   v.References(),
			}
		}
		typedResourcesCopy[typeName] = resourcesCopy
	}
	return &GenericSnapshot{typedResources: typedResourcesCopy}
}

// MarshalJSON renders the snapshot's resources, grouped by type URL. Without it
// the resources are unreachable to a marshaller, because they are held in an
// unexported field, and a snapshot renders as an empty object.
func (s *GenericSnapshot) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	return json.Marshal(s.typedResources)
}

// clonedResource is a Resource produced by copying another one. It holds a copy
// of the original's proto alongside the identity the original reported.
type clonedResource struct {
	// ProtoMessage is exported so that a cloned resource marshals the way
	// resource.EnvoyResource does, which names the same field.
	ProtoMessage ResourceProto

	self       XdsResourceReference
	references []XdsResourceReference
}

var _ Resource = new(clonedResource)

func (r *clonedResource) Self() XdsResourceReference         { return r.self }
func (r *clonedResource) ResourceProto() ResourceProto       { return r.ProtoMessage }
func (r *clonedResource) References() []XdsResourceReference { return r.references }
