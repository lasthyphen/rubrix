// Package identity provides an interface that bundles all relevant information about staked nodes.
// A simple in-memory implementation is included.
package identity

import (
	"fmt"
	"math/big"
	"sort"
)

type NodeRole string

const (
	CollectorRole NodeRole = "collector"
	ConsensusRole NodeRole = "consensus"
	ExecutorRole  NodeRole = "executor"
	VerifierRole  NodeRole = "verifier"
	ObserverRole  NodeRole = "observer"
)

// NodeIdentity contains the properties of one node. Often, a node is part of a group.
// We assume that there exists a canonical ordering of nodes. Hence, for any fixed group
// of nodes, each is assigned a (zero-based) Index within this specific group.
// In other words, a node's index is with respect to a specific `identity.Table`.
type NodeIdentity interface {
	ID() uint
	Address() string
	Role() NodeRole
	Stake() *big.Int
	Index() uint
}

// Table holds a group of nodes. There are Count nodes in the table. Each node has a
// unique index in the range [0,1, ..., Count-1]. A new Table with a sub-set of nodes
// can be generated by filtering according to various node properties. When generating
// a new table, the nodes' indices in this new table are generated from scratch.
type Table interface {
	Count() uint
	Nodes() []NodeIdentity

	GetByIndex(uint) (NodeIdentity, error)
	GetByID(uint) (NodeIdentity, error)
	GetByAddress(string) (NodeIdentity, error)

	TotalStake() *big.Int

	// FilterByID returns a new Table that contains only nodes that are listed in ids.
	// Note that the new Table is only populated with the SUBSET of nodes that are known.
	// Nodes with unknown IDs are skipped. In this case, an error is returned along side
	// the Table with the known nodes. The error lists all nodes that are unknown.
	// The error is nil, if all nodes foe all desired IDs were found.
	FilterByID(ids []uint) (Table, error)

	// FilterByAddress returns a new Table that contains only nodes that are listed in addresses.
	// Note that the new Table is only populated with the SUBSET of nodes that are known.
	// Nodes with unknown addresses are skipped. In this case, an error is returned along side
	// the Table with the known nodes. The error lists all nodes that are unknown.
	// The error is nil, if all nodes foe all desired addresses were found.
	FilterByAddress(addresses []string) (Table, error)

	// FilterByRole returns a new Table that contains only nodes with the given role.
	// If there are no nodes in the table with the given role, an error is returned
	// and the returned table is empty. (this behaviour is consistewnt with the other FilterBy methods)
	FilterByRole(role NodeRole) (Table, error)

	// FilterByIndex returns a new Table that contains only nodes that are listed in indices.
	// Note that the new Table is only populated with the SUBSET of nodes that are known.
	// Nodes with unknown indices are skipped. In this case, an error is returned along side
	// the Table withg the known nodes. The error lists all nodes that are unknown.
	// The error is nil, if all nodes foe all desired indices were found.
	FilterByIndex(indices []uint) (Table, error)
}

// NodeRecord provides information about one Node that (independent of potential other nodes)
type NodeRecord struct {
	id      uint
	address string
	role    NodeRole
	stake   *big.Int
}

func (i NodeRecord) ID() uint        { return i.id }
func (i NodeRecord) Address() string { return i.address }
func (i NodeRecord) Role() NodeRole  { return i.role }
func (i NodeRecord) Stake() *big.Int { return i.stake }

// NodeRecords is a slice of *NodeRecord which implements sort.Interface
// Sorting is based solely on NodeRecord.ID
type NodeRecords []*NodeRecord

func (ns NodeRecords) Len() int           { return len(ns) }
func (ns NodeRecords) Less(i, j int) bool { return ns[i].id < ns[j].id }
func (ns NodeRecords) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }

// Implementation of NodeIdentity interface
type nodeIdentity struct {
	*NodeRecord
	index uint
}

func (i nodeIdentity) Index() uint { return i.index }

// nodeIdentities is a slice of *nodeIdentity which implements sort.Interface
// Sorting is based solely on nodeIdentity.ID
type nodeIdentities []*nodeIdentity

func (ns nodeIdentities) Len() int           { return len(ns) }
func (ns nodeIdentities) Less(i, j int) bool { return ns[i].id < ns[j].id }
func (ns nodeIdentities) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }

// InMemoryIdentityTable is an in-memory implementation of the interface identity.Table
type InMemoryIdentityTable struct {
	nodes      []*nodeIdentity
	addressMap map[string]*nodeIdentity
	idMap      map[uint]*nodeIdentity
}

func (t InMemoryIdentityTable) Count() uint {
	return uint(len(t.nodes))
}

func (t InMemoryIdentityTable) Nodes() []NodeIdentity {
	identities := make([]NodeIdentity, len(t.nodes))
	// converting explicitly from nodeIdentity to interface NodeIdentity
	for i, n := range t.nodes {
		identities[i] = n
	}
	return identities
}

