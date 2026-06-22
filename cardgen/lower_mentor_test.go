package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceMentorKeyword(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Mentor (Whenever this creature attacks, put a +1/+1 counter on target attacking creature with lesser power.)",
		"Mentor",
	} {
		card := &ScryfallCard{
			Name:       "Test Mentor",
			Layout:     "normal",
			TypeLine:   "Creature — Soldier",
			OracleText: oracle,
		}
		source, diagnostics, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatalf("oracle %q: %v", oracle, err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("oracle %q: diagnostics = %#v", oracle, diagnostics)
		}
		for _, want := range []string{
			"Primitive: game.AddCounter",
			"CounterKind: counter.PlusOnePlusOne",
			"CombatState:         game.CombatStateAttacking,",
			"PowerLessThanSource: true,",
		} {
			if !strings.Contains(source, want) {
				t.Fatalf("oracle %q: source missing %q:\n%s", oracle, want, source)
			}
		}
	}
}
