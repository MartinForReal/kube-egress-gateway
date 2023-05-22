// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/wgctrlwrapper/wgctrl.go

// Package mockwgctrlwrapper is a generated GoMock package.
package mockwgctrlwrapper

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	wgtypes "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	wgctrlwrapper "github.com/Azure/kube-egress-gateway/pkg/wgctrlwrapper"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockClient)(nil).Close))
}

// ConfigureDevice mocks base method.
func (m *MockClient) ConfigureDevice(name string, cfg wgtypes.Config) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConfigureDevice", name, cfg)
	ret0, _ := ret[0].(error)
	return ret0
}

// ConfigureDevice indicates an expected call of ConfigureDevice.
func (mr *MockClientMockRecorder) ConfigureDevice(name, cfg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConfigureDevice", reflect.TypeOf((*MockClient)(nil).ConfigureDevice), name, cfg)
}

// Device mocks base method.
func (m *MockClient) Device(name string) (*wgtypes.Device, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Device", name)
	ret0, _ := ret[0].(*wgtypes.Device)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Device indicates an expected call of Device.
func (mr *MockClientMockRecorder) Device(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Device", reflect.TypeOf((*MockClient)(nil).Device), name)
}

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

// New mocks base method.
func (m *MockInterface) New() (wgctrlwrapper.Client, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New")
	ret0, _ := ret[0].(wgctrlwrapper.Client)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// New indicates an expected call of New.
func (mr *MockInterfaceMockRecorder) New() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockInterface)(nil).New))
}