func (t InMemoryIdentityTable) GetByIndex(idx uint) (NodeIdentity, error) {
	if int(idx) > len(t.nodes) {
		return nil, &NodeNotFoundError{fmt.Sprint(idx)}
	}
	return t.nodes[idx], nil
}

func (t InMemoryIdentityTable) GetByID(id uint) (NodeIdentity, error) {
	value, found := t.idMap[id]
	if !found {
		return nil, &NodeNotFoundError{fmt.Sprint(id)}
	}
	return value, nil
}

func (t InMemoryIdentityTable) GetByAddress(address string) (NodeIdentity, error) {
	value, found := t.addressMap[address]
	if !found {
		return nil, &NodeNotFoundError{address}
	}
	return value, nil
}

func (t InMemoryIdentityTable) TotalStake() *big.Int {
	s := big.NewInt(0)
	for _, n := range t.nodes {
		s.Add(s, n.Stake())
	}
	return s
}

func (t InMemoryIdentityTable) FilterByID(ids []uint) (Table, error) {
	nodes := make([]*NodeRecord, len(ids))
	var missing []uint
	var n *nodeIdentity
	var found bool
	var idx int = 0
	for _, id := range ids {
		n, found = t.idMap[id]
		if found {
			nodes[idx] = n.NodeRecord
			idx++
		} else {
			missing = append(missing, id)
		}
	}
	tnew := NewInMemoryIdentityTable(nodes[0:idx])
	if len(missing) > 0 {
		return tnew, &NodeNotFoundError{fmt.Sprint(missing)}
	}
	return tnew, nil
}

func (t InMemoryIdentityTable) FilterByAddress(addresses []string) (Table, error) {
	nodes := make([]*NodeRecord, len(addresses))
	var missing []string
	var n *nodeIdentity
	var found bool
	var idx int = 0
	for _, addr := range addresses {
		n, found = t.addressMap[addr]
		if found {
			nodes[idx] = n.NodeRecord
			idx++
		} else {
			missing = append(missing, addr)
		}
	}
	tnew := NewInMemoryIdentityTable(nodes[0:idx])
	if len(missing) > 0 {
		return tnew, &NodeNotFoundError{fmt.Sprint(missing)}
	}
	return tnew, nil
}

func (t InMemoryIdentityTable) FilterByRole(role NodeRole) (Table, error) {
	nodes := make([]*NodeRecord, t.Count())
	var idx int = 0
	for _, n := range t.nodes {
		if n.Role() == role {
			nodes[idx] = n.NodeRecord
			idx++
		}
	}
	tnew := NewInMemoryIdentityTable(nodes[0:idx])
	if idx == 0 {
		return tnew, &NodeNotFoundError{fmt.Sprint(role)}
	}
	return tnew, nil
}

func (t InMemoryIdentityTable) FilterByIndex(indices []uint) (Table, error) {
	nodes := make([]*NodeRecord, len(indices))
	var missing []uint
	var idx int = 0
	nodeCount := uint(t.Count())
	allNodes := t.nodes
	for _, i := range indices {
		if i < nodeCount {
			nodes[idx] = allNodes[i].NodeRecord
			idx++
		} else {
			missing = append(missing, i)
		}
	}
	tnew := NewInMemoryIdentityTable(nodes[0:idx])
	if len(missing) > 0 {
		return tnew, &NodeNotFoundError{fmt.Sprint(missing)}
	}
	return tnew, nil
}

type NodeNotFoundError struct {
	key string
}

func (e *NodeNotFoundError) Error() string {
	return fmt.Sprintf("node with '%s' not found", e.key)
}

// NewInMemoryIdentityTable generates an `identity.Table` which is maintained in-memory
func NewInMemoryIdentityTable(nodes []*NodeRecord) *InMemoryIdentityTable {
	nidentities := newSortedNodeIdentities(nodes)

	addressMap := make(map[string]*nodeIdentity)
	idMap := make(map[uint]*nodeIdentity)
	var last *NodeRecord = nil // reference to previous nodeIdentity to detect duplicates
	for i, n := range nidentities {
		if last == n.NodeRecord {
			panic("Duplicate NodeRecord not supported")
		}
		last = n.NodeRecord
		n.index = uint(i)
		addressMap[n.Address()] = n
		idMap[n.ID()] = n
	}

	return &InMemoryIdentityTable{nidentities, addressMap, idMap}
}

// newSortedNodeIdentities wraps each NodeRecord into a nodeIdentity type
// with default `index=0` and sorts the elements. Checks for nil elements.
func newSortedNodeIdentities(nodes []*NodeRecord) []*nodeIdentity {
	// While the slice `nodes` is copied, the data in the slice is not sufficient to sort the slice without mutating the underlying array
	// For more details, see https://blog.golang.org/go-slices-usage-and-internals
	var nidentities nodeIdentities = make([]*nodeIdentity, len(nodes))
	for i, n := range nodes {
		if n == nil {
			panic("NodeRecord cannot be nil")
		}
		nidentities[i] = &nodeIdentity{NodeRecord: n}
	}
	sort.Sort(nidentities)
	return nidentities
}