package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateSummonFenrir verifies the Final Fantasy Saga "Summon: Fenrir"
// lowers end to end. Chapter I is a basic-land search. Chapter II ("When you
// next cast a creature spell this turn, that creature enters with an additional
// +1/+1 counter on it.") is a one-shot event delayed trigger whose body lowers
// to a CreateReplacement bound to the future-cast spell's stack object: an
// enters-the-battlefield replacement, lasting until end of turn, that adds a
// +1/+1 counter when that creature enters. Chapter III ("Draw a card if you
// control the creature with the greatest power or tied for the greatest power.")
// is a conditional draw gated by the greatest-power condition.
func TestGenerateSummonFenrir(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Summon: Fenrir",
		Layout:   "saga",
		ManaCost: "{2}{G}",
		TypeLine: "Enchantment — Saga",
		Colors:   []string{"G"},
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Crescent Fang — Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.\n" +
			"II — Heavenward Howl — When you next cast a creature spell this turn, that creature enters with an additional +1/+1 counter on it.\n" +
			"III — Ecliptic Growl — Draw a card if you control the creature with the greatest power or tied for the greatest power.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// Chapter II: one-shot delayed cast trigger producing the replacement.
		"Primitive: game.CreateDelayedTrigger",
		"game.EventSpellCast,",
		"OneShot: true,",
		"Primitive: game.CreateReplacement{",
		"game.EventPermanentEnteredBattlefield,",
		"game.EventStackObjectReference(),",
		"game.DurationUntilEndOfTurn,",
		"EntersWithCounters: []game.CounterPlacement{game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}}",
		// Chapter III: conditional draw on the greatest-power predicate.
		"Primitive: game.Draw{",
		"ControllerControlsGreatestPowerCreature: true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
