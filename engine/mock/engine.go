// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import mock "github.com/stretchr/testify/mock"

// Engine is an autogenerated mock type for the Engine type
type Engine struct {
	mock.Mock
}

// Submit provides a mock function with given fields: event
func (_m *Engine) Submit(event interface{}) error {
	ret := _m.Called(event)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(event)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}