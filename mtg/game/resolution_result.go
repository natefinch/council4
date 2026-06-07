package game

// InstructionResolutionResult records whether a named instruction was accepted
// and actually did anything rules-relevant. "Succeeded" is distinct from "Accepted"
// because impossible instructions do only as much as possible (CR 101.3).
type InstructionResolutionResult struct {
	Accepted  bool
	Succeeded bool
	Amount    int
}
