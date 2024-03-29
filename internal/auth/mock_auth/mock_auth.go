// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aria3ppp/watchlist-server/internal/auth (interfaces: Interface)

// Package mock_auth is a generated GoMock package.
package mock_auth

import (
	reflect "reflect"
	time "time"

	auth "github.com/aria3ppp/watchlist-server/internal/auth"
	gomock "github.com/golang/mock/gomock"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// GenerateJwtToken mocks base method.
func (m *MockInterface) GenerateJwtToken(arg0 *auth.Payload) (string, time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateJwtToken", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(time.Time)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GenerateJwtToken indicates an expected call of GenerateJwtToken.
func (mr *MockInterfaceMockRecorder) GenerateJwtToken(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateJwtToken", reflect.TypeOf((*MockInterface)(nil).GenerateJwtToken), arg0)
}

// GenerateRefreshToken mocks base method.
func (m *MockInterface) GenerateRefreshToken() (string, time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateRefreshToken")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(time.Time)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GenerateRefreshToken indicates an expected call of GenerateRefreshToken.
func (mr *MockInterfaceMockRecorder) GenerateRefreshToken() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateRefreshToken", reflect.TypeOf((*MockInterface)(nil).GenerateRefreshToken))
}
