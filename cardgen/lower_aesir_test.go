package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceAesirEscapeValhalla exercises the
// cross-chapter linked-exile mechanism of a Saga (issue #1486): chapter I exiles
// a permanent card from the controller's graveyard under a source-keyed link and
// gains life equal to its mana value, chapter II places that many +1/+1 counters
// on a target, and chapter III returns both the Saga and the exiled card to
// hand. All three chapters must bind the same constant exile-graveyard-card key
// so the runtime threads the one exiled card through every chapter.
func TestGenerateExecutableCardSourceAesirEscapeValhalla(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "The Aesir Escape Valhalla",
		Layout:   "normal",
		TypeLine: "Enchantment — Saga",
		ManaCost: "{2}{R}",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter.)\n" +
			"I — Exile a permanent card from your graveyard. You gain life equal to its mana value.\n" +
			"II — Put a number of +1/+1 counters on target creature you control equal to the mana value of the exiled card.\n" +
			"III — Return this Saga and the exiled card to their owner's hand.",
		Colors: []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.ChooseFromZone{",
		`PublishLinked: game.LinkedKey("exile-graveyard-card"),`,
		"Primitive: game.GainLife{",
		"Kind:       game.DynamicAmountObjectManaValue,",
		`Object:     game.LinkedObjectReference("exile-graveyard-card"),`,
		"Primitive: game.AddCounter{",
		"Object:      game.TargetPermanentReference(0),",
		"CounterKind: counter.PlusOnePlusOne,",
		"Primitive: game.ReturnExiledCardsToHand{",
		`LinkedKey: game.LinkedKey("exile-graveyard-card"),`,
		"Primitive: game.Bounce{",
		"Object: game.SourcePermanentReference(),",
	} {
		if !strings.Contains(normalizeSource(source), normalizeSource(wanted)) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
