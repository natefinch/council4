package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceGigglingSkitterspike proves the full
// Giggling Skitterspike card lowers with no diagnostics: the shared-subject
// three-way combat/target trigger union becomes three triggered abilities that
// each deal source-power damage to each opponent, and the Monstrosity activated
// ability and Indestructible static ability are preserved.
func TestGenerateExecutableCardSourceGigglingSkitterspike(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Giggling Skitterspike",
		Layout:   "normal",
		ManaCost: "{4}",
		TypeLine: "Artifact Creature — Toy",
		OracleText: "Indestructible\n" +
			"Whenever this creature attacks, blocks, or becomes the target of a spell, it deals damage equal to its power to each opponent.\n" +
			"{5}: Monstrosity 5. (If this creature isn't monstrous, put five +1/+1 counters on it and it becomes monstrous.)",
		Power:     new("1"),
		Toughness: new("1"),
	}, "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	// Three triggered abilities: one per event in the shared-subject union.
	for _, want := range []string{
		"game.EventAttackerDeclared,",
		"game.EventBlockerDeclared,",
		"game.EventObjectBecameTarget,",
		"game.StackSpell,",
		"game.IndestructibleStaticBody,",
		"game.Monstrosity{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
	// Each trigger deals source-power damage to each opponent.
	if got := strings.Count(source, "game.DynamicAmountObjectPower,"); got != 3 {
		t.Fatalf("source-power amounts = %d, want 3 (one per trigger):\n%s", got, source)
	}
	if got := strings.Count(source, "game.PlayerGroupDamageRecipient(game.OpponentsReference())"); got != 3 {
		t.Fatalf("each-opponent recipients = %d, want 3 (one per trigger):\n%s", got, source)
	}
	if got := strings.Count(source, "DamageSource: opt.Val(game.EventPermanentReference())"); got != 3 {
		t.Fatalf("event-permanent damage sources = %d, want 3 (one per trigger):\n%s", got, source)
	}
}
