package rules

import "github.com/natefinch/council4/mtg/game"

// UnsupportedError reports that the engine hit a mechanic it does not yet
// support at runtime. It is attributable so a simulation can record which game
// failed and why instead of treating it as an opaque engine bug: the runtime
// dispatch path panics with this value, and the simulation failure capture
// recognizes it (see mtg/sim) to flag the failure as unsupported rather than a
// genuine engine defect.
type UnsupportedError struct {
	// Kind is the primitive kind the engine could not resolve.
	Kind game.PrimitiveKind
	// Reason is a human-readable explanation, e.g. "primitive kind 42 has no
	// registered handler".
	Reason string
}

// Error implements the error interface.
func (e UnsupportedError) Error() string {
	return "rules: unsupported mechanic: " + e.Reason
}
