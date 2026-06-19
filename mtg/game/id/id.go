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

// Current returns the value of the most recently issued ID without advancing
// the generator. The zero value reports that no IDs have been issued yet.
func (g *Generator) Current() uint64 {
	return g.next.Load()
}

// Restore sets the generator's counter so the next issued ID is value+1. It
// supports cloning a generator's state into an existing struct: the atomic
// counter is written in place with Store rather than copied by value, which go
// vet's copylocks analysis forbids for a value-returning clone.
func (g *Generator) Restore(value uint64) {
	g.next.Store(value)
}
