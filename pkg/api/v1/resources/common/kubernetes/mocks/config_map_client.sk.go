// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/api/v1/resources/common/kubernetes/config_map_client.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

// MockConfigMapWatcher is a mock of ConfigMapWatcher interface
type MockConfigMapWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockConfigMapWatcherMockRecorder
}

// MockConfigMapWatcherMockRecorder is the mock recorder for MockConfigMapWatcher
type MockConfigMapWatcherMockRecorder struct {
	mock *MockConfigMapWatcher
}

// NewMockConfigMapWatcher creates a new mock instance
func NewMockConfigMapWatcher(ctrl *gomock.Controller) *MockConfigMapWatcher {
	mock := &MockConfigMapWatcher{ctrl: ctrl}
	mock.recorder = &MockConfigMapWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfigMapWatcher) EXPECT() *MockConfigMapWatcherMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockConfigMapWatcher) Watch(namespace string, opts clients.WatchOpts) (<-chan kubernetes.ConfigMapList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan kubernetes.ConfigMapList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockConfigMapWatcherMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockConfigMapWatcher)(nil).Watch), namespace, opts)
}

// MockConfigMapClient is a mock of ConfigMapClient interface
type MockConfigMapClient struct {
	ctrl     *gomock.Controller
	recorder *MockConfigMapClientMockRecorder
}

// MockConfigMapClientMockRecorder is the mock recorder for MockConfigMapClient
type MockConfigMapClientMockRecorder struct {
	mock *MockConfigMapClient
}

// NewMockConfigMapClient creates a new mock instance
func NewMockConfigMapClient(ctrl *gomock.Controller) *MockConfigMapClient {
	mock := &MockConfigMapClient{ctrl: ctrl}
	mock.recorder = &MockConfigMapClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfigMapClient) EXPECT() *MockConfigMapClientMockRecorder {
	return m.recorder
}

// BaseClient mocks base method
func (m *MockConfigMapClient) BaseClient() clients.ResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseClient")
	ret0, _ := ret[0].(clients.ResourceClient)
	return ret0
}

// BaseClient indicates an expected call of BaseClient
func (mr *MockConfigMapClientMockRecorder) BaseClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseClient", reflect.TypeOf((*MockConfigMapClient)(nil).BaseClient))
}

// Register mocks base method
func (m *MockConfigMapClient) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockConfigMapClientMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockConfigMapClient)(nil).Register))
}

// Read mocks base method
func (m *MockConfigMapClient) Read(namespace, name string, opts clients.ReadOpts) (*kubernetes.ConfigMap, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", namespace, name, opts)
	ret0, _ := ret[0].(*kubernetes.ConfigMap)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockConfigMapClientMockRecorder) Read(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockConfigMapClient)(nil).Read), namespace, name, opts)
}

// Write mocks base method
func (m *MockConfigMapClient) Write(resource *kubernetes.ConfigMap, opts clients.WriteOpts) (*kubernetes.ConfigMap, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", resource, opts)
	ret0, _ := ret[0].(*kubernetes.ConfigMap)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write
func (mr *MockConfigMapClientMockRecorder) Write(resource, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockConfigMapClient)(nil).Write), resource, opts)
}

// Delete mocks base method
func (m *MockConfigMapClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", namespace, name, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockConfigMapClientMockRecorder) Delete(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockConfigMapClient)(nil).Delete), namespace, name, opts)
}

// List mocks base method
func (m *MockConfigMapClient) List(namespace string, opts clients.ListOpts) (kubernetes.ConfigMapList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", namespace, opts)
	ret0, _ := ret[0].(kubernetes.ConfigMapList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockConfigMapClientMockRecorder) List(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockConfigMapClient)(nil).List), namespace, opts)
}

// Watch mocks base method
func (m *MockConfigMapClient) Watch(namespace string, opts clients.WatchOpts) (<-chan kubernetes.ConfigMapList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan kubernetes.ConfigMapList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockConfigMapClientMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockConfigMapClient)(nil).Watch), namespace, opts)
}
