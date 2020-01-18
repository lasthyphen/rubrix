package verification

import (
	"github.com/dapperlabs/flow-go/model/flow"
)

// CompleteExecutionResult represents an execution result that is ready to
// be verified. It contains all execution result and all resources required to
// verify it.
// TODO update this as needed based on execution requirements
type CompleteExecutionResult struct {
	Receipt     *flow.ExecutionReceipt
	Block       *flow.Block
	Collections []*flow.Collection
	ChunkStates []*flow.ChunkState
}