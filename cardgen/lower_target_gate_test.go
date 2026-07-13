package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// branchGatedInstruction builds an instruction carrying a single cast-branch condition.
func branchGatedInstruction(prim game.Primitive, cond game.Condition) game.Instruction {
	return game.Instruction{
		Primitive: prim,
		Condition: opt.Val(game.EffectCondition{Condition: opt.Val(cond)}),
	}
}

// TestAssignTargetGatesGatesKickerOnlyTarget verifies the compiler assigns the
// kicked gate to a target referenced only by a kicked instruction while leaving
// an always-referenced target ungated (the Jilt shape).
func TestAssignTargetGatesGatesKickerOnlyTarget(t *testing.T) {
	targets := []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
	}
	sequence := []game.Instruction{
		{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}},
		branchGatedInstruction(
			game.Damage{Recipient: game.AnyTargetDamageRecipient(1), Amount: game.Fixed(2)},
			game.Condition{SpellWasKicked: true},
		),
	}

	out, ok := assignTargetGates(targets, sequence)
	if !ok {
		t.Fatal("assignTargetGates failed for a well-formed kicked target")
	}
	if out[0].Gate != game.TargetGateAlways {
		t.Errorf("base spec gate = %v, want TargetGateAlways", out[0].Gate)
	}
	if out[1].Gate != game.TargetGateSpellKicked {
		t.Errorf("kicker spec gate = %v, want TargetGateSpellKicked", out[1].Gate)
	}
}

// TestAssignTargetGatesGatesPromisedOnlyTarget verifies the promised-only target
// (the Gift additional-target shape) is gated on the gift being promised.
func TestAssignTargetGatesGatesPromisedOnlyTarget(t *testing.T) {
	targets := []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "artifact"},
	}
	sequence := []game.Instruction{
		{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}},
		branchGatedInstruction(
			game.Destroy{Object: game.TargetPermanentReference(1)},
			game.Condition{GiftPromised: true},
		),
	}

	out, ok := assignTargetGates(targets, sequence)
	if !ok {
		t.Fatal("assignTargetGates failed for a well-formed promised target")
	}
	if out[1].Gate != game.TargetGateGiftPromised {
		t.Errorf("promised spec gate = %v, want TargetGateGiftPromised", out[1].Gate)
	}
}

// TestAssignTargetGatesComplementaryGatesResolveToAlways verifies that a target
// referenced by both branches of one mechanic (the kicked clause and its
// not-kicked "instead" counterpart — Hypnotic Cloud's discard) is required
// either way and resolves to TargetGateAlways rather than a conflict.
func TestAssignTargetGatesComplementaryGatesResolveToAlways(t *testing.T) {
	targets := []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}
	sequence := []game.Instruction{
		branchGatedInstruction(
			game.Bounce{Object: game.TargetPermanentReference(0)},
			game.Condition{SpellWasKicked: true, Negate: true},
		),
		branchGatedInstruction(
			game.Destroy{Object: game.TargetPermanentReference(0)},
			game.Condition{SpellWasKicked: true},
		),
	}

	out, ok := assignTargetGates(targets, sequence)
	if !ok {
		t.Fatal("assignTargetGates failed for complementary kicked/not-kicked references")
	}
	if out[0].Gate != game.TargetGateAlways {
		t.Errorf("complementary spec gate = %v, want TargetGateAlways (required on both branches)", out[0].Gate)
	}
}

// TestAssignTargetGatesFailsClosedOnCrossMechanicConflict covers the fail-closed
// requirement: a single target referenced under two different mechanics' gates
// (kicked and gift-promised) cannot be expressed by one gate field, so the
// assigner reports failure and the card is left unsupported rather than
// mis-announced.
func TestAssignTargetGatesFailsClosedOnCrossMechanicConflict(t *testing.T) {
	targets := []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}
	sequence := []game.Instruction{
		branchGatedInstruction(
			game.Bounce{Object: game.TargetPermanentReference(0)},
			game.Condition{SpellWasKicked: true},
		),
		branchGatedInstruction(
			game.Destroy{Object: game.TargetPermanentReference(0)},
			game.Condition{GiftPromised: true},
		),
	}

	if _, ok := assignTargetGates(targets, sequence); ok {
		t.Fatal("assignTargetGates() ok = true for a cross-mechanic conflict, want fail closed")
	}
}

