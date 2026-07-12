package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// dualModifyPTSequence lowers an instant whose body is two power/toughness
// modifications joined by "and" and returns the two ordered ModifyPT
// instructions together with the mode's target specs.
func dualModifyPTSequence(t *testing.T, oracleText string) (targets []game.TargetSpec, first, second game.ModifyPT) {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dual Pump",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Instant",
		OracleText: oracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two modify-PT instructions", mode.Sequence)
	}
	first, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("instruction 0 = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	second, ok = mode.Sequence[1].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("instruction 1 = %T, want game.ModifyPT", mode.Sequence[1].Primitive)
	}
	return mode.Targets, first, second
}

// TestLowerDualModifyPTOpposingTargets covers the canonical two-step "buff one
// creature and debuff another" sequence (Skulduggery): each modification owns
// its own target with the correct controller restriction, and the deltas apply
// to the matching target until end of turn.
func TestLowerDualModifyPTOpposingTargets(t *testing.T) {
	t.Parallel()
	targets, first, second := dualModifyPTSequence(t,
		"Until end of turn, target creature you control gets +1/+1 and target creature an opponent controls gets -1/-1.")
	if len(targets) != 2 {
		t.Fatalf("targets = %+v, want two single-creature targets", targets)
	}
	if targets[0].Selection.Val.Controller != game.ControllerYou {
		t.Fatalf("target[0] controller = %v, want you", targets[0].Selection.Val.Controller)
	}
	if targets[1].Selection.Val.Controller != game.ControllerOpponent {
		t.Fatalf("target[1] controller = %v, want opponent", targets[1].Selection.Val.Controller)
	}
	if first.Object != game.TargetPermanentReference(0) ||
		first.PowerDelta.Value() != 1 || first.ToughnessDelta.Value() != 1 ||
		first.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("instruction 0 = %+v, want +1/+1 on target 0 until end of turn", first)
	}
	if second.Object != game.TargetPermanentReference(1) ||
		second.PowerDelta.Value() != -1 || second.ToughnessDelta.Value() != -1 ||
		second.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("instruction 1 = %+v, want -1/-1 on target 1 until end of turn", second)
	}
}

// TestLowerDualModifyPTAnotherTarget covers "another target creature" (Rookie
// Mistake): the second modification's target must be distinct from the first.
func TestLowerDualModifyPTAnotherTarget(t *testing.T) {
	t.Parallel()
	targets, first, second := dualModifyPTSequence(t,
		"Until end of turn, target creature gets +0/+2 and another target creature gets -2/-0.")
	if len(targets) != 2 {
		t.Fatalf("targets = %+v, want two single-creature targets", targets)
	}
	if targets[0].DistinctFromPriorTargets {
		t.Fatalf("target[0] = %+v, want no distinctness restriction", targets[0])
	}
	if targets[0].Selection.Exists && targets[0].Selection.Val.ExcludeSource {
		t.Fatalf("target[0] = %+v, want no self-exclusion", targets[0])
	}
	if !targets[1].DistinctFromPriorTargets {
		t.Fatalf("target[1] = %+v, want distinct (another) restriction", targets[1])
	}
	if targets[1].Selection.Exists && targets[1].Selection.Val.ExcludeSource {
		t.Fatalf("target[1] = %+v, want cross-target distinctness, not self-exclusion", targets[1])
	}
	if first.PowerDelta.Value() != 0 || first.ToughnessDelta.Value() != 2 {
		t.Fatalf("instruction 0 = %+v, want +0/+2", first)
	}
	if second.PowerDelta.Value() != -2 || second.ToughnessDelta.Value() != 0 {
		t.Fatalf("instruction 1 = %+v, want -2/-0", second)
	}
}

// TestLowerCounterThenProliferatePair covers the "put a counter on target
// creature, then proliferate" sequence (Grim Affliction, Courage in Crisis):
// the generic ordered-sequence lowerer now sequences it faithfully into an
// AddCounter on the lone target followed by a standalone Proliferate.
func TestLowerCounterThenProliferatePair(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Counter Proliferate",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Instant",
		OracleText: "Put a -1/-1 counter on target creature, then proliferate.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1 (the counter recipient only)", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	add, addOK := mode.Sequence[0].Primitive.(game.AddCounter)
	_, prolifOK := mode.Sequence[1].Primitive.(game.Proliferate)
	if !addOK || !prolifOK {
		t.Fatalf("primitives = %T, %T; want game.AddCounter, game.Proliferate",
			mode.Sequence[0].Primitive, mode.Sequence[1].Primitive)
	}
	if add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("add.Object = %+v, want target 0", add.Object)
	}
	if add.CounterKind != counter.MinusOneMinusOne {
		t.Fatalf("add.CounterKind = %v, want -1/-1", add.CounterKind)
	}
}
