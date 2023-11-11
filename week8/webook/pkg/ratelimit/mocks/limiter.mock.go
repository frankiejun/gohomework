// Code generated by MockGen. DO NOT EDIT.
// Source: ./types.go

// Package limitmocks is a generated GoMock package.
package limitmocks

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	context "golang.org/x/net/context"
)

// MockLimiter is a mock of Limiter interface.
type MockLimiter struct {
	ctrl     *gomock.Controller
	recorder *MockLimiterMockRecorder
}

// MockLimiterMockRecorder is the mock recorder for MockLimiter.
type MockLimiterMockRecorder struct {
	mock *MockLimiter
}

// NewMockLimiter creates a new mock instance.
func NewMockLimiter(ctrl *gomock.Controller) *MockLimiter {
	mock := &MockLimiter{ctrl: ctrl}
	mock.recorder = &MockLimiterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLimiter) EXPECT() *MockLimiterMockRecorder {
	return m.recorder
}

// Limit mocks base method.
func (m *MockLimiter) Limit(ctx context.Context, key string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Limit", ctx, key)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Limit indicates an expected call of Limit.
func (mr *MockLimiterMockRecorder) Limit(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Limit", reflect.TypeOf((*MockLimiter)(nil).Limit), ctx, key)
}
