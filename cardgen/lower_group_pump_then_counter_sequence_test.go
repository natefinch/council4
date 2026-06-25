package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerGroupPumpThenGroupCounterSequence verifies the ordered pair "Other
// creatures you control get +2/+2 until end of turn. Put an indestructible
// counter on each of them." lowers to a group power/toughness pump followed by a
// fixed counter placement on that same back-referenced group.
func TestLowerGroupPumpThenGroupCounterSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Knightfall",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Other creatures you control get +2/+2 until end of turn. Put an indestructible counter on each of them.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want pump then counter placement", mode.Sequence)
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
		pump.PowerDelta != 2 ||
		pump.ToughnessDelta != 2 {
		t.Fatalf("pump effect = %+v, want +2/+2 power/toughness modify", pump)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.AddCounter", mode.Sequence[1].Primitive)
	}
	if add.CounterKind != counter.Indestructible {
		t.Fatalf("counter kind = %v, want indestructible", add.CounterKind)
	}
	if add.Amount.IsDynamic() || add.Amount.Value() != 1 {
		t.Fatalf("counter amount = %+v, want fixed 1", add.Amount)
	}
	// The counter's group must be exactly the pump's group so "each of them"
	// resolves to the just-pumped set.
	pumpSelection := pump.Group.Selection()
	addSelection := add.Group.Selection()
	if add.Group.Domain() != game.GroupDomainBattlefield ||
		add.Group.Domain() != pump.Group.Domain() ||
		addSelection.Controller != game.ControllerYou ||
		len(addSelection.RequiredTypes) != 1 ||
		addSelection.RequiredTypes[0] != types.Creature ||
		addSelection.Controller != pumpSelection.Controller {
		t.Fatalf("add group = %+v, want same controlled-creature group as pump %+v", addSelection, pumpSelection)
	}
}

// TestLowerSummonKnightsOfRoundChapter verifies the anchor Saga's final chapter
// "Other creatures you control get +2/+2 until end of turn. Put an indestructible
// counter on each of them." fully lowers as a chapter ability.
func TestLowerSummonKnightsOfRoundChapter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Summon: Knights of Round",
		Layout:   "saga",
		TypeLine: "Enchantment Creature — Saga Knight",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter.)\n" +
			"I, II, III, IV — Create three 2/2 white Knight creature tokens.\n" +
			"V — Ultimate End — Other creatures you control get +2/+2 until end of turn. Put an indestructible counter on each of them.",
		Power:     new("0"),
		Toughness: new("0"),
	})
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("chapter abilities = %d, want 2", len(face.ChapterAbilities))
	}
	final := face.ChapterAbilities[1]
	mode := final.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("final chapter sequence = %#v, want pump then counter", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous); !ok {
		t.Fatalf("sequence[0] = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok || add.CounterKind != counter.Indestructible {
		t.Fatalf("sequence[1] = %#v, want indestructible AddCounter", mode.Sequence[1].Primitive)
	}
}
