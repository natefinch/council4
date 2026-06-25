package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestAssembleResultGatedBranches proves the centralized branch-assembly helper
// produces the same publisher-then-gated-branches sequence the coin/vote/die
// lowerers previously built by hand: the publisher carries PublishResult set to
// the shared key, and each branch instruction carries the branch predicate with
// its Key filled in from the same shared key, in branch order.
func TestAssembleResultGatedBranches(t *testing.T) {
	t.Parallel()
	const key = game.ResultKey("test-result")
	publisher := game.Instruction{Primitive: game.RollDie{Sides: 2}}
	branches := []resultGatedBranch{
		{
			predicate: game.InstructionResultGate{AmountRange: opt.Val(game.IntRange{Min: 2, Max: 2})},
			sequence:  []game.Instruction{{Primitive: game.GainLife{}}},
		},
		{
			predicate: game.InstructionResultGate{AmountRange: opt.Val(game.IntRange{Min: 1, Max: 1})},
			sequence:  []game.Instruction{{Primitive: game.LoseLife{}}},
		},
	}
	sequence, ok := assembleResultGatedBranches(publisher, key, branches)
	if !ok {
		t.Fatal("assembleResultGatedBranches failed closed unexpectedly")
	}
	if len(sequence) != 3 {
		t.Fatalf("sequence length = %d, want 3", len(sequence))
	}
	if sequence[0].PublishResult != key {
		t.Fatalf("publisher PublishResult = %q, want %q", sequence[0].PublishResult, key)
	}
	if sequence[0].ResultGate.Exists {
		t.Fatal("publisher must not carry a result gate")
	}
	wantRanges := []game.IntRange{{Min: 2, Max: 2}, {Min: 1, Max: 1}}
	for i, want := range wantRanges {
		gate := sequence[i+1].ResultGate
		if !gate.Exists {
			t.Fatalf("branch instruction %d missing result gate", i)
		}
		if gate.Val.Key != key {
			t.Fatalf("branch %d gate key = %q, want %q", i, gate.Val.Key, key)
		}
		if gate.Val.AmountRange.Val != want {
			t.Fatalf("branch %d gate range = %#v, want %#v", i, gate.Val.AmountRange.Val, want)
		}
	}
}

// TestAssembleResultGatedBranchesFailsClosedOnExistingGate proves the helper
// rejects (fails closed) a branch instruction that already carries a result
// gate, the ambiguous nested-gate case the recognized constructs must not
// silently overwrite.
func TestAssembleResultGatedBranchesFailsClosedOnExistingGate(t *testing.T) {
	t.Parallel()
	const key = game.ResultKey("test-result")
	publisher := game.Instruction{Primitive: game.RollDie{Sides: 2}}
	branches := []resultGatedBranch{
		{
			predicate: game.InstructionResultGate{AmountRange: opt.Val(game.IntRange{Min: 2, Max: 2})},
			sequence: []game.Instruction{{
				Primitive:  game.GainLife{},
				ResultGate: opt.Val(game.InstructionResultGate{Key: "preexisting"}),
			}},
		},
	}
	if _, ok := assembleResultGatedBranches(publisher, key, branches); ok {
		t.Fatal("assembleResultGatedBranches accepted a branch with a preexisting gate")
	}
}

// TestAppendResultGatedBranchLeavesSourceUnmodified proves the helper gates a
// copy of each branch instruction, leaving the caller's branch slice untouched
// so a shared branch sequence is not mutated.
func TestAppendResultGatedBranchLeavesSourceUnmodified(t *testing.T) {
	t.Parallel()
	gate := game.InstructionResultGate{Key: "k", Succeeded: game.TriTrue}
	branch := []game.Instruction{{Primitive: game.GainLife{}}}
	out, ok := appendResultGatedBranch(nil, gate, branch)
	if !ok {
		t.Fatal("appendResultGatedBranch failed closed unexpectedly")
	}
	if branch[0].ResultGate.Exists {
		t.Fatal("source branch instruction was mutated")
	}
	if !out[0].ResultGate.Exists || out[0].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("gated instruction = %#v", out[0])
	}
}
