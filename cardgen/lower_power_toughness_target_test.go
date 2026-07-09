package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
)

// TestLowerPowerToughnessTargetFilter proves a "target N/M creature" activated
// ability (Pendelhaven, Aegis of the Meek) lowers its target into a Selection
// pinned to current power N and toughness M, so the pump only reaches creatures
// matching both, and drives the asymmetric +1/+2 ModifyPT until end of turn.
func TestLowerPowerToughnessTargetFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Pendelhaven",
		Layout:   "normal",
		TypeLine: "Legendary Land",
		OracleText: "{T}: Add {G}.\n" +
			"{T}: Target 1/1 creature gets +1/+2 until end of turn.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	sel := mode.Targets[0].Selection
	if !sel.Exists {
		t.Fatalf("target selection missing on %#v", mode.Targets[0])
	}
	power := sel.Val.Power
	if !power.Exists || power.Val.Op != compare.Equal || power.Val.Value != 1 {
		t.Fatalf("target power = %#v, want exact power 1", power)
	}
	toughness := sel.Val.Toughness
	if !toughness.Exists || toughness.Val.Op != compare.Equal || toughness.Val.Value != 1 {
		t.Fatalf("target toughness = %#v, want exact toughness 1", toughness)
	}
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modify.PowerDelta != game.Fixed(1) || modify.ToughnessDelta != game.Fixed(2) {
		t.Fatalf("modify deltas = (%v,%v), want (+1,+2)", modify.PowerDelta, modify.ToughnessDelta)
	}
	if modify.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("modify duration = %v, want until end of turn", modify.Duration)
	}
}
