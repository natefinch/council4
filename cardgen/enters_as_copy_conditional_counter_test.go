package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceEntersAsCopyConditionalCounters(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Spark Double",
		Layout:     "normal",
		ManaCost:   "{1}{U}{U}",
		TypeLine:   "Legendary Creature — Shapeshifter",
		OracleText: "You may have this creature enter as a copy of a creature or planeswalker you control, except it enters with an additional +1/+1 counter on it if it's a creature, it enters with an additional loyalty counter on it if it's a planeswalker, and it isn't legendary.",
		Colors:     []string{"U"},
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntersAsCopyReplacement(",
		"RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}",
		"Controller: game.ControllerYou",
		"[]game.ConditionalCounterPlacement{",
		"game.ConditionalCounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1, IfType: types.Creature}",
		"game.ConditionalCounterPlacement{Kind: counter.Loyalty, Amount: 1, IfType: types.Planeswalker}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
