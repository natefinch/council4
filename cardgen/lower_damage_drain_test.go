package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDamageDealtThisWayDrain proves a "deals N damage to any target. You
// gain life equal to the damage dealt this way." drain (Corrupt) lowers to a
// damage instruction that publishes its dealt amount and a follow-on life gain
// that reads it.
func TestLowerDamageDealtThisWayDrain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Corrupt",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{5}{B}",
		OracleText: "Corrupt deals damage to any target equal to the number of Swamps you control. You gain life equal to the damage dealt this way.",
	})
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}
	damage, ok := sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("sequence[0] primitive = %T, want game.Damage", sequence[0].Primitive)
	}
	if !damage.Amount.IsDynamic() && damage.Amount.Value() == 0 {
		t.Fatal("damage amount is empty")
	}
	if sequence[0].PublishResult == "" {
		t.Fatal("damage instruction does not publish a result")
	}
	gain, ok := sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("sequence[1] primitive = %T, want game.GainLife", sequence[1].Primitive)
	}
	dyn := gain.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountPreviousEffectResult {
		t.Fatalf("gain amount = %#v, want DynamicAmountPreviousEffectResult", gain.Amount)
	}
	if dyn.Val.ResultKey != sequence[0].PublishResult {
		t.Fatalf("gain ResultKey = %q, damage PublishResult = %q; want equal", dyn.Val.ResultKey, sequence[0].PublishResult)
	}
}

// TestLowerExcessDamageDealtThisWayDrain proves Razor Rings' "You gain life
// equal to the excess damage dealt this way." reads the published excess
// damage of the preceding damage instruction.
func TestLowerExcessDamageDealtThisWayDrain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Razor Rings",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{W}",
		OracleText: "Razor Rings deals 4 damage to target attacking or blocking creature. You gain life equal to the excess damage dealt this way.",
	})
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}
	if sequence[0].PublishResult == "" {
		t.Fatal("damage instruction does not publish a result")
	}
	gain, ok := sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("sequence[1] primitive = %T, want game.GainLife", sequence[1].Primitive)
	}
	dyn := gain.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountPreviousEffectExcessDamage {
		t.Fatalf("gain amount = %#v, want DynamicAmountPreviousEffectExcessDamage", gain.Amount)
	}
	if dyn.Val.ResultKey != sequence[0].PublishResult {
		t.Fatalf("gain ResultKey = %q, damage PublishResult = %q; want equal", dyn.Val.ResultKey, sequence[0].PublishResult)
	}
}
