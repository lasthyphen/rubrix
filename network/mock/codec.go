// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import io "io"
import mock "github.com/stretchr/testify/mock"
import network "github.com/dapperlabs/flow-go/network"

// Codec is an autogenerated mock type for the Codec type
type Codec struct {
	mock.Mock
}

// Decode provides a mock function with given fields: data
func (_m *Codec) Decode(data []byte) (interface{}, error) {
	ret := _m.Called(data)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func([]byte) interface{}); ok {
		r0 = rf(data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Encode provides a mock function with given fields: v
func (_m *Codec) Encode(v interface{}) ([]byte, error) {
	ret := _m.Called(v)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(interface{}) []byte); ok {
		r0 = rf(v)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}) error); ok {
		r1 = rf(v)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewDecoder provides a mock function with given fields: r
func (_m *Codec) NewDecoder(r io.Reader) network.Decoder {
	ret := _m.Called(r)

	var r0 network.Decoder
	if rf, ok := ret.Get(0).(func(io.Reader) network.Decoder); ok {
		r0 = rf(r)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(network.Decoder)
		}
	}

	return r0
}

// NewEncoder provides a mock function with given fields: w
func (_m *Codec) NewEncoder(w io.Writer) network.Encoder {
	ret := _m.Called(w)

	var r0 network.Encoder
	if rf, ok := ret.Get(0).(func(io.Writer) network.Encoder); ok {
		r0 = rf(w)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(network.Encoder)
		}
	}

	return r0
}