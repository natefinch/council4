package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceGrolnokTheOmnivore covers the generic "exile
// with a named counter, then play/cast cards in exile with that counter"
// mechanic on Grolnok, the Omnivore. The library-to-graveyard trigger "exile it
// with a croak counter on it" lowers to a MoveCard into exile carrying a Croak
// counter rider, and the static "You may play lands and cast spells from among
// cards you own in exile with croak counters on them." lowers to a paired
// RuleEffectPlayLandsFromZone (lands only) and RuleEffectCastSpellsFromZone over
// the exile zone, each filtered by an ExileCounterFilter of counter.Croak. The
// whole card lowers without diagnostics.
func TestGenerateExecutableCardSourceGrolnokTheOmnivore(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Grolnok, the Omnivore",
		Layout:   "normal",
		ManaCost: "{2}{G}{U}",
		TypeLine: "Legendary Creature — Frog",
		OracleText: "Whenever a Frog you control attacks, mill three cards.\n" +
			"Whenever a permanent card is put into your graveyard from your library, exile it with a croak counter on it.\n" +
			"You may play lands and cast spells from among cards you own in exile with croak counters on them.",
	}, "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.MoveCard{",
		"Destination: zone.Exile,",
		"Counter:     opt.Val(counter.Croak),",
		"Kind:               game.RuleEffectPlayLandsFromZone,",
		"PermanentTypes:     []types.Card{types.Land},",
		"Kind:               game.RuleEffectCastSpellsFromZone,",
		"CastFromZone:       zone.Exile,",
		"ExileCounterFilter: opt.Val(counter.Croak),",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
