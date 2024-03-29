// Code generated by MockGen. DO NOT EDIT.
// Source: test/mocks/v1/another_mock_resource_reconciler.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

// MockAnotherMockResourceReconciler is a mock of AnotherMockResourceReconciler interface
type MockAnotherMockResourceReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockAnotherMockResourceReconcilerMockRecorder
}

// MockAnotherMockResourceReconcilerMockRecorder is the mock recorder for MockAnotherMockResourceReconciler
type MockAnotherMockResourceReconcilerMockRecorder struct {
	mock *MockAnotherMockResourceReconciler
}

// NewMockAnotherMockResourceReconciler creates a new mock instance
func NewMockAnotherMockResourceReconciler(ctrl *gomock.Controller) *MockAnotherMockResourceReconciler {
	mock := &MockAnotherMockResourceReconciler{ctrl: ctrl}
	mock.recorder = &MockAnotherMockResourceReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAnotherMockResourceReconciler) EXPECT() *MockAnotherMockResourceReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method
func (m *MockAnotherMockResourceReconciler) Reconcile(namespace string, desiredResources v1.AnotherMockResourceList, transition v1.TransitionAnotherMockResourceFunc, opts clients.ListOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", namespace, desiredResources, transition, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile
func (mr *MockAnotherMockResourceReconcilerMockRecorder) Reconcile(namespace, desiredResources, transition, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockAnotherMockResourceReconciler)(nil).Reconcile), namespace, desiredResources, transition, opts)
}
