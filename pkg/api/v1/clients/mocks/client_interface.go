// Code generated by MockGen. DO NOT EDIT.
// Source: client_interface.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	resources "github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// MockResourceWatcher is a mock of ResourceWatcher interface.
type MockResourceWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockResourceWatcherMockRecorder
}

// MockResourceWatcherMockRecorder is the mock recorder for MockResourceWatcher.
type MockResourceWatcherMockRecorder struct {
	mock *MockResourceWatcher
}

// NewMockResourceWatcher creates a new mock instance.
func NewMockResourceWatcher(ctrl *gomock.Controller) *MockResourceWatcher {
	mock := &MockResourceWatcher{ctrl: ctrl}
	mock.recorder = &MockResourceWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResourceWatcher) EXPECT() *MockResourceWatcherMockRecorder {
	return m.recorder
}

// Watch mocks base method.
func (m *MockResourceWatcher) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan resources.ResourceList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch.
func (mr *MockResourceWatcherMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockResourceWatcher)(nil).Watch), namespace, opts)
}

// MockResourceClient is a mock of ResourceClient interface.
type MockResourceClient struct {
	ctrl     *gomock.Controller
	recorder *MockResourceClientMockRecorder
}

// MockResourceClientMockRecorder is the mock recorder for MockResourceClient.
type MockResourceClientMockRecorder struct {
	mock *MockResourceClient
}

// NewMockResourceClient creates a new mock instance.
func NewMockResourceClient(ctrl *gomock.Controller) *MockResourceClient {
	mock := &MockResourceClient{ctrl: ctrl}
	mock.recorder = &MockResourceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResourceClient) EXPECT() *MockResourceClientMockRecorder {
	return m.recorder
}

// Kind mocks base method.
func (m *MockResourceClient) Kind() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Kind")
	ret0, _ := ret[0].(string)
	return ret0
}

// Kind indicates an expected call of Kind.
func (mr *MockResourceClientMockRecorder) Kind() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kind", reflect.TypeOf((*MockResourceClient)(nil).Kind))
}

// NewResource mocks base method.
func (m *MockResourceClient) NewResource() resources.Resource {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewResource")
	ret0, _ := ret[0].(resources.Resource)
	return ret0
}

// NewResource indicates an expected call of NewResource.
func (mr *MockResourceClientMockRecorder) NewResource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewResource", reflect.TypeOf((*MockResourceClient)(nil).NewResource))
}

// Register mocks base method.
func (m *MockResourceClient) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockResourceClientMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockResourceClient)(nil).Register))
}

// Read mocks base method.
func (m *MockResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", namespace, name, opts)
	ret0, _ := ret[0].(resources.Resource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockResourceClientMockRecorder) Read(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockResourceClient)(nil).Read), namespace, name, opts)
}

// Write mocks base method.
func (m *MockResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", resource, opts)
	ret0, _ := ret[0].(resources.Resource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockResourceClientMockRecorder) Write(resource, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockResourceClient)(nil).Write), resource, opts)
}

// Delete mocks base method.
func (m *MockResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", namespace, name, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockResourceClientMockRecorder) Delete(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockResourceClient)(nil).Delete), namespace, name, opts)
}

// List mocks base method.
func (m *MockResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", namespace, opts)
	ret0, _ := ret[0].(resources.ResourceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockResourceClientMockRecorder) List(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockResourceClient)(nil).List), namespace, opts)
}

// Watch mocks base method.
func (m *MockResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan resources.ResourceList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch.
func (mr *MockResourceClientMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockResourceClient)(nil).Watch), namespace, opts)
}

// MockStorageWriteOpts is a mock of StorageWriteOpts interface.
type MockStorageWriteOpts struct {
	ctrl     *gomock.Controller
	recorder *MockStorageWriteOptsMockRecorder
}

// MockStorageWriteOptsMockRecorder is the mock recorder for MockStorageWriteOpts.
type MockStorageWriteOptsMockRecorder struct {
	mock *MockStorageWriteOpts
}

// NewMockStorageWriteOpts creates a new mock instance.
func NewMockStorageWriteOpts(ctrl *gomock.Controller) *MockStorageWriteOpts {
	mock := &MockStorageWriteOpts{ctrl: ctrl}
	mock.recorder = &MockStorageWriteOptsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageWriteOpts) EXPECT() *MockStorageWriteOptsMockRecorder {
	return m.recorder
}

// StorageWriteOptsTag mocks base method.
func (m *MockStorageWriteOpts) StorageWriteOptsTag() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StorageWriteOptsTag")
}

// StorageWriteOptsTag indicates an expected call of StorageWriteOptsTag.
func (mr *MockStorageWriteOptsMockRecorder) StorageWriteOptsTag() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StorageWriteOptsTag", reflect.TypeOf((*MockStorageWriteOpts)(nil).StorageWriteOptsTag))
}
