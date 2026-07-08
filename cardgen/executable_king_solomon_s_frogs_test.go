package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableKingSolomonSFrogs exercises the per-opponent
// distributive exile mechanism (King Solomon's Frogs): an ETB triggered ability,
// gated on "if you cast it", that exiles up to one target permanent with mana
// value 3 or greater that each opponent controls under a link, paired with a
// per-controller payoff that draws one card for each exiled permanent's
// last-known controller. The exile and the draw payoff must read the same
// exiled-for-each-opponent key so the runtime binding pairs them. The separate
// {3},{T},Exile-self activated ability makes the controller the monarch.
func TestGenerateExecutableKingSolomonSFrogs(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "King Solomon's Frogs",
		Layout:   "normal",
		ManaCost: "{3}{W}",
		TypeLine: "Legendary Artifact",
		OracleText: "Flash\n" +
			"When King Solomon's Frogs enters, if you cast it, for each opponent, exile up to one target permanent that player controls with mana value 3 or greater. For each permanent exiled this way, its controller draws a card.\n" +
			"{3}, {T}, Exile King Solomon's Frogs: You become the monarch.",
		Colors: []string{"W"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "k")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		// The ETB trigger is gated on the "if you cast it" intervening condition.
		"InterveningIfEventPermanentWasCastByController: true,",
		// The paired distributive exile-per-opponent under a link.
		"Primitive: game.ExileForEachOpponent{",
		"Chooser:   game.ControllerReference(),",
		"Selection: game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})},",
		`LinkedKey: game.LinkedKey("exiled-for-each-opponent"),`,
		// The per-controller draw payoff consuming the same link.
		"Primitive: game.DrawForEachExiled{",
		// The separate become-monarch activated ability.
		"Primitive: game.BecomeMonarch{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	// The exile primitive must precede the draw payoff in the sequence so the link
	// is published before it is consumed.
	exileIdx := strings.Index(source, "game.ExileForEachOpponent{")
	drawIdx := strings.Index(source, "game.DrawForEachExiled{")
	if exileIdx < 0 || drawIdx < 0 || exileIdx > drawIdx {
		t.Fatalf("exile (%d) must precede draw (%d):\n%s", exileIdx, drawIdx, source)
	}
}
