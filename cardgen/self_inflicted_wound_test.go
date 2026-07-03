package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceSelfInflictedWound covers the affirmative
// non-controller edict → event-player consequence gate combined with a
// color-disjunction sacrifice selection. The edicted opponent sacrifices "a
// green or white creature of their choice" (a Selection.ColorsAny filter), the
// edict publishes whether it happened, and "they lose 2 life" resolves for the
// event player only when the sacrifice succeeded (TriTrue). It exercises two
// pieces working together: the color-disjunction sacrifice noun and the
// resolving-gate consequence whose "they" reference must survive alongside the
// gate's own "that player" anaphor without failing closed.
func TestGenerateExecutableCardSourceSelfInflictedWound(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Self-Inflicted Wound",
		Layout:     "normal",
		ManaCost:   "{B}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Target opponent sacrifices a green or white creature of their choice. If that player does, they lose 2 life.",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.SacrificePermanents",
		"Amount:    game.Fixed(1)",
		"Player:    game.TargetPlayerReference(0)",
		"RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green, color.White}",
		"PublishResult: game.ResultKey(\"if-you-do\")",
		"Primitive: game.LoseLife",
		"Amount: game.Fixed(2)",
		"Player: game.EventPlayerReference()",
		"Key:       \"if-you-do\"",
		"Succeeded: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
