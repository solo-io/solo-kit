// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: metadata.proto

package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/solo-io/protoc-gen-ext/pkg/clone"
	"google.golang.org/protobuf/proto"

	github_com_golang_protobuf_ptypes_wrappers "github.com/golang/protobuf/ptypes/wrappers"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = bytes.Compare
	_ = strings.Compare
	_ = clone.Cloner(nil)
	_ = proto.Message(nil)
)

// Clone function
func (m *Metadata) Clone() proto.Message {
	var target *Metadata
	if m == nil {
		return target
	}
	target = &Metadata{}

	target.Name = m.GetName()

	target.Namespace = m.GetNamespace()

	target.Cluster = m.GetCluster()

	target.ResourceVersion = m.GetResourceVersion()

	if m.GetLabels() != nil {
		target.Labels = make(map[string]string, len(m.GetLabels()))
		for k, v := range m.GetLabels() {

			target.Labels[k] = v

		}
	}

	if m.GetAnnotations() != nil {
		target.Annotations = make(map[string]string, len(m.GetAnnotations()))
		for k, v := range m.GetAnnotations() {

			target.Annotations[k] = v

		}
	}

	target.Generation = m.GetGeneration()

	if m.GetOwnerReferences() != nil {
		target.OwnerReferences = make([]*Metadata_OwnerReference, len(m.GetOwnerReferences()))
		for idx, v := range m.GetOwnerReferences() {

			if h, ok := interface{}(v).(clone.Cloner); ok {
				target.OwnerReferences[idx] = h.Clone().(*Metadata_OwnerReference)
			} else {
				target.OwnerReferences[idx] = proto.Clone(v).(*Metadata_OwnerReference)
			}

		}
	}

	return target
}

// Clone function
func (m *Metadata_OwnerReference) Clone() proto.Message {
	var target *Metadata_OwnerReference
	if m == nil {
		return target
	}
	target = &Metadata_OwnerReference{}

	target.ApiVersion = m.GetApiVersion()

	if h, ok := interface{}(m.GetBlockOwnerDeletion()).(clone.Cloner); ok {
		target.BlockOwnerDeletion = h.Clone().(*github_com_golang_protobuf_ptypes_wrappers.BoolValue)
	} else {
		target.BlockOwnerDeletion = proto.Clone(m.GetBlockOwnerDeletion()).(*github_com_golang_protobuf_ptypes_wrappers.BoolValue)
	}

	if h, ok := interface{}(m.GetController()).(clone.Cloner); ok {
		target.Controller = h.Clone().(*github_com_golang_protobuf_ptypes_wrappers.BoolValue)
	} else {
		target.Controller = proto.Clone(m.GetController()).(*github_com_golang_protobuf_ptypes_wrappers.BoolValue)
	}

	target.Kind = m.GetKind()

	target.Name = m.GetName()

	target.Uid = m.GetUid()

	return target
}
