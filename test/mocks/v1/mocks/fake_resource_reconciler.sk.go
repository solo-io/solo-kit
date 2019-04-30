// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/eitanya/go/src/github.com/solo-io/solo-kit/test/mocks/v1/fake_resource_reconciler.sk.go

// Package mock_v1 is a generated GoMock package.
package mock_v1

import (
	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	reflect "reflect"
)

// MockFakeResourceReconciler is a mock of FakeResourceReconciler interface
type MockFakeResourceReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockFakeResourceReconcilerMockRecorder
}

// MockFakeResourceReconcilerMockRecorder is the mock recorder for MockFakeResourceReconciler
type MockFakeResourceReconcilerMockRecorder struct {
	mock *MockFakeResourceReconciler
}

// NewMockFakeResourceReconciler creates a new mock instance
func NewMockFakeResourceReconciler(ctrl *gomock.Controller) *MockFakeResourceReconciler {
	mock := &MockFakeResourceReconciler{ctrl: ctrl}
	mock.recorder = &MockFakeResourceReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFakeResourceReconciler) EXPECT() *MockFakeResourceReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method
func (m *MockFakeResourceReconciler) Reconcile(namespace string, desiredResources v1.FakeResourceList, transition v1.TransitionFakeResourceFunc, opts clients.ListOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", namespace, desiredResources, transition, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile
func (mr *MockFakeResourceReconcilerMockRecorder) Reconcile(namespace, desiredResources, transition, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockFakeResourceReconciler)(nil).Reconcile), namespace, desiredResources, transition, opts)
}
