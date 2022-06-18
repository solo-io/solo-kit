// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/api/v1/resources/common/kubernetes/service_reconciler.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

// MockServiceReconciler is a mock of ServiceReconciler interface
type MockServiceReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockServiceReconcilerMockRecorder
}

// MockServiceReconcilerMockRecorder is the mock recorder for MockServiceReconciler
type MockServiceReconcilerMockRecorder struct {
	mock *MockServiceReconciler
}

// NewMockServiceReconciler creates a new mock instance
func NewMockServiceReconciler(ctrl *gomock.Controller) *MockServiceReconciler {
	mock := &MockServiceReconciler{ctrl: ctrl}
	mock.recorder = &MockServiceReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockServiceReconciler) EXPECT() *MockServiceReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method
func (m *MockServiceReconciler) Reconcile(namespace string, desiredResources kubernetes.ServiceList, transition kubernetes.TransitionServiceFunc, opts clients.ListOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", namespace, desiredResources, transition, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile
func (mr *MockServiceReconcilerMockRecorder) Reconcile(namespace, desiredResources, transition, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockServiceReconciler)(nil).Reconcile), namespace, desiredResources, transition, opts)
}
