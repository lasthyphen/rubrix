// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package trickle

// Announce tells our neighbours about a new entity we received.
type Announce struct {
	EngineID uint8
	EventID  []byte
}