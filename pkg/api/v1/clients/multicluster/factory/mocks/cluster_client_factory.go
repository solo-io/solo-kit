// Code generated by MockGen. DO NOT EDIT.
// Source: cluster_client_factory.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	rest "k8s.io/client-go/rest"
)

// MockClusterClientFactory is a mock of ClusterClientFactory interface
type MockClusterClientFactory struct {
	ctrl     *gomock.Controller
	recorder *MockClusterClientFactoryMockRecorder
}

// MockClusterClientFactoryMockRecorder is the mock recorder for MockClusterClientFactory
type MockClusterClientFactoryMockRecorder struct {
	mock *MockClusterClientFactory
}

// NewMockClusterClientFactory creates a new mock instance
func NewMockClusterClientFactory(ctrl *gomock.Controller) *MockClusterClientFactory {
	mock := &MockClusterClientFactory{ctrl: ctrl}
	mock.recorder = &MockClusterClientFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClusterClientFactory) EXPECT() *MockClusterClientFactoryMockRecorder {
	return m.recorder
}

// GetClient mocks base method
func (m *MockClusterClientFactory) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClient", cluster, restConfig)
	ret0, _ := ret[0].(clients.ResourceClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClient indicates an expected call of GetClient
func (mr *MockClusterClientFactoryMockRecorder) GetClient(cluster, restConfig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClient", reflect.TypeOf((*MockClusterClientFactory)(nil).GetClient), cluster, restConfig)
}
