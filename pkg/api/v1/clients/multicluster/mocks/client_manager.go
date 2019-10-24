// Code generated by MockGen. DO NOT EDIT.
// Source: client_manager.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	rest "k8s.io/client-go/rest"
)

// MockClusterClientHandler is a mock of ClusterClientHandler interface
type MockClusterClientHandler struct {
	ctrl     *gomock.Controller
	recorder *MockClusterClientHandlerMockRecorder
}

// MockClusterClientHandlerMockRecorder is the mock recorder for MockClusterClientHandler
type MockClusterClientHandlerMockRecorder struct {
	mock *MockClusterClientHandler
}

// NewMockClusterClientHandler creates a new mock instance
func NewMockClusterClientHandler(ctrl *gomock.Controller) *MockClusterClientHandler {
	mock := &MockClusterClientHandler{ctrl: ctrl}
	mock.recorder = &MockClusterClientHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClusterClientHandler) EXPECT() *MockClusterClientHandlerMockRecorder {
	return m.recorder
}

// HandleNewClusterClient mocks base method
func (m *MockClusterClientHandler) HandleNewClusterClient(cluster string, client clients.ResourceClient) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandleNewClusterClient", cluster, client)
}

// HandleNewClusterClient indicates an expected call of HandleNewClusterClient
func (mr *MockClusterClientHandlerMockRecorder) HandleNewClusterClient(cluster, client interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleNewClusterClient", reflect.TypeOf((*MockClusterClientHandler)(nil).HandleNewClusterClient), cluster, client)
}

// HandleRemovedClusterClient mocks base method
func (m *MockClusterClientHandler) HandleRemovedClusterClient(cluster string, client clients.ResourceClient) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandleRemovedClusterClient", cluster, client)
}

// HandleRemovedClusterClient indicates an expected call of HandleRemovedClusterClient
func (mr *MockClusterClientHandlerMockRecorder) HandleRemovedClusterClient(cluster, client interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleRemovedClusterClient", reflect.TypeOf((*MockClusterClientHandler)(nil).HandleRemovedClusterClient), cluster, client)
}

// MockClusterClientManager is a mock of ClusterClientManager interface
type MockClusterClientManager struct {
	ctrl     *gomock.Controller
	recorder *MockClusterClientManagerMockRecorder
}

// MockClusterClientManagerMockRecorder is the mock recorder for MockClusterClientManager
type MockClusterClientManagerMockRecorder struct {
	mock *MockClusterClientManager
}

// NewMockClusterClientManager creates a new mock instance
func NewMockClusterClientManager(ctrl *gomock.Controller) *MockClusterClientManager {
	mock := &MockClusterClientManager{ctrl: ctrl}
	mock.recorder = &MockClusterClientManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClusterClientManager) EXPECT() *MockClusterClientManagerMockRecorder {
	return m.recorder
}

// ClusterAdded mocks base method
func (m *MockClusterClientManager) ClusterAdded(cluster string, restConfig *rest.Config) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ClusterAdded", cluster, restConfig)
}

// ClusterAdded indicates an expected call of ClusterAdded
func (mr *MockClusterClientManagerMockRecorder) ClusterAdded(cluster, restConfig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClusterAdded", reflect.TypeOf((*MockClusterClientManager)(nil).ClusterAdded), cluster, restConfig)
}

// ClusterRemoved mocks base method
func (m *MockClusterClientManager) ClusterRemoved(cluster string, restConfig *rest.Config) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ClusterRemoved", cluster, restConfig)
}

// ClusterRemoved indicates an expected call of ClusterRemoved
func (mr *MockClusterClientManagerMockRecorder) ClusterRemoved(cluster, restConfig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClusterRemoved", reflect.TypeOf((*MockClusterClientManager)(nil).ClusterRemoved), cluster, restConfig)
}

// ClientForCluster mocks base method
func (m *MockClusterClientManager) ClientForCluster(cluster string) (clients.ResourceClient, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClientForCluster", cluster)
	ret0, _ := ret[0].(clients.ResourceClient)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// ClientForCluster indicates an expected call of ClientForCluster
func (mr *MockClusterClientManagerMockRecorder) ClientForCluster(cluster interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClientForCluster", reflect.TypeOf((*MockClusterClientManager)(nil).ClientForCluster), cluster)
}
