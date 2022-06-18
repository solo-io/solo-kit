// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/multicluster/v1/kube_config_reconciler.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
)

// MockKubeConfigReconciler is a mock of KubeConfigReconciler interface
type MockKubeConfigReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockKubeConfigReconcilerMockRecorder
}

// MockKubeConfigReconcilerMockRecorder is the mock recorder for MockKubeConfigReconciler
type MockKubeConfigReconcilerMockRecorder struct {
	mock *MockKubeConfigReconciler
}

// NewMockKubeConfigReconciler creates a new mock instance
func NewMockKubeConfigReconciler(ctrl *gomock.Controller) *MockKubeConfigReconciler {
	mock := &MockKubeConfigReconciler{ctrl: ctrl}
	mock.recorder = &MockKubeConfigReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockKubeConfigReconciler) EXPECT() *MockKubeConfigReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method
func (m *MockKubeConfigReconciler) Reconcile(namespace string, desiredResources v1.KubeConfigList, transition v1.TransitionKubeConfigFunc, opts clients.ListOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", namespace, desiredResources, transition, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile
func (mr *MockKubeConfigReconcilerMockRecorder) Reconcile(namespace, desiredResources, transition, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockKubeConfigReconciler)(nil).Reconcile), namespace, desiredResources, transition, opts)
}
