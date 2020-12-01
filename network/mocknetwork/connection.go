// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocknetwork

import mock "github.com/stretchr/testify/mock"

// Connection is an autogenerated mock type for the Connection type
type Connection struct {
	mock.Mock
}

// Receive provides a mock function with given fields:
func (_m *Connection) Receive() (interface{}, error) {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Send provides a mock function with given fields: msg
func (_m *Connection) Send(msg interface{}) error {
	ret := _m.Called(msg)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}