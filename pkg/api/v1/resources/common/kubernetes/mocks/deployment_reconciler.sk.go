// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/api/v1/resources/common/kubernetes/deployment_reconciler.sk.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

// MockDeploymentReconciler is a mock of DeploymentReconciler interface
type MockDeploymentReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockDeploymentReconcilerMockRecorder
}

// MockDeploymentReconcilerMockRecorder is the mock recorder for MockDeploymentReconciler
type MockDeploymentReconcilerMockRecorder struct {
	mock *MockDeploymentReconciler
}

// NewMockDeploymentReconciler creates a new mock instance
func NewMockDeploymentReconciler(ctrl *gomock.Controller) *MockDeploymentReconciler {
	mock := &MockDeploymentReconciler{ctrl: ctrl}
	mock.recorder = &MockDeploymentReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeploymentReconciler) EXPECT() *MockDeploymentReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method
func (m *MockDeploymentReconciler) Reconcile(namespace string, desiredResources kubernetes.DeploymentList, transition kubernetes.TransitionDeploymentFunc, opts clients.ListOpts) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", namespace, desiredResources, transition, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile
func (mr *MockDeploymentReconcilerMockRecorder) Reconcile(namespace, desiredResources, transition, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockDeploymentReconciler)(nil).Reconcile), namespace, desiredResources, transition, opts)
}
