package cardgen

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// resultGatedBranch is one branch of a publish-result-then-gate-branches
// assembly: the predicate that gates the branch on the publisher's published
// result and the branch's already-lowered instruction sequence. The predicate's
// Key is filled in by assembleResultGatedBranches from the assembly's shared
// result key, so callers populate only the Accepted/Succeeded/AmountRange
// predicate that selects the branch.
type resultGatedBranch struct {
	predicate game.InstructionResultGate
	sequence  []game.Instruction
}

// appendResultGatedBranch appends each instruction of branch to dst with gate
// attached as its ResultGate, returning the grown sequence. It fails closed
// (returns false) if any branch instruction already carries a result gate — the
// ambiguous nested-gate case a recognized result-gated construct must reject
// rather than silently overwrite. Each instruction is copied before its gate is
// set, so branch is left unmodified.
func appendResultGatedBranch(dst []game.Instruction, gate game.InstructionResultGate, branch []game.Instruction) ([]game.Instruction, bool) {
	for k := range branch {
		instruction := branch[k]
		if instruction.ResultGate.Exists {
			return nil, false
		}
		instruction.ResultGate = opt.Val(gate)
		dst = append(dst, instruction)
	}
	return dst, true
}

// assembleResultGatedBranches builds the instruction sequence for a
// publish-result-then-gate-branches construct (coin flip, vote, die-roll outcome
// table). The sequence begins with publisher — the instruction that publishes the
// branch-selecting result under key — followed by every branch's instructions
// with the branch predicate attached as a result gate. It centralizes result-key
// wiring (set once on the publisher's PublishResult and on every branch gate's
// Key) and duplicate-gate rejection, failing closed if any branch instruction
// already carries a result gate.
func assembleResultGatedBranches(publisher game.Instruction, key game.ResultKey, branches []resultGatedBranch) ([]game.Instruction, bool) {
	publisher.PublishResult = key
	sequence := make([]game.Instruction, 0, len(branches)+1)
	sequence = append(sequence, publisher)
	for b := range branches {
		gate := branches[b].predicate
		gate.Key = key
		var ok bool
		sequence, ok = appendResultGatedBranch(sequence, gate, branches[b].sequence)
		if !ok {
			return nil, false
		}
	}
	return sequence, true
}
