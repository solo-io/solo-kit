// Code generated by MockGen. DO NOT EDIT.
// Source: test/mocks/v1/another_mock_resource_client.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

// MockAnotherMockResourceWatcher is a mock of AnotherMockResourceWatcher interface
type MockAnotherMockResourceWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockAnotherMockResourceWatcherMockRecorder
}

// MockAnotherMockResourceWatcherMockRecorder is the mock recorder for MockAnotherMockResourceWatcher
type MockAnotherMockResourceWatcherMockRecorder struct {
	mock *MockAnotherMockResourceWatcher
}

// NewMockAnotherMockResourceWatcher creates a new mock instance
func NewMockAnotherMockResourceWatcher(ctrl *gomock.Controller) *MockAnotherMockResourceWatcher {
	mock := &MockAnotherMockResourceWatcher{ctrl: ctrl}
	mock.recorder = &MockAnotherMockResourceWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAnotherMockResourceWatcher) EXPECT() *MockAnotherMockResourceWatcherMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockAnotherMockResourceWatcher) Watch(namespace string, opts clients.WatchOpts) (<-chan v1.AnotherMockResourceList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan v1.AnotherMockResourceList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockAnotherMockResourceWatcherMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockAnotherMockResourceWatcher)(nil).Watch), namespace, opts)
}

// MockAnotherMockResourceClient is a mock of AnotherMockResourceClient interface
type MockAnotherMockResourceClient struct {
	ctrl     *gomock.Controller
	recorder *MockAnotherMockResourceClientMockRecorder
}

// MockAnotherMockResourceClientMockRecorder is the mock recorder for MockAnotherMockResourceClient
type MockAnotherMockResourceClientMockRecorder struct {
	mock *MockAnotherMockResourceClient
}

// NewMockAnotherMockResourceClient creates a new mock instance
func NewMockAnotherMockResourceClient(ctrl *gomock.Controller) *MockAnotherMockResourceClient {
	mock := &MockAnotherMockResourceClient{ctrl: ctrl}
	mock.recorder = &MockAnotherMockResourceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAnotherMockResourceClient) EXPECT() *MockAnotherMockResourceClientMockRecorder {
	return m.recorder
}

// BaseClient mocks base method
func (m *MockAnotherMockResourceClient) BaseClient() clients.ResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseClient")
	ret0, _ := ret[0].(clients.ResourceClient)
	return ret0
}

// BaseClient indicates an expected call of BaseClient
func (mr *MockAnotherMockResourceClientMockRecorder) BaseClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseClient", reflect.TypeOf((*MockAnotherMockResourceClient)(nil).BaseClient))
}

// Register mocks base method
func (m *MockAnotherMockResourceClient) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockAnotherMockResourceClientMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockAnotherMockResourceClient)(nil).Register))
}

// Read mocks base method
func (m *MockAnotherMockResourceClient) Read(namespace, name string, opts clients.ReadOpts) (*v1.AnotherMockResource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", namespace, name, opts)
	ret0, _ := ret[0].(*v1.AnotherMockResource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockAnotherMockResourceClientMockRecorder) Read(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockAnotherMockResourceClient)(nil).Read), namespace, name, opts)
}

// Write mocks base method
func (m *MockAnotherMockResourceClient) Write(resource *v1.AnotherMockResource, opts clients.WriteOpts) (*v1.AnotherMockResource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", resource, opts)
	ret0, _ := ret[0].(*v1.AnotherMockResource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write
func (mr *MockAnotherMockResourceClientMockRecorder) Write(resource, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockAnotherMockResourceClient)(nil).Write), resource, opts)
}

// Delete mocks base method
func (m *MockAnotherMockResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", namespace, name, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockAnotherMockResourceClientMockRecorder) Delete(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockAnotherMockResourceClient)(nil).Delete), namespace, name, opts)
}

// List mocks base method
func (m *MockAnotherMockResourceClient) List(namespace string, opts clients.ListOpts) (v1.AnotherMockResourceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", namespace, opts)
	ret0, _ := ret[0].(v1.AnotherMockResourceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockAnotherMockResourceClientMockRecorder) List(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockAnotherMockResourceClient)(nil).List), namespace, opts)
}

// Watch mocks base method
func (m *MockAnotherMockResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan v1.AnotherMockResourceList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan v1.AnotherMockResourceList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockAnotherMockResourceClientMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockAnotherMockResourceClient)(nil).Watch), namespace, opts)
}
