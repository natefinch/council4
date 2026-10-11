package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestGenerateExecutableCardSourceCausticBronco exercises the saddled-state
// per-effect conditional: "You lose life equal to that card's mana value if this
// creature isn't saddled. Otherwise, each opponent loses that much life." The
// reveal-to-hand prelude plus the two mutually exclusive life-loss branches must
// all lower, with the controller branch gated on the source not being saddled
// and the each-opponent branch gated on the negation.
func TestGenerateExecutableCardSourceCausticBronco(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Caustic Bronco",
		Layout:     "normal",
		TypeLine:   "Creature — Horse",
		ManaCost:   "{1}{B}",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "Whenever this creature attacks, reveal the top card of your library and put it into your hand. You lose life equal to that card's mana value if this creature isn't saddled. Otherwise, each opponent loses that much life.\nSaddle 3 (Tap any number of other creatures you control with total power 3 or more: This Mount becomes saddled until end of turn. Saddle only as a sorcery.)",
	}
	assertCardPaths(t, card,
		"ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.BecomeSaddled)",
		"TriggeredAbilities[0].Trigger.Pattern.Event = game.EventAttackerDeclared",
		`Sequence[0].Primitive.(game.Reveal).PublishLinked = "sequence-effect-0-product"`,
		`Sequence[0].Primitive.(game.Reveal).PublishCharacteristics.ManaValue = "sequence-effect-0-mana-value"`,
		`Sequence[1].Primitive.(game.MoveCard).Card.LinkID = "sequence-effect-0-product"`,
		"Sequence[1].Primitive.(game.MoveCard).FromZone = zone.Library",
		"Sequence[1].Primitive.(game.MoveCard).Destination = zone.Hand",
		"Sequence[2].Condition.Val.Condition.Val.SourceSaddled = true",
		"Sequence[2].Condition.Val.Condition.Val.Negate = true",
		"Sequence[3].Condition.Val.Condition.Val.SourceSaddled = true",
	)
	sequence := lowerSingleFace(t, card).TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 4 {
		t.Fatalf("instructions=%d, want four complete Bronco clauses", len(sequence))
	}
	for _, instruction := range sequence[2:] {
		loss, ok := instruction.Primitive.(game.LoseLife)
		if !ok {
			t.Fatalf("consumer=%T, want LoseLife", instruction.Primitive)
		}
		amount := loss.Amount.DynamicAmount()
		if !amount.Exists || amount.Val.Kind != game.DynamicAmountPreviousEffectResult ||
			amount.Val.ResultKey != "sequence-effect-0-mana-value" ||
			!instruction.ResultGate.Exists || !instruction.ResultGate.Val.AmountAvailable ||
			instruction.ResultGate.Val.Key != amount.Val.ResultKey {
			t.Fatalf("postmove characteristic lost its frozen availability: %#v", instruction)
		}
	}
	if !sequence[0].LocalProducts.HasLink("sequence-effect-0-product") ||
		!sequence[0].LocalProducts.HasResult("sequence-effect-0-mana-value") {
		t.Fatal("observation identity/scalar do not have independent local lifetimes")
	}
}
