package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerGroupPumpThenGroupUntapSequence verifies the ordered pair "Creatures
// you control get +1/+1 until end of turn. Untap them." lowers to a group
// power/toughness pump followed by a mass untap of that same back-referenced
// group (Rallying Roar).
func TestLowerGroupPumpThenGroupUntapSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Rallying Roar",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Creatures you control get +1/+1 until end of turn. Untap them.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want pump then untap", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok || apply.Object.Exists || apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("apply = %#v, want unanchored group pump until end of turn", mode.Sequence[0].Primitive)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(apply.ContinuousEffects))
	}
	pump := apply.ContinuousEffects[0]
	if pump.Layer != game.LayerPowerToughnessModify ||
		pump.PowerDelta != 1 ||
		pump.ToughnessDelta != 1 {
		t.Fatalf("pump effect = %+v, want +1/+1 power/toughness modify", pump)
	}
	untap, ok := mode.Sequence[1].Primitive.(game.Untap)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.Untap", mode.Sequence[1].Primitive)
	}
	if untap.Object != (game.ObjectReference{}) || untap.ChooseUpTo || untap.ChooseOne || untap.Amount.Value() != 0 {
		t.Fatalf("untap = %#v, want a plain mass group untap", untap)
	}
	// The untap's group must be exactly the pump's group so "them" resolves to
	// the just-pumped set.
	pumpSelection := pump.Group.Selection()
	untapSelection := untap.Group.Selection()
	if untap.Group.Domain() != game.GroupDomainBattlefield ||
		untap.Group.Domain() != pump.Group.Domain() ||
		untapSelection.Controller != game.ControllerYou ||
		len(untapSelection.RequiredTypes) != 1 ||
		untapSelection.RequiredTypes[0] != types.Creature ||
		untapSelection.Controller != pumpSelection.Controller {
		t.Fatalf("untap group = %+v, want same controlled-creature group as pump %+v", untapSelection, pumpSelection)
	}
}

// TestLowerGroupPumpThenGroupUntapThoseCreatures verifies the "Untap those
// creatures." wording of the same back-reference lowers identically (Jeskai
// Ascendancy's trigger, Gleam of Resistance, War Flare).
func TestLowerGroupPumpThenGroupUntapThoseCreatures(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test War Flare",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Creatures you control get +2/+1 until end of turn. Untap those creatures.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want pump then untap", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous); !ok {
		t.Fatalf("sequence[0] = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	untap, ok := mode.Sequence[1].Primitive.(game.Untap)
	if !ok || untap.ChooseUpTo || untap.ChooseOne {
		t.Fatalf("sequence[1] = %#v, want a plain mass group untap", mode.Sequence[1].Primitive)
	}
}
