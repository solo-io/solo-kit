// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/eitanya/go/src/github.com/solo-io/solo-kit/test/mocks/v1/cluster_resource_reconciler.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	reflect "reflect"
)

// MockClusterResourceReconciler is a mock of ClusterResourceReconciler interface
type MockClusterResourceReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockClusterResourceReconcilerMockRecorder
}

// MockClusterResourceReconcilerMockRecorder is the mock recorder for MockClusterResourceReconciler
type MockClusterResourceReconcilerMockRecorder struct {
	mock *MockClusterResourceReconciler
}

// NewMockClusterResourceReconciler creates a new mock instance
func NewMockClusterResourceReconciler(ctrl *gomock.Controller) *MockClusterResourceReconciler {
	mock := &MockClusterResourceReconciler{ctrl: ctrl}
	mock.recorder = &MockClusterResourceReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClusterResourceReconciler) EXPECT() *MockClusterResourceReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method
func (m *MockClusterResourceReconciler) Reconcile(namespace string, desiredResources v1.ClusterResourceList, transition v1.TransitionClusterResourceFunc, opts clients.ListOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", namespace, desiredResources, transition, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile
func (mr *MockClusterResourceReconcilerMockRecorder) Reconcile(namespace, desiredResources, transition, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockClusterResourceReconciler)(nil).Reconcile), namespace, desiredResources, transition, opts)
}
