package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerNoMaximumHandSizeForRestOfGame proves that "Draw cards equal to the
// number of cards in your hand plus one. You have no maximum hand size for the
// rest of the game." lowers to a draw whose dynamic amount carries the offset
// followed by a permanent ApplyRule removing the controller's maximum hand size
// (Sea Gate Restoration).
func TestLowerNoMaximumHandSizeForRestOfGame(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sea Gate Restoration",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw cards equal to the number of cards in your hand plus one. You have no maximum hand size for the rest of the game.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two primitives", mode.Sequence)
	}
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("first primitive = %T, want game.Draw", mode.Sequence[0].Primitive)
	}
	if !draw.Amount.IsDynamic() {
		t.Fatalf("draw amount = %#v, want a dynamic amount", draw.Amount)
	}
	dynamic := draw.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountCountCardsInZone {
		t.Fatalf("draw amount kind = %v, want DynamicAmountCountCardsInZone", dynamic.Kind)
	}
	if dynamic.Addend != 1 {
		t.Fatalf("draw amount addend = %d, want 1", dynamic.Addend)
	}
	apply, ok := mode.Sequence[1].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("second primitive = %T, want game.ApplyRule", mode.Sequence[1].Primitive)
	}
	if apply.Duration != game.DurationPermanent {
		t.Fatalf("duration = %v, want DurationPermanent", apply.Duration)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectNoMaximumHandSize {
		t.Fatalf("kind = %v, want RuleEffectNoMaximumHandSize", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
}
