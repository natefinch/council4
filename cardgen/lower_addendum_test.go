package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerAddendumGroupBackReferenceBonus verifies the cross-ability "those
// creatures"/"they" plural group back-reference inside an Addendum paragraph
// resolves to the first paragraph's affected group (Unbreakable Formation). The
// first paragraph grants indestructible unconditionally; the Addendum paragraph
// places a +1/+1 counter on and grants vigilance to that same group, both gated
// on casting during the controller's main phase. The gate must be present on the
// Addendum bonus instructions and absent on the base grant, so the bonus applies
// only when the spell is cast during the main phase.
func TestLowerAddendumGroupBackReferenceBonus(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Unbreakable Formation",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Instant",
		OracleText: "Creatures you control gain indestructible until end of turn.\nAddendum — If you cast this spell during your main phase, put a +1/+1 counter on each of those creatures and they gain vigilance until end of turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %d, want 3", len(mode.Sequence))
	}

	controlledCreatures := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	})

	grant, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("instruction 0 = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if mode.Sequence[0].Condition.Exists {
		t.Fatal("base indestructible grant must not be gated on cast timing")
	}
	if len(grant.ContinuousEffects) != 1 ||
		!reflect.DeepEqual(grant.ContinuousEffects[0].AddKeywords, []game.Keyword{game.Indestructible}) {
		t.Fatalf("instruction 0 keywords = %#v, want [Indestructible]", grant.ContinuousEffects)
	}

	addCounter, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("instruction 1 = %T, want game.AddCounter", mode.Sequence[1].Primitive)
	}
	if addCounter.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want +1/+1", addCounter.CounterKind)
	}
	if !reflect.DeepEqual(addCounter.Group, controlledCreatures) {
		t.Fatalf("counter group = %#v, want creatures you control", addCounter.Group)
	}
	requireCastDuringMainGate(t, &mode.Sequence[1], "counter placement")

	vigilance, ok := mode.Sequence[2].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("instruction 2 = %T, want game.ApplyContinuous", mode.Sequence[2].Primitive)
	}
	if len(vigilance.ContinuousEffects) != 1 ||
		!reflect.DeepEqual(vigilance.ContinuousEffects[0].AddKeywords, []game.Keyword{game.Vigilance}) {
		t.Fatalf("instruction 2 keywords = %#v, want [Vigilance]", vigilance.ContinuousEffects)
	}
	if !reflect.DeepEqual(vigilance.ContinuousEffects[0].Group, controlledCreatures) {
		t.Fatalf("vigilance group = %#v, want creatures you control", vigilance.ContinuousEffects[0].Group)
	}
	requireCastDuringMainGate(t, &mode.Sequence[2], "vigilance grant")
}

func requireCastDuringMainGate(t *testing.T, instruction *game.Instruction, label string) {
	t.Helper()
	if !instruction.Condition.Exists {
		t.Fatalf("%s instruction is not gated on cast timing", label)
	}
	if !instruction.Condition.Val.Condition.Val.CastDuringControllerMainPhase {
		t.Fatalf("%s gate = %#v, want CastDuringControllerMainPhase", label, instruction.Condition.Val.Condition.Val)
	}
}
