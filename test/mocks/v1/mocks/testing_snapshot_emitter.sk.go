// Code generated by MockGen. DO NOT EDIT.
// Source: ./test/mocks/v1/testing_snapshot_emitter.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	reflect "reflect"
)

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
func (m *MockTestingEmitter) MockResource() v1.MockResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MockResource")
	ret0, _ := ret[0].(v1.MockResourceClient)
	return ret0
}

// MockResource indicates an expected call of MockResource
func (mr *MockTestingEmitterMockRecorder) MockResource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MockResource", reflect.TypeOf((*MockTestingEmitter)(nil).MockResource))
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

// AnotherMockResource mocks base method
func (m *MockTestingEmitter) AnotherMockResource() v1.AnotherMockResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AnotherMockResource")
	ret0, _ := ret[0].(v1.AnotherMockResourceClient)
	return ret0
}

// AnotherMockResource indicates an expected call of AnotherMockResource
func (mr *MockTestingEmitterMockRecorder) AnotherMockResource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AnotherMockResource", reflect.TypeOf((*MockTestingEmitter)(nil).AnotherMockResource))
}

// ClusterResource mocks base method
func (m *MockTestingEmitter) ClusterResource() v1.ClusterResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClusterResource")
	ret0, _ := ret[0].(v1.ClusterResourceClient)
	return ret0
}

// ClusterResource indicates an expected call of ClusterResource
func (mr *MockTestingEmitterMockRecorder) ClusterResource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClusterResource", reflect.TypeOf((*MockTestingEmitter)(nil).ClusterResource))
}

// MockCustomType mocks base method
func (m *MockTestingEmitter) MockCustomType() v1.MockCustomTypeClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MockCustomType")
	ret0, _ := ret[0].(v1.MockCustomTypeClient)
	return ret0
}

// MockCustomType indicates an expected call of MockCustomType
func (mr *MockTestingEmitterMockRecorder) MockCustomType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MockCustomType", reflect.TypeOf((*MockTestingEmitter)(nil).MockCustomType))
}

// Pod mocks base method
func (m *MockTestingEmitter) Pod() kubernetes.PodClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pod")
	ret0, _ := ret[0].(kubernetes.PodClient)
	return ret0
}

// Pod indicates an expected call of Pod
func (mr *MockTestingEmitterMockRecorder) Pod() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pod", reflect.TypeOf((*MockTestingEmitter)(nil).Pod))
}

// Snapshots mocks base method
func (m *MockTestingEmitter) Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *v1.TestingSnapshot, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Snapshots", watchNamespaces, opts)
	ret0, _ := ret[0].(<-chan *v1.TestingSnapshot)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Snapshots indicates an expected call of Snapshots
func (mr *MockTestingEmitterMockRecorder) Snapshots(watchNamespaces, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Snapshots", reflect.TypeOf((*MockTestingEmitter)(nil).Snapshots), watchNamespaces, opts)
}
