package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTragicSlipMorbidInsteadModifyPT verifies that Tragic Slip's base
// "-1/-1" paragraph and its "Morbid — ... -13/-13 ... instead if a creature
// died this turn." paragraph fuse into a single spell over one target creature
// whose -1/-1 modification resolves only when no creature died this turn and
// whose -13/-13 modification resolves only when one did.
func TestLowerTragicSlipMorbidInsteadModifyPT(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Tragic Slip",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "Target creature gets -1/-1 until end of turn.\n" +
			"Morbid — That creature gets -13/-13 until end of turn instead if a creature died this turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Tragic Slip produced no spell ability")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %#v, want 1", modes)
	}
	mode := modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowPermanent {
		t.Fatalf("targets = %#v, want one permanent target", mode.Targets)
	}
	seq := mode.Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2 (base + morbid)", len(seq))
	}
	for i, instr := range seq {
		modify, ok := instr.Primitive.(game.ModifyPT)
		if !ok {
			t.Fatalf("instruction[%d] = %#v, want ModifyPT", i, instr.Primitive)
		}
		if modify.Object != game.TargetPermanentReference(0) {
			t.Fatalf("instruction[%d] object = %#v, want target 0", i, modify.Object)
		}
		if modify.Duration != game.DurationUntilEndOfTurn {
			t.Fatalf("instruction[%d] duration = %v, want until end of turn", i, modify.Duration)
		}
		if !instr.Condition.Exists || !instr.Condition.Val.Condition.Exists {
			t.Fatalf("instruction[%d] is ungated: %#v", i, instr)
		}
		if !instr.Condition.Val.Condition.Val.EventHistory.Exists {
			t.Fatalf("instruction[%d] condition is not event-history gated: %#v", i, instr.Condition.Val)
		}
	}
	base, ok := seq[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("base instruction = %#v, want ModifyPT", seq[0].Primitive)
	}
	morbid, ok := seq[1].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("morbid instruction = %#v, want ModifyPT", seq[1].Primitive)
	}
	if !seq[0].Condition.Val.Condition.Val.Negate {
		t.Fatal("base modification must be gated on NOT(a creature died this turn)")
	}
	if seq[1].Condition.Val.Condition.Val.Negate {
		t.Fatal("morbid modification must be gated on (a creature died this turn)")
	}
	if got := base.PowerDelta; got != game.Fixed(-1) {
		t.Fatalf("base power delta = %#v, want -1", got)
	}
	if got := morbid.PowerDelta; got != game.Fixed(-13) {
		t.Fatalf("morbid power delta = %#v, want -13", got)
	}
	if got := morbid.ToughnessDelta; got != game.Fixed(-13) {
		t.Fatalf("morbid toughness delta = %#v, want -13", got)
	}
}
