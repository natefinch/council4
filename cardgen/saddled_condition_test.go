package cardgen

import (
	"strings"
	"testing"
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
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.SaddleActivatedAbility(3)",
		"game.EventAttackerDeclared",
		`game.Reveal{`,
		`PublishLinked: game.LinkedKey("revealed-card-1")`,
		"game.MoveCard{",
		`game.CardReference{Kind: game.CardReferenceLinked, LinkID: "revealed-card-1"}`,
		"FromZone:    zone.Library,",
		"Destination: zone.Hand,",
		"game.LoseLife{",
		"game.DynamicAmountObjectManaValue",
		`game.LinkedObjectReference("revealed-card-1")`,
		"game.ControllerReference()",
		"game.OpponentsReference()",
		"SourceSaddled: true,",
		"Negate:        true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
