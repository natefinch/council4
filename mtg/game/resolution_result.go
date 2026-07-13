package game

// InstructionResolutionResult records whether a named instruction was accepted
// and actually did anything rules-relevant. "Succeeded" is distinct from "Accepted"
// because impossible instructions do only as much as possible (CR 101.3).
type InstructionResolutionResult struct {
	Accepted  bool
	Succeeded bool
	Amount    int
	// AcceptedActors is the set of players who accepted a group offer or Tempting
	// offer this instruction published (empty for a single-decider instruction).
	// It is the generic accepted-member publication: Count gives the number who
	// accepted (which a Tempting offer's controller repeat matches) and Members
	// gives their identities for a future per-accepter consequence. It is a
	// comparable bitmask so InstructionResolutionResult stays comparable for the
	// cloneable resolution-results map.
	AcceptedActors PlayerSet
}
