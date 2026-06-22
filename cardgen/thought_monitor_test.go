package cardgen

import (
	"strings"
	"testing"
)

const thoughtMonitorOracle = "Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)\nFlying\nWhen this creature enters, draw two cards."

// TestGenerateExecutableCardSourceThoughtMonitor confirms the Affinity keyword
// lowers into an AffectedSource spell cost modifier whose per-object reduction
// counts the artifacts the caster controls, alongside the card's other abilities.
func TestGenerateExecutableCardSourceThoughtMonitor(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Thought Monitor",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Thopter",
		OracleText: thoughtMonitorOracle,
		ManaCost:   "{6}{U}",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"AffectedSource: true,",
		"PerObjectReduction: 1,",
		"RequiredTypes: []types.Card{types.Artifact}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
