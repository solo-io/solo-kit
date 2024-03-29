// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/api/v1/resources/common/kubernetes/job_client.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

// MockJobWatcher is a mock of JobWatcher interface
type MockJobWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockJobWatcherMockRecorder
}

// MockJobWatcherMockRecorder is the mock recorder for MockJobWatcher
type MockJobWatcherMockRecorder struct {
	mock *MockJobWatcher
}

// NewMockJobWatcher creates a new mock instance
func NewMockJobWatcher(ctrl *gomock.Controller) *MockJobWatcher {
	mock := &MockJobWatcher{ctrl: ctrl}
	mock.recorder = &MockJobWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockJobWatcher) EXPECT() *MockJobWatcherMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockJobWatcher) Watch(namespace string, opts clients.WatchOpts) (<-chan kubernetes.JobList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan kubernetes.JobList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockJobWatcherMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockJobWatcher)(nil).Watch), namespace, opts)
}

// MockJobClient is a mock of JobClient interface
type MockJobClient struct {
	ctrl     *gomock.Controller
	recorder *MockJobClientMockRecorder
}

// MockJobClientMockRecorder is the mock recorder for MockJobClient
type MockJobClientMockRecorder struct {
	mock *MockJobClient
}

// NewMockJobClient creates a new mock instance
func NewMockJobClient(ctrl *gomock.Controller) *MockJobClient {
	mock := &MockJobClient{ctrl: ctrl}
	mock.recorder = &MockJobClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockJobClient) EXPECT() *MockJobClientMockRecorder {
	return m.recorder
}

// BaseClient mocks base method
func (m *MockJobClient) BaseClient() clients.ResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseClient")
	ret0, _ := ret[0].(clients.ResourceClient)
	return ret0
}

// BaseClient indicates an expected call of BaseClient
func (mr *MockJobClientMockRecorder) BaseClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseClient", reflect.TypeOf((*MockJobClient)(nil).BaseClient))
}

// Register mocks base method
func (m *MockJobClient) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockJobClientMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockJobClient)(nil).Register))
}

// Read mocks base method
func (m *MockJobClient) Read(namespace, name string, opts clients.ReadOpts) (*kubernetes.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", namespace, name, opts)
	ret0, _ := ret[0].(*kubernetes.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockJobClientMockRecorder) Read(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockJobClient)(nil).Read), namespace, name, opts)
}

// Write mocks base method
func (m *MockJobClient) Write(resource *kubernetes.Job, opts clients.WriteOpts) (*kubernetes.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", resource, opts)
	ret0, _ := ret[0].(*kubernetes.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write
func (mr *MockJobClientMockRecorder) Write(resource, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockJobClient)(nil).Write), resource, opts)
}

// Delete mocks base method
func (m *MockJobClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", namespace, name, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockJobClientMockRecorder) Delete(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockJobClient)(nil).Delete), namespace, name, opts)
}

// List mocks base method
func (m *MockJobClient) List(namespace string, opts clients.ListOpts) (kubernetes.JobList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", namespace, opts)
	ret0, _ := ret[0].(kubernetes.JobList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockJobClientMockRecorder) List(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockJobClient)(nil).List), namespace, opts)
}

// Watch mocks base method
func (m *MockJobClient) Watch(namespace string, opts clients.WatchOpts) (<-chan kubernetes.JobList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan kubernetes.JobList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockJobClientMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockJobClient)(nil).Watch), namespace, opts)
}
