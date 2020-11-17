// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: metadata.proto

package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	equality "github.com/solo-io/protoc-gen-ext/pkg/equality"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
)

// Equal function
func (m *Metadata) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Metadata)
	if !ok {
		that2, ok := that.(Metadata)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetName(), target.GetName()) != 0 {
		return false
	}

	if strings.Compare(m.GetNamespace(), target.GetNamespace()) != 0 {
		return false
	}

	if strings.Compare(m.GetCluster(), target.GetCluster()) != 0 {
		return false
	}

	if strings.Compare(m.GetResourceVersion(), target.GetResourceVersion()) != 0 {
		return false
	}

	if len(m.GetLabels()) != len(target.GetLabels()) {
		return false
	}
	for k, v := range m.GetLabels() {

		if strings.Compare(v, target.GetLabels()[k]) != 0 {
			return false
		}

	}

	if len(m.GetAnnotations()) != len(target.GetAnnotations()) {
		return false
	}
	for k, v := range m.GetAnnotations() {

		if strings.Compare(v, target.GetAnnotations()[k]) != 0 {
			return false
		}

	}

	if m.GetGeneration() != target.GetGeneration() {
		return false
	}

	if len(m.GetOwnerReferences()) != len(target.GetOwnerReferences()) {
		return false
	}
	for idx, v := range m.GetOwnerReferences() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetOwnerReferences()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetOwnerReferences()[idx]) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *Metadata_OwnerReference) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Metadata_OwnerReference)
	if !ok {
		that2, ok := that.(Metadata_OwnerReference)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetApiVersion(), target.GetApiVersion()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetBlockOwnerDeletion()).(equality.Equalizer); ok {
		if !h.Equal(target.GetBlockOwnerDeletion()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetBlockOwnerDeletion(), target.GetBlockOwnerDeletion()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetController()).(equality.Equalizer); ok {
		if !h.Equal(target.GetController()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetController(), target.GetController()) {
			return false
		}
	}

	if strings.Compare(m.GetKind(), target.GetKind()) != 0 {
		return false
	}

	if strings.Compare(m.GetName(), target.GetName()) != 0 {
		return false
	}

	if strings.Compare(m.GetUid(), target.GetUid()) != 0 {
		return false
	}

	return true
}
