// Package id provides unique identifiers for game objects.
//
// Every card instance, permanent, stack object, and token in a game
// receives a unique ID from a Generator. IDs are used for cross-referencing
// objects across zones and data structures without creating import cycles.
package id

import "sync/atomic"

// ID is a unique identifier for a game object.
type ID uint64

// Generator produces unique sequential IDs.
// It is safe for concurrent use.
type Generator struct {
	next atomic.Uint64
}

// Next returns the next unique ID.
func (g *Generator) Next() ID {
	return ID(g.next.Add(1))
}
