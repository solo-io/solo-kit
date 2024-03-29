// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/api/v1/resources/common/kubernetes/pod_client.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

// MockPodWatcher is a mock of PodWatcher interface
type MockPodWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockPodWatcherMockRecorder
}

// MockPodWatcherMockRecorder is the mock recorder for MockPodWatcher
type MockPodWatcherMockRecorder struct {
	mock *MockPodWatcher
}

// NewMockPodWatcher creates a new mock instance
func NewMockPodWatcher(ctrl *gomock.Controller) *MockPodWatcher {
	mock := &MockPodWatcher{ctrl: ctrl}
	mock.recorder = &MockPodWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPodWatcher) EXPECT() *MockPodWatcherMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockPodWatcher) Watch(namespace string, opts clients.WatchOpts) (<-chan kubernetes.PodList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan kubernetes.PodList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockPodWatcherMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockPodWatcher)(nil).Watch), namespace, opts)
}

// MockPodClient is a mock of PodClient interface
type MockPodClient struct {
	ctrl     *gomock.Controller
	recorder *MockPodClientMockRecorder
}

// MockPodClientMockRecorder is the mock recorder for MockPodClient
type MockPodClientMockRecorder struct {
	mock *MockPodClient
}

// NewMockPodClient creates a new mock instance
func NewMockPodClient(ctrl *gomock.Controller) *MockPodClient {
	mock := &MockPodClient{ctrl: ctrl}
	mock.recorder = &MockPodClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPodClient) EXPECT() *MockPodClientMockRecorder {
	return m.recorder
}

// BaseClient mocks base method
func (m *MockPodClient) BaseClient() clients.ResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseClient")
	ret0, _ := ret[0].(clients.ResourceClient)
	return ret0
}

// BaseClient indicates an expected call of BaseClient
func (mr *MockPodClientMockRecorder) BaseClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseClient", reflect.TypeOf((*MockPodClient)(nil).BaseClient))
}

// Register mocks base method
func (m *MockPodClient) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockPodClientMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockPodClient)(nil).Register))
}

// Read mocks base method
func (m *MockPodClient) Read(namespace, name string, opts clients.ReadOpts) (*kubernetes.Pod, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", namespace, name, opts)
	ret0, _ := ret[0].(*kubernetes.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockPodClientMockRecorder) Read(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockPodClient)(nil).Read), namespace, name, opts)
}

// Write mocks base method
func (m *MockPodClient) Write(resource *kubernetes.Pod, opts clients.WriteOpts) (*kubernetes.Pod, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", resource, opts)
	ret0, _ := ret[0].(*kubernetes.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write
func (mr *MockPodClientMockRecorder) Write(resource, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockPodClient)(nil).Write), resource, opts)
}

// Delete mocks base method
func (m *MockPodClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", namespace, name, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockPodClientMockRecorder) Delete(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockPodClient)(nil).Delete), namespace, name, opts)
}

// List mocks base method
func (m *MockPodClient) List(namespace string, opts clients.ListOpts) (kubernetes.PodList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", namespace, opts)
	ret0, _ := ret[0].(kubernetes.PodList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockPodClientMockRecorder) List(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockPodClient)(nil).List), namespace, opts)
}

// Watch mocks base method
func (m *MockPodClient) Watch(namespace string, opts clients.WatchOpts) (<-chan kubernetes.PodList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan kubernetes.PodList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockPodClientMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockPodClient)(nil).Watch), namespace, opts)
}
