package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceSpellCreatesThisTurnDelayedTrigger covers a
// sorcery whose second paragraph is a "Whenever ... this turn" trigger (Forth
// Eorlingas!): on a spell such a paragraph is not a standing triggered ability but
// a delayed trigger the spell schedules as it resolves, so it merges into the
// spell as a CreateDelayedTrigger alongside the first paragraph's effect.
func TestGenerateExecutableCardSourceSpellCreatesThisTurnDelayedTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Test Riders of Rohan",
		Layout:   "normal",
		ManaCost: "{2}{R}{W}",
		TypeLine: "Sorcery",
		OracleText: "Create X 2/2 red Human Knight creature tokens with trample and haste.\n" +
			"Whenever one or more creatures you control deal combat damage to one or more players this turn, you become the monarch.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"SpellAbility:",
		"Primitive: game.CreateToken{",
		"Primitive: game.CreateDelayedTrigger{",
		"RequireCombatDamage:",
		"game.DelayedWindowThisTurn,",
		"Primitive: game.BecomeMonarch{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
