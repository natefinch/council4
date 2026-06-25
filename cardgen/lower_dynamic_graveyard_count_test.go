package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerLifeBurstGainLifeSequence verifies that Life Burst lowers its
// "Target player gains 4 life, then gains 4 life for each card named Life Burst
// in each graveyard." into two GainLife instructions: a fixed 4 and a dynamic 4
// per card named Life Burst across every graveyard
// (DynamicAmountCardsNamedSourceInGraveyards).
func TestLowerLifeBurstGainLifeSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:          "Life Burst",
		Layout:        "normal",
		ManaCost:      "{1}{W}",
		TypeLine:      "Instant",
		ColorIdentity: []string{"W"},
		OracleText:    "Target player gains 4 life, then gains 4 life for each card named Life Burst in each graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Life Burst did not lower a spell ability")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two GainLife instructions", sequence)
	}
	first, ok := sequence[0].Primitive.(game.GainLife)
	if !ok || first.Amount.IsDynamic() || first.Amount.Value() != 4 {
		t.Fatalf("first instruction = %#v, want fixed 4 GainLife", sequence[0].Primitive)
	}
	second, ok := sequence[1].Primitive.(game.GainLife)
	if !ok || !second.Amount.IsDynamic() {
		t.Fatalf("second instruction = %#v, want dynamic GainLife", sequence[1].Primitive)
	}
	dynamic := second.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountCardsNamedSourceInGraveyards || dynamic.Multiplier != 4 {
		t.Fatalf("dynamic amount = %#v", dynamic)
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
}

// TestLowerGrowthCycleControllerGraveyardPump verifies that Growth Cycle lowers
// its "It gets an additional +2/+2 ... for each card named Growth Cycle in your
// graveyard." clause to a ModifyPT scaled by the controller's graveyard count
// (DynamicAmountCardsNamedSourceInControllerGraveyard), distinct from the
// all-graveyards kind.
func TestLowerGrowthCycleControllerGraveyardPump(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:          "Growth Cycle",
		Layout:        "normal",
		ManaCost:      "{1}{G}",
		TypeLine:      "Instant",
		ColorIdentity: []string{"G"},
		OracleText:    "Target creature gets +3/+3 until end of turn. It gets an additional +2/+2 until end of turn for each card named Growth Cycle in your graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Growth Cycle did not lower a spell ability")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two ModifyPT instructions", sequence)
	}
	first, ok := sequence[0].Primitive.(game.ModifyPT)
	if !ok || first.PowerDelta.IsDynamic() || first.PowerDelta.Value() != 3 {
		t.Fatalf("first instruction = %#v, want fixed +3/+3 ModifyPT", sequence[0].Primitive)
	}
	second, ok := sequence[1].Primitive.(game.ModifyPT)
	if !ok || !second.PowerDelta.IsDynamic() || !second.ToughnessDelta.IsDynamic() {
		t.Fatalf("second instruction = %#v, want dynamic ModifyPT", sequence[1].Primitive)
	}
	dynamic := second.PowerDelta.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountCardsNamedSourceInControllerGraveyard || dynamic.Multiplier != 2 {
		t.Fatalf("dynamic power delta = %#v", dynamic)
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
}

// TestLowerControllerGraveyardForeignNameFailsClosed verifies that a foreign
// card name in the "in your graveyard" count tail does not lower to a dynamic
// pump (the recognizer is text-blind to the card's own name only).
func TestLowerControllerGraveyardForeignNameFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:          "Growth Cycle",
		Layout:        "normal",
		ManaCost:      "{1}{G}",
		TypeLine:      "Instant",
		ColorIdentity: []string{"G"},
		OracleText:    "Target creature gets +3/+3 until end of turn. It gets an additional +2/+2 until end of turn for each card named Lightning Bolt in your graveyard.",
	}, "g")
	if err == nil && len(diagnostics) == 0 {
		t.Fatal("expected fail-closed diagnostics for a foreign graveyard name")
	}
}
