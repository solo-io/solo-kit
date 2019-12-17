// Code generated by MockGen. DO NOT EDIT.
// Source: multi_watcher.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	clients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	resources "github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// MockWatchAggregator is a mock of WatchAggregator interface
type MockWatchAggregator struct {
	ctrl     *gomock.Controller
	recorder *MockWatchAggregatorMockRecorder
}

// MockWatchAggregatorMockRecorder is the mock recorder for MockWatchAggregator
type MockWatchAggregatorMockRecorder struct {
	mock *MockWatchAggregator
}

// NewMockWatchAggregator creates a new mock instance
func NewMockWatchAggregator(ctrl *gomock.Controller) *MockWatchAggregator {
	mock := &MockWatchAggregator{ctrl: ctrl}
	mock.recorder = &MockWatchAggregatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWatchAggregator) EXPECT() *MockWatchAggregatorMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockWatchAggregator) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", namespace, opts)
	ret0, _ := ret[0].(<-chan resources.ResourceList)
	ret1, _ := ret[1].(<-chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Watch indicates an expected call of Watch
func (mr *MockWatchAggregatorMockRecorder) Watch(namespace, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockWatchAggregator)(nil).Watch), namespace, opts)
}

// AddWatch mocks base method
func (m *MockWatchAggregator) AddWatch(w clients.ResourceWatcher) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddWatch", w)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddWatch indicates an expected call of AddWatch
func (mr *MockWatchAggregatorMockRecorder) AddWatch(w interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddWatch", reflect.TypeOf((*MockWatchAggregator)(nil).AddWatch), w)
}

// RemoveWatch mocks base method
func (m *MockWatchAggregator) RemoveWatch(w clients.ResourceWatcher) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RemoveWatch", w)
}

// RemoveWatch indicates an expected call of RemoveWatch
func (mr *MockWatchAggregatorMockRecorder) RemoveWatch(w interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveWatch", reflect.TypeOf((*MockWatchAggregator)(nil).RemoveWatch), w)
}
