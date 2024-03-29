// Code generated by MockGen. DO NOT EDIT.
// Source: test/mocks/v1alpha1/mock_resource_client.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1alpha1 "github.com/solo-io/solo-kit/test/mocks/v1alpha1"
)

// MockMockResourceWatcher is a mock of MockResourceWatcher interface
type MockMockResourceWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockMockResourceWatcherMockRecorder
}

// MockMockResourceWatcherMockRecorder is the mock recorder for MockMockResourceWatcher
type MockMockResourceWatcherMockRecorder struct {
	mock *MockMockResourceWatcher
}

// NewMockMockResourceWatcher creates a new mock instance
func NewMockMockResourceWatcher(ctrl *gomock.Controller) *MockMockResourceWatcher {
	mock := &MockMockResourceWatcher{ctrl: ctrl}
	mock.recorder = &MockMockResourceWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMockResourceWatcher) EXPECT() *MockMockResourceWatcherMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockMockResourceWatcher) Watch(namespace string, opts clients.WatchOpts) (<-chan v1alpha1.MockResourceList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan v1alpha1.MockResourceList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockMockResourceWatcherMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockMockResourceWatcher)(nil).Watch), namespace, opts)
}

// MockMockResourceClient is a mock of MockResourceClient interface
type MockMockResourceClient struct {
	ctrl     *gomock.Controller
	recorder *MockMockResourceClientMockRecorder
}

// MockMockResourceClientMockRecorder is the mock recorder for MockMockResourceClient
type MockMockResourceClientMockRecorder struct {
	mock *MockMockResourceClient
}

// NewMockMockResourceClient creates a new mock instance
func NewMockMockResourceClient(ctrl *gomock.Controller) *MockMockResourceClient {
	mock := &MockMockResourceClient{ctrl: ctrl}
	mock.recorder = &MockMockResourceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMockResourceClient) EXPECT() *MockMockResourceClientMockRecorder {
	return m.recorder
}

// BaseClient mocks base method
func (m *MockMockResourceClient) BaseClient() clients.ResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseClient")
	ret0, _ := ret[0].(clients.ResourceClient)
	return ret0
}

// BaseClient indicates an expected call of BaseClient
func (mr *MockMockResourceClientMockRecorder) BaseClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseClient", reflect.TypeOf((*MockMockResourceClient)(nil).BaseClient))
}

// Register mocks base method
func (m *MockMockResourceClient) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockMockResourceClientMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockMockResourceClient)(nil).Register))
}

// Read mocks base method
func (m *MockMockResourceClient) Read(namespace, name string, opts clients.ReadOpts) (*v1alpha1.MockResource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", namespace, name, opts)
	ret0, _ := ret[0].(*v1alpha1.MockResource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockMockResourceClientMockRecorder) Read(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockMockResourceClient)(nil).Read), namespace, name, opts)
}

// Write mocks base method
func (m *MockMockResourceClient) Write(resource *v1alpha1.MockResource, opts clients.WriteOpts) (*v1alpha1.MockResource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", resource, opts)
	ret0, _ := ret[0].(*v1alpha1.MockResource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write
func (mr *MockMockResourceClientMockRecorder) Write(resource, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockMockResourceClient)(nil).Write), resource, opts)
}

// Delete mocks base method
func (m *MockMockResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", namespace, name, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockMockResourceClientMockRecorder) Delete(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockMockResourceClient)(nil).Delete), namespace, name, opts)
}

// List mocks base method
func (m *MockMockResourceClient) List(namespace string, opts clients.ListOpts) (v1alpha1.MockResourceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", namespace, opts)
	ret0, _ := ret[0].(v1alpha1.MockResourceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockMockResourceClientMockRecorder) List(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockMockResourceClient)(nil).List), namespace, opts)
}

// Watch mocks base method
func (m *MockMockResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan v1alpha1.MockResourceList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan v1alpha1.MockResourceList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockMockResourceClientMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockMockResourceClient)(nil).Watch), namespace, opts)
}
