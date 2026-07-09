package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceEvelynTheCovetous covers Evelyn, the Covetous'
// full static play/cast-from-exile permission on top of the named-counter exile
// substrate. The ETB trigger "Whenever Evelyn or another Vampire you control
// enters, exile the top card of each player's library with a collection counter
// on it." lowers to an ExileTopOfLibrary over every player carrying a collection
// counter. The static "Once each turn, you may play a card from exile with a
// collection counter on it if it was exiled by an ability you controlled, and you
// may spend mana as though it were mana of any color to cast it." lowers to a
// paired RuleEffectPlayLandsFromZone and RuleEffectCastSpellsFromZone over the
// exile zone, each filtered by counter.Collection, restricted to cards the
// controller's own ability exiled (ExileCounterExiledByController), sharing a
// single OncePerTurn use, with the cast grant carrying SpendAnyMana. The whole
// card lowers without diagnostics.
func TestGenerateExecutableCardSourceEvelynTheCovetous(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Evelyn, the Covetous",
		Layout:   "normal",
		ManaCost: "{2}{U/B}{B}{B/R}",
		TypeLine: "Legendary Creature — Vampire Rogue",
		OracleText: "Flash\n" +
			"Whenever Evelyn or another Vampire you control enters, exile the top card of each player's library with a collection counter on it.\n" +
			"Once each turn, you may play a card from exile with a collection counter on it if it was exiled by an ability you controlled, and you may spend mana as though it were mana of any color to cast it.",
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.ExileTopOfLibrary{",
		"PlayerGroup: game.AllPlayersReference(),",
		"Counter:     opt.Val(counter.Collection),",
		"Kind:                           game.RuleEffectPlayLandsFromZone,",
		"PermanentTypes:                 []types.Card{types.Land},",
		"Kind:                           game.RuleEffectCastSpellsFromZone,",
		"CastFromZone:                   zone.Exile,",
		"ExileCounterFilter:             opt.Val(counter.Collection),",
		"ExileCounterExiledByController: true,",
		"OncePerTurn:                    true,",
		"SpendAnyMana:                   true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
