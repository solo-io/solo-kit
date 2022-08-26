// Code generated by MockGen. DO NOT EDIT.
// Source: test/mocks/v2alpha1/frequently_changing_annotations_resource_reconciler.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v2alpha1 "github.com/solo-io/solo-kit/test/mocks/v2alpha1"
)

// MockFrequentlyChangingAnnotationsResourceReconciler is a mock of FrequentlyChangingAnnotationsResourceReconciler interface
type MockFrequentlyChangingAnnotationsResourceReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockFrequentlyChangingAnnotationsResourceReconcilerMockRecorder
}

// MockFrequentlyChangingAnnotationsResourceReconcilerMockRecorder is the mock recorder for MockFrequentlyChangingAnnotationsResourceReconciler
type MockFrequentlyChangingAnnotationsResourceReconcilerMockRecorder struct {
	mock *MockFrequentlyChangingAnnotationsResourceReconciler
}

// NewMockFrequentlyChangingAnnotationsResourceReconciler creates a new mock instance
func NewMockFrequentlyChangingAnnotationsResourceReconciler(ctrl *gomock.Controller) *MockFrequentlyChangingAnnotationsResourceReconciler {
	mock := &MockFrequentlyChangingAnnotationsResourceReconciler{ctrl: ctrl}
	mock.recorder = &MockFrequentlyChangingAnnotationsResourceReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFrequentlyChangingAnnotationsResourceReconciler) EXPECT() *MockFrequentlyChangingAnnotationsResourceReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method
func (m *MockFrequentlyChangingAnnotationsResourceReconciler) Reconcile(namespace string, desiredResources v2alpha1.FrequentlyChangingAnnotationsResourceList, transition v2alpha1.TransitionFrequentlyChangingAnnotationsResourceFunc, opts clients.ListOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", namespace, desiredResources, transition, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile
func (mr *MockFrequentlyChangingAnnotationsResourceReconcilerMockRecorder) Reconcile(namespace, desiredResources, transition, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockFrequentlyChangingAnnotationsResourceReconciler)(nil).Reconcile), namespace, desiredResources, transition, opts)
}