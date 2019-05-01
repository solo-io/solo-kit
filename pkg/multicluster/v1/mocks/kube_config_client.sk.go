// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/multicluster/v1/kube_config_client.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
)

// MockKubeConfigClient is a mock of KubeConfigClient interface
type MockKubeConfigClient struct {
	ctrl     *gomock.Controller
	recorder *MockKubeConfigClientMockRecorder
}

// MockKubeConfigClientMockRecorder is the mock recorder for MockKubeConfigClient
type MockKubeConfigClientMockRecorder struct {
	mock *MockKubeConfigClient
}

// NewMockKubeConfigClient creates a new mock instance
func NewMockKubeConfigClient(ctrl *gomock.Controller) *MockKubeConfigClient {
	mock := &MockKubeConfigClient{ctrl: ctrl}
	mock.recorder = &MockKubeConfigClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockKubeConfigClient) EXPECT() *MockKubeConfigClientMockRecorder {
	return m.recorder
}

// BaseClient mocks base method
func (m *MockKubeConfigClient) BaseClient() clients.ResourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseClient")
	ret0, _ := ret[0].(clients.ResourceClient)
	return ret0
}

// BaseClient indicates an expected call of BaseClient
func (mr *MockKubeConfigClientMockRecorder) BaseClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseClient", reflect.TypeOf((*MockKubeConfigClient)(nil).BaseClient))
}

// Register mocks base method
func (m *MockKubeConfigClient) Register() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register")
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockKubeConfigClientMockRecorder) Register() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockKubeConfigClient)(nil).Register))
}

// Read mocks base method
func (m *MockKubeConfigClient) Read(namespace, name string, opts clients.ReadOpts) (*v1.KubeConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", namespace, name, opts)
	ret0, _ := ret[0].(*v1.KubeConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockKubeConfigClientMockRecorder) Read(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockKubeConfigClient)(nil).Read), namespace, name, opts)
}

// Write mocks base method
func (m *MockKubeConfigClient) Write(resource *v1.KubeConfig, opts clients.WriteOpts) (*v1.KubeConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", resource, opts)
	ret0, _ := ret[0].(*v1.KubeConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write
func (mr *MockKubeConfigClientMockRecorder) Write(resource, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockKubeConfigClient)(nil).Write), resource, opts)
}

// Delete mocks base method
func (m *MockKubeConfigClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", namespace, name, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockKubeConfigClientMockRecorder) Delete(namespace, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockKubeConfigClient)(nil).Delete), namespace, name, opts)
}

// List mocks base method
func (m *MockKubeConfigClient) List(namespace string, opts clients.ListOpts) (v1.KubeConfigList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", namespace, opts)
	ret0, _ := ret[0].(v1.KubeConfigList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockKubeConfigClientMockRecorder) List(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockKubeConfigClient)(nil).List), namespace, opts)
}

// Watch mocks base method
func (m *MockKubeConfigClient) Watch(namespace string, opts clients.WatchOpts) (<-chan v1.KubeConfigList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan v1.KubeConfigList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockKubeConfigClientMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockKubeConfigClient)(nil).Watch), namespace, opts)
}
