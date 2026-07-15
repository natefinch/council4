package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceDannyPink proves the full Danny Pink card
// generates cleanly: the Mentor keyword lowers to its attack trigger and the
// group grant lowers to a continuous LayerAbility effect over the controller's
// creatures whose granted quoted ability is a coalesced, first-time-each-turn,
// any-counter placement trigger drawing for that creature's controller.
func TestGenerateExecutableCardSourceDannyPink(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Danny Pink",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Human Soldier Advisor",
		OracleText: "Mentor (Whenever this creature attacks, put a +1/+1 counter on target attacking creature with lesser power.)\n" +
			`Creatures you control have "Whenever one or more counters are put on this creature for the first time each turn, draw a card."`,
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// Mentor remains supported: its own attack trigger.
		"Event:  game.EventAttackerDeclared",
		"Primitive: game.AddCounter",
		"CounterKind: counter.PlusOnePlusOne",
		// Group grant: a continuous ability layer over controlled creatures.
		"Layer: game.LayerAbility",
		"Group: game.ObjectControlledGroup(",
		"AddAbilities:",
		// Granted quoted ability: any-counter, coalesced, first-time-each-turn.
		"Event:     game.EventCountersAdded",
		"Source:    game.TriggerSourceSelf",
		"OneOrMore: true",
		"MaxTriggersPerTurn: 1",
		// Draws for that creature's controller.
		"Player: game.ControllerReference()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
	// The granted trigger matches any counter kind: no counter-kind filter is
	// emitted on the granted ability's pattern.
	if strings.Contains(source, "MatchCounterKind: true") {
		t.Fatalf("granted trigger unexpectedly filtered by counter kind:\n%s", source)
	}
}
