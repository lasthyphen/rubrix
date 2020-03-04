package assignment

import (
	"bytes"

	"github.com/rs/zerolog/log"

	"github.com/dapperlabs/flow-go/model/flow"
)

// IdentifierList defines a sortable list of identifiers
type IdentifierList []flow.Identifier

// Len returns length of the IdentiferList in the number of stored identifiers.
// It satisfies the sort.Interface making the IdentifierList sortable.
func (il IdentifierList) Len() int {
	return len(il)
}

// Less returns true if element i in the IdentifierList is less than j based on its identifier.
// Otherwise it returns true.
// It satisfies the sort.Interface making the IdentifierList sortable.
func (il IdentifierList) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(il[i][:], il[j][:]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Error().Msg("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

// Swap swaps the element i and j in the IdentifierList.
// It satisfies the sort.Interface making the IdentifierList sortable.
func (il IdentifierList) Swap(i, j int) {
	il[j], il[i] = il[i], il[j]
}

// JoinIdentifierLists appends and returns two IdentifierLists
func JoinIdentifierLists(this, other IdentifierList) IdentifierList {
	joined := make([]flow.Identifier, 0)
	for _, id := range this {
		joined = append(joined, id)
	}

	for _, id := range other {
		joined = append(joined, id)
	}

	return joined
}