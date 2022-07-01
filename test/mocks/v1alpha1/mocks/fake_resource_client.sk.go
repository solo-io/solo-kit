// Code generated by MockGen. DO NOT EDIT.
// Source: test/mocks/v1alpha1/fake_resource_client.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1alpha1 "github.com/solo-io/solo-kit/test/mocks/v1alpha1"
)

// MockFakeResourceWatcher is a mock of FakeResourceWatcher interface
type MockFakeResourceWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockFakeResourceWatcherMockRecorder
}

// MockFakeResourceWatcherMockRecorder is the mock recorder for MockFakeResourceWatcher
type MockFakeResourceWatcherMockRecorder struct {
	mock *MockFakeResourceWatcher
}

// NewMockFakeResourceWatcher creates a new mock instance
func NewMockFakeResourceWatcher(ctrl *gomock.Controller) *MockFakeResourceWatcher {
	mock := &MockFakeResourceWatcher{ctrl: ctrl}
	mock.recorder = &MockFakeResourceWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFakeResourceWatcher) EXPECT() *MockFakeResourceWatcherMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockFakeResourceWatcher) Watch(namespace string, opts clients.WatchOpts) (<-chan v1alpha1.FakeResourceList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan v1alpha1.FakeResourceList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockFakeResourceWatcherMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockFakeResourceWatcher)(nil).Watch), namespace, opts)
}

// MockFakeResourceClient is a mock of FakeResourceClient interface
type MockFakeResourceClient struct {
	ctrl     *gomock.Controller
	recorder *MockFakeResourceClientMockRecorder
}

// MockFakeResourceClientMockRecorder is the mock recorder for MockFakeResourceClient
type MockFakeResourceClientMockRecorder struct {
	mock *MockFakeResourceClient
}

// NewMockFakeResourceClient creates a new mock instance
func NewMockFakeResourceClient(ctrl *gomock.Controller) *MockFakeResourceClient {
	mock := &MockFakeResourceClient{ctrl: ctrl}
	mock.recorder = &MockFakeResourceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFakeResourceClient) EXPECT() *MockFakeResourceClientMockRecorder {
	return m.recorder
}

// BaseClient mocks base method
func (m *MockFakeResourceClient) BaseClient() clients.ResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseClient")
	ret0, _ := ret[0].(clients.ResourceClient)
	return ret0
}

// BaseClient indicates an expected call of BaseClient
func (mr *MockFakeResourceClientMockRecorder) BaseClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseClient", reflect.TypeOf((*MockFakeResourceClient)(nil).BaseClient))
}

// Register mocks base method
func (m *MockFakeResourceClient) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockFakeResourceClientMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockFakeResourceClient)(nil).Register))
}

// Read mocks base method
func (m *MockFakeResourceClient) Read(namespace, name string, opts clients.ReadOpts) (*v1alpha1.FakeResource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", namespace, name, opts)
	ret0, _ := ret[0].(*v1alpha1.FakeResource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockFakeResourceClientMockRecorder) Read(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockFakeResourceClient)(nil).Read), namespace, name, opts)
}

// Write mocks base method
func (m *MockFakeResourceClient) Write(resource *v1alpha1.FakeResource, opts clients.WriteOpts) (*v1alpha1.FakeResource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", resource, opts)
	ret0, _ := ret[0].(*v1alpha1.FakeResource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write
func (mr *MockFakeResourceClientMockRecorder) Write(resource, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockFakeResourceClient)(nil).Write), resource, opts)
}

// Delete mocks base method
func (m *MockFakeResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", namespace, name, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockFakeResourceClientMockRecorder) Delete(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockFakeResourceClient)(nil).Delete), namespace, name, opts)
}

// List mocks base method
func (m *MockFakeResourceClient) List(namespace string, opts clients.ListOpts) (v1alpha1.FakeResourceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", namespace, opts)
	ret0, _ := ret[0].(v1alpha1.FakeResourceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockFakeResourceClientMockRecorder) List(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockFakeResourceClient)(nil).List), namespace, opts)
}

// Watch mocks base method
func (m *MockFakeResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan v1alpha1.FakeResourceList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan v1alpha1.FakeResourceList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockFakeResourceClientMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockFakeResourceClient)(nil).Watch), namespace, opts)
}