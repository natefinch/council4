package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceHighTide(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "High Tide",
		Layout:     "normal",
		ManaCost:   "{U}",
		TypeLine:   "Instant",
		OracleText: "Until end of turn, whenever a player taps an Island for mana, that player adds an additional {U}.",
	}, "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// The spell sets up a repeating this-turn delayed trigger.
		"game.CreateDelayedTrigger{",
		"Window: game.DelayedWindowThisTurn",
		// The inner trigger fires off the authoritative mana-produced event,
		// filtered to Islands tapped for mana.
		"game.EventManaProduced",
		"RequireTappedForMana: true",
		`game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}}`,
		// Each tapping player adds an additional {U}.
		"game.AddMana{",
		"ManaColor: mana.U",
		"game.EventPlayerReference()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
