// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/api/v1/resources/common/kubernetes/deployment_client.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

// MockDeploymentWatcher is a mock of DeploymentWatcher interface
type MockDeploymentWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockDeploymentWatcherMockRecorder
}

// MockDeploymentWatcherMockRecorder is the mock recorder for MockDeploymentWatcher
type MockDeploymentWatcherMockRecorder struct {
	mock *MockDeploymentWatcher
}

// NewMockDeploymentWatcher creates a new mock instance
func NewMockDeploymentWatcher(ctrl *gomock.Controller) *MockDeploymentWatcher {
	mock := &MockDeploymentWatcher{ctrl: ctrl}
	mock.recorder = &MockDeploymentWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeploymentWatcher) EXPECT() *MockDeploymentWatcherMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockDeploymentWatcher) Watch(namespace string, opts clients.WatchOpts) (<-chan kubernetes.DeploymentList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan kubernetes.DeploymentList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockDeploymentWatcherMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockDeploymentWatcher)(nil).Watch), namespace, opts)
}

// MockDeploymentClient is a mock of DeploymentClient interface
type MockDeploymentClient struct {
	ctrl     *gomock.Controller
	recorder *MockDeploymentClientMockRecorder
}

// MockDeploymentClientMockRecorder is the mock recorder for MockDeploymentClient
type MockDeploymentClientMockRecorder struct {
	mock *MockDeploymentClient
}

// NewMockDeploymentClient creates a new mock instance
func NewMockDeploymentClient(ctrl *gomock.Controller) *MockDeploymentClient {
	mock := &MockDeploymentClient{ctrl: ctrl}
	mock.recorder = &MockDeploymentClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeploymentClient) EXPECT() *MockDeploymentClientMockRecorder {
	return m.recorder
}

// BaseClient mocks base method
func (m *MockDeploymentClient) BaseClient() clients.ResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseClient")
	ret0, _ := ret[0].(clients.ResourceClient)
	return ret0
}

// BaseClient indicates an expected call of BaseClient
func (mr *MockDeploymentClientMockRecorder) BaseClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseClient", reflect.TypeOf((*MockDeploymentClient)(nil).BaseClient))
}

// Register mocks base method
func (m *MockDeploymentClient) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockDeploymentClientMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockDeploymentClient)(nil).Register))
}

// Read mocks base method
func (m *MockDeploymentClient) Read(namespace, name string, opts clients.ReadOpts) (*kubernetes.Deployment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", namespace, name, opts)
	ret0, _ := ret[0].(*kubernetes.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockDeploymentClientMockRecorder) Read(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockDeploymentClient)(nil).Read), namespace, name, opts)
}

// Write mocks base method
func (m *MockDeploymentClient) Write(resource *kubernetes.Deployment, opts clients.WriteOpts) (*kubernetes.Deployment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", resource, opts)
	ret0, _ := ret[0].(*kubernetes.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write
func (mr *MockDeploymentClientMockRecorder) Write(resource, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockDeploymentClient)(nil).Write), resource, opts)
}

// Delete mocks base method
func (m *MockDeploymentClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", namespace, name, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockDeploymentClientMockRecorder) Delete(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockDeploymentClient)(nil).Delete), namespace, name, opts)
}

// List mocks base method
func (m *MockDeploymentClient) List(namespace string, opts clients.ListOpts) (kubernetes.DeploymentList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", namespace, opts)
	ret0, _ := ret[0].(kubernetes.DeploymentList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockDeploymentClientMockRecorder) List(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockDeploymentClient)(nil).List), namespace, opts)
}

// Watch mocks base method
func (m *MockDeploymentClient) Watch(namespace string, opts clients.WatchOpts) (<-chan kubernetes.DeploymentList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan kubernetes.DeploymentList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockDeploymentClientMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockDeploymentClient)(nil).Watch), namespace, opts)
}
