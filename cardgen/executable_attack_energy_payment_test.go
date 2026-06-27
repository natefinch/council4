package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceAttackEnergyPayment proves the Kaladesh energy
// cycle's "Whenever this creature attacks, you may pay {E}{E}. If you do, ..."
// rider lowers to a resolution payment whose energy additional cost gates the
// consequence (Thriving Rats, Aether Swooper, and the rest).
func TestGenerateExecutableCardSourceAttackEnergyPayment(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventAttackerDeclared",
		"game.TriggerSourceSelf",
		"game.ResolutionPayment{",
		"cost.AdditionalEnergy",
		"Amount: 2",
		"game.AddCounter{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
