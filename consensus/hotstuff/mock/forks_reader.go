// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import flow "github.com/dapperlabs/flow-go/model/flow"

import mock "github.com/stretchr/testify/mock"
import model "github.com/dapperlabs/flow-go/consensus/hotstuff/model"

// ForksReader is an autogenerated mock type for the ForksReader type
type ForksReader struct {
	mock.Mock
}

// FinalizedBlock provides a mock function with given fields:
func (_m *ForksReader) FinalizedBlock() *model.Block {
	ret := _m.Called()

	var r0 *model.Block
	if rf, ok := ret.Get(0).(func() *model.Block); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Block)
		}
	}

	return r0
}

// FinalizedView provides a mock function with given fields:
func (_m *ForksReader) FinalizedView() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetBlock provides a mock function with given fields: id
func (_m *ForksReader) GetBlock(id flow.Identifier) (*model.Block, bool) {
	ret := _m.Called(id)

	var r0 *model.Block
	if rf, ok := ret.Get(0).(func(flow.Identifier) *model.Block); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Block)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(flow.Identifier) bool); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// GetBlocksForView provides a mock function with given fields: view
func (_m *ForksReader) GetBlocksForView(view uint64) []*model.Block {
	ret := _m.Called(view)

	var r0 []*model.Block
	if rf, ok := ret.Get(0).(func(uint64) []*model.Block); ok {
		r0 = rf(view)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Block)
		}
	}

	return r0
}

// IsSafeBlock provides a mock function with given fields: block
func (_m *ForksReader) IsSafeBlock(block *model.Block) bool {
	ret := _m.Called(block)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*model.Block) bool); ok {
		r0 = rf(block)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}