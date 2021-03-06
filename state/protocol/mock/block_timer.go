// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// BlockTimer is an autogenerated mock type for the BlockTimer type
type BlockTimer struct {
	mock.Mock
}

// Build provides a mock function with given fields: parentTimestamp
func (_m *BlockTimer) Build(parentTimestamp time.Time) time.Time {
	ret := _m.Called(parentTimestamp)

	var r0 time.Time
	if rf, ok := ret.Get(0).(func(time.Time) time.Time); ok {
		r0 = rf(parentTimestamp)
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// Validate provides a mock function with given fields: parentTimestamp, currentTimestamp
func (_m *BlockTimer) Validate(parentTimestamp time.Time, currentTimestamp time.Time) error {
	ret := _m.Called(parentTimestamp, currentTimestamp)

	var r0 error
	if rf, ok := ret.Get(0).(func(time.Time, time.Time) error); ok {
		r0 = rf(parentTimestamp, currentTimestamp)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
