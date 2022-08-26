// Code generated by MockGen. DO NOT EDIT.
// Source: test/mocks/v2alpha1/testing_snapshot_emitter.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	v2alpha1 "github.com/solo-io/solo-kit/test/mocks/v2alpha1"
)

// MockTestingSnapshotEmitter is a mock of TestingSnapshotEmitter interface
type MockTestingSnapshotEmitter struct {
	ctrl     *gomock.Controller
	recorder *MockTestingSnapshotEmitterMockRecorder
}

// MockTestingSnapshotEmitterMockRecorder is the mock recorder for MockTestingSnapshotEmitter
type MockTestingSnapshotEmitterMockRecorder struct {
	mock *MockTestingSnapshotEmitter
}

// NewMockTestingSnapshotEmitter creates a new mock instance
func NewMockTestingSnapshotEmitter(ctrl *gomock.Controller) *MockTestingSnapshotEmitter {
	mock := &MockTestingSnapshotEmitter{ctrl: ctrl}
	mock.recorder = &MockTestingSnapshotEmitterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTestingSnapshotEmitter) EXPECT() *MockTestingSnapshotEmitterMockRecorder {
	return m.recorder
}

// Snapshots mocks base method
func (m *MockTestingSnapshotEmitter) Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *v2alpha1.TestingSnapshot, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Snapshots", watchNamespaces, opts)
	ret0, _ := ret[0].(<-chan *v2alpha1.TestingSnapshot)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Snapshots indicates an expected call of Snapshots
func (mr *MockTestingSnapshotEmitterMockRecorder) Snapshots(watchNamespaces, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Snapshots", reflect.TypeOf((*MockTestingSnapshotEmitter)(nil).Snapshots), watchNamespaces, opts)
}

// MockTestingEmitter is a mock of TestingEmitter interface
type MockTestingEmitter struct {
	ctrl     *gomock.Controller
	recorder *MockTestingEmitterMockRecorder
}

// MockTestingEmitterMockRecorder is the mock recorder for MockTestingEmitter
type MockTestingEmitterMockRecorder struct {
	mock *MockTestingEmitter
}

// NewMockTestingEmitter creates a new mock instance
func NewMockTestingEmitter(ctrl *gomock.Controller) *MockTestingEmitter {
	mock := &MockTestingEmitter{ctrl: ctrl}
	mock.recorder = &MockTestingEmitterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTestingEmitter) EXPECT() *MockTestingEmitterMockRecorder {
	return m.recorder
}

// Snapshots mocks base method
func (m *MockTestingEmitter) Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *v2alpha1.TestingSnapshot, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Snapshots", watchNamespaces, opts)
	ret0, _ := ret[0].(<-chan *v2alpha1.TestingSnapshot)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Snapshots indicates an expected call of Snapshots
func (mr *MockTestingEmitterMockRecorder) Snapshots(watchNamespaces, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Snapshots", reflect.TypeOf((*MockTestingEmitter)(nil).Snapshots), watchNamespaces, opts)
}

// Register mocks base method
func (m *MockTestingEmitter) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockTestingEmitterMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockTestingEmitter)(nil).Register))
}

// MockResource mocks base method
func (m *MockTestingEmitter) MockResource() v2alpha1.MockResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MockResource")
	ret0, _ := ret[0].(v2alpha1.MockResourceClient)
	return ret0
}

// MockResource indicates an expected call of MockResource
func (mr *MockTestingEmitterMockRecorder) MockResource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MockResource", reflect.TypeOf((*MockTestingEmitter)(nil).MockResource))
}

// FrequentlyChangingAnnotationsResource mocks base method
func (m *MockTestingEmitter) FrequentlyChangingAnnotationsResource() v2alpha1.FrequentlyChangingAnnotationsResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FrequentlyChangingAnnotationsResource")
	ret0, _ := ret[0].(v2alpha1.FrequentlyChangingAnnotationsResourceClient)
	return ret0
}

// FrequentlyChangingAnnotationsResource indicates an expected call of FrequentlyChangingAnnotationsResource
func (mr *MockTestingEmitterMockRecorder) FrequentlyChangingAnnotationsResource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FrequentlyChangingAnnotationsResource", reflect.TypeOf((*MockTestingEmitter)(nil).FrequentlyChangingAnnotationsResource))
}

// FakeResource mocks base method
func (m *MockTestingEmitter) FakeResource() v1.FakeResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FakeResource")
	ret0, _ := ret[0].(v1.FakeResourceClient)
	return ret0
}

// FakeResource indicates an expected call of FakeResource
func (mr *MockTestingEmitterMockRecorder) FakeResource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FakeResource", reflect.TypeOf((*MockTestingEmitter)(nil).FakeResource))
}