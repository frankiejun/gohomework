// Code generated by MockGen. DO NOT EDIT.
// Source: internal/repository/smssave.go

// Package repomocks is a generated GoMock package.
package repomocks

import (
	context "context"
	reflect "reflect"

	domain "gitee.com/geekbang/basic-go/webook/internal/domain"
	gomock "go.uber.org/mock/gomock"
)

// MockSmsRepository is a mock of SmsRepository interface.
type MockSmsRepository struct {
	ctrl     *gomock.Controller
	recorder *MockSmsRepositoryMockRecorder
}

// MockSmsRepositoryMockRecorder is the mock recorder for MockSmsRepository.
type MockSmsRepositoryMockRecorder struct {
	mock *MockSmsRepository
}

// NewMockSmsRepository creates a new mock instance.
func NewMockSmsRepository(ctrl *gomock.Controller) *MockSmsRepository {
	mock := &MockSmsRepository{ctrl: ctrl}
	mock.recorder = &MockSmsRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSmsRepository) EXPECT() *MockSmsRepositoryMockRecorder {
	return m.recorder
}

// Fetch mocks base method.
func (m *MockSmsRepository) Fetch(ctx context.Context) (domain.Sms, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Fetch", ctx)
	ret0, _ := ret[0].(domain.Sms)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Fetch indicates an expected call of Fetch.
func (mr *MockSmsRepositoryMockRecorder) Fetch(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fetch", reflect.TypeOf((*MockSmsRepository)(nil).Fetch), ctx)
}

// Save mocks base method.
func (m *MockSmsRepository) Save(ctx context.Context, sms domain.Sms) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, sms)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockSmsRepositoryMockRecorder) Save(ctx, sms interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockSmsRepository)(nil).Save), ctx, sms)
}
