// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/api/v1/resources/common/kubernetes/pod_reconciler.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	reflect "reflect"
)

// MockPodReconciler is a mock of PodReconciler interface
type MockPodReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockPodReconcilerMockRecorder
}

// MockPodReconcilerMockRecorder is the mock recorder for MockPodReconciler
type MockPodReconcilerMockRecorder struct {
	mock *MockPodReconciler
}

// NewMockPodReconciler creates a new mock instance
func NewMockPodReconciler(ctrl *gomock.Controller) *MockPodReconciler {
	mock := &MockPodReconciler{ctrl: ctrl}
	mock.recorder = &MockPodReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPodReconciler) EXPECT() *MockPodReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method
func (m *MockPodReconciler) Reconcile(namespace string, desiredResources kubernetes.PodList, transition kubernetes.TransitionPodFunc, opts clients.ListOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", namespace, desiredResources, transition, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile
func (mr *MockPodReconcilerMockRecorder) Reconcile(namespace, desiredResources, transition, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockPodReconciler)(nil).Reconcile), namespace, desiredResources, transition, opts)
}
