package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceRakdosPatronOfChaos covers the non-controller
// negative resolving gate: a non-controller optional edict ("target opponent may
// sacrifice ...") whose failure branch ("If they don't, ...") resolves for the
// controller. The offered sacrifice must be asked of the edicted opponent (its
// OptionalActor names the target player, not the controller), publish whether it
// happened, and the controller's draw must be gated on it having been declined
// (TriFalse). This is the resolving-failure mirror of the affirmative
// event-player payment gate (Smothering Tithe).
func TestGenerateExecutableCardSourceRakdosPatronOfChaos(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Rakdos, Patron of Chaos",
		Layout:     "normal",
		ManaCost:   "{4}{B}{R}",
		TypeLine:   "Legendary Creature — Demon",
		OracleText: "Flying, trample\nAt the beginning of your end step, target opponent may sacrifice two nonland, nontoken permanents of their choice. If they don't, you draw two cards.",
	}, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EventBeginningOfStep",
		"Constraint: \"target opponent\"",
		"Primitive: game.SacrificePermanents",
		"Amount:    game.Fixed(2)",
		"Player:    game.TargetPlayerReference(0)",
		"ExcludedTypes: []types.Card{types.Land}, NonToken: true",
		"Optional:      true",
		"OptionalActor: opt.Val(game.TargetPlayerReference(0))",
		"PublishResult: game.ResultKey(\"non-controller-optional-action\")",
		"Primitive: game.Draw",
		"Player: game.ControllerReference()",
		"Key:       \"non-controller-optional-action\"",
		"Succeeded: game.TriFalse",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