// TestAssignTargetGatesUngatedSequenceUnchanged verifies the byte-identity fast
// path: a spell with no cast-branch-gated instruction returns its target specs
// untouched (same slice), so every non-gift, non-kicker card is unaffected.
func TestAssignTargetGatesUngatedSequenceUnchanged(t *testing.T) {
	targets := []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}
	sequence := []game.Instruction{{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}}}

	out, ok := assignTargetGates(targets, sequence)
	if !ok {
		t.Fatal("assignTargetGates failed for an ungated sequence")
	}
	if &out[0] != &targets[0] {
		t.Error("ungated sequence did not return the original target slice (byte-identity fast path broken)")
	}
}

// TestCastBranchGateClassification checks the per-instruction gate classifier for
// each cast-branch condition and its fail-closed case.
func TestCastBranchGateClassification(t *testing.T) {
	cases := []struct {
		name      string
		cond      opt.V[game.EffectCondition]
		wantGate  game.TargetGate
		wantGated bool
		wantOK    bool
	}{
		{"ungated", opt.V[game.EffectCondition]{}, game.TargetGateAlways, false, true},
		{"kicked", opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{SpellWasKicked: true})}), game.TargetGateSpellKicked, true, true},
		{"not kicked", opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{SpellWasKicked: true, Negate: true})}), game.TargetGateSpellNotKicked, true, true},
		{"promised", opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{GiftPromised: true})}), game.TargetGateGiftPromised, true, true},
		{"not promised", opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{GiftPromised: true, Negate: true})}), game.TargetGateGiftNotPromised, true, true},
		{"bargained", opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{SpellWasBargained: true})}), game.TargetGateSpellBargained, true, true},
		{"not bargained", opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{SpellWasBargained: true, Negate: true})}), game.TargetGateSpellNotBargained, true, true},
		{"both mechanics fail closed", opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{SpellWasKicked: true, GiftPromised: true})}), game.TargetGateAlways, false, false},
		{"kicker and bargain fail closed", opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{SpellWasKicked: true, SpellWasBargained: true})}), game.TargetGateAlways, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inst := &game.Instruction{Condition: tc.cond}
			gate, gated, ok := castBranchGate(inst)
			if gate != tc.wantGate || gated != tc.wantGated || ok != tc.wantOK {
				t.Errorf("castBranchGate = (%v, %v, %v), want (%v, %v, %v)", gate, gated, ok, tc.wantGate, tc.wantGated, tc.wantOK)
			}
		})
	}
}

// TestAssignTargetGatesFailsClosedOnUnwalkableGatedTarget covers the fail-closed
// audit: when a cast-branch-gated instruction references a target but its
// primitive cannot be fully expressed by the target-index walker, assignTargetGates
// must fail closed rather than silently leave the branch-only target
// unconditionally required. Here the kicked instruction is a Fight that references
// target 1 (recorded) but carries an unexpressible related object, so the walk
// fails after reaching the target.
func TestAssignTargetGatesFailsClosedOnUnwalkableGatedTarget(t *testing.T) {
	targets := []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
	}
	sequence := []game.Instruction{
		{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}},
		branchGatedInstruction(
			game.Fight{Object: game.TargetPermanentReference(1), RelatedObject: game.ObjectReference{}},
			game.Condition{SpellWasKicked: true},
		),
	}

	if _, ok := assignTargetGates(targets, sequence); ok {
		t.Fatal("assignTargetGates returned ok for a gated target-bearing primitive it cannot walk; want fail closed")
	}
}

// TestAssignTargetGatesSkipsUnwalkableTargetlessGatedInstruction proves the
// complement: a gated instruction whose unwalkable primitive references no
// announced target (a "scry" gated on "if this spell was kicked") does not fail
// the card closed; the always-required base target keeps its ungated form, so a
// card like Runic Shot stays supported and byte-identical.
func TestAssignTargetGatesSkipsUnwalkableTargetlessGatedInstruction(t *testing.T) {
	targets := []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
	}
	sequence := []game.Instruction{
		{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}},
		branchGatedInstruction(
			game.Scry{Amount: game.Fixed(2)},
			game.Condition{SpellWasKicked: true},
		),
	}

	out, ok := assignTargetGates(targets, sequence)
	if !ok {
		t.Fatal("assignTargetGates failed closed for a targetless gated scry; want supported")
	}
	if out[0].Gate != game.TargetGateAlways {
		t.Errorf("base spec gate = %v, want TargetGateAlways (unaffected by targetless gated scry)", out[0].Gate)
	}
}
