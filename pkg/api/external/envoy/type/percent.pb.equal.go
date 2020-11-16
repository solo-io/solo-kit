// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/type/percent.proto

package _type

import (
	"errors"
	"fmt"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
)

// Equal function
func (m *Percent) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Percent)
	if !ok {
		that2, ok := that.(Percent)
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

	if m.GetValue() != target.GetValue() {
		return false
	}

	return true
}

// Equal function
func (m *FractionalPercent) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*FractionalPercent)
	if !ok {
		that2, ok := that.(FractionalPercent)
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

	if m.GetNumerator() != target.GetNumerator() {
		return false
	}

	if m.GetDenominator() != target.GetDenominator() {
		return false
	}

	return true
}
