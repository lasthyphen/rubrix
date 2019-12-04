// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package operation

import (
	"github.com/dgraph-io/badger/v2"

	"github.com/dapperlabs/flow-go/model/flow"
)

func InsertRole(nodeID flow.Identifier, role flow.Role) func(*badger.Txn) error {
	return insert(makePrefix(codeRole, nodeID), role)
}

func RetrieveRole(nodeID flow.Identifier, role *flow.Role) func(*badger.Txn) error {
	return retrieve(makePrefix(codeRole, nodeID), role)
}