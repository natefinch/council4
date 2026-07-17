package cardgen

import (
	"strings"
	"testing"
)

// TestGeneratePadeem verifies Padeem, Consul of Innovation lowers end to end.
// Its first ability, "Artifacts you control have hexproof.", reuses the existing
// static controlled-group keyword grant. Its upkeep trigger carries a true
// intervening-if — "if you control the artifact with the greatest mana value or
// tied for the greatest mana value" — that lowers to the generic
// ControlsGreatestManaValueInGroup condition filtered to artifacts, gating a
// single card draw.
func TestGeneratePadeem(t *testing.T) {
	t.Parallel()
	pt := func(s string) *string { return &s }
	card := &ScryfallCard{
		Name:     "Padeem, Consul of Innovation",
		Layout:   "normal",
		ManaCost: "{3}{U}",
		TypeLine: "Legendary Creature — Vedalken Artificer",
		Colors:   []string{"U"},
		OracleText: "Artifacts you control have hexproof. (They can't be the targets of spells or abilities your opponents control.)\n" +
			"At the beginning of your upkeep, if you control the artifact with the greatest mana value or tied for the greatest mana value, draw a card.",
		Power:     pt("1"),
		Toughness: pt("4"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// Static hexproof grant over controlled artifacts.
		"Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact}}),",
		"game.Hexproof,",
		// Upkeep trigger with a true intervening-if condition.
		"Event:      game.EventBeginningOfStep,",
		"Step:       game.StepUpkeep,",
		"InterveningCondition: opt.Val(game.Condition{",
		"ControlsGreatestManaValueInGroup: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),",
		"Primitive: game.Draw{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
