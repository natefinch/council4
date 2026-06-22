package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceControlledCreaturesCantBeBlocked verifies that
// the unconditional mass-evasion static "Creatures you control can't be blocked."
// lowers to a battlefield static ability carrying a single can't-be-blocked rule
// effect scoped to the controller's creatures, with no per-creature target.
func TestGenerateExecutableCardSourceControlledCreaturesCantBeBlocked(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Mass Evasion",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Enchantment",
		OracleText: "Creatures you control can't be blocked.",
		Colors:     []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"StaticAbilities: []game.StaticAbility{",
		"RuleEffects: []game.RuleEffect{",
		"Kind:               game.RuleEffectCantBeBlocked,",
		"AffectedController: game.ControllerYou,",
		"PermanentTypes:     []types.Card{types.Creature},",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceControlledCreaturesCantBeBlockedFailsClosed
// verifies that filtered or comparison-scoped mass-evasion wordings, which have
// no runtime affected-permanent predicate, never lower a controller-scoped
// can't-be-blocked rule effect.
func TestGenerateExecutableCardSourceControlledCreaturesCantBeBlockedFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Creatures you control with power greater than Test Mass Evasion's power can't be blocked.",
		"Creatures you control with flying can't be blocked.",
		"Goblins you control can't be blocked.",
	}
	for _, oracle := range rejected {
		card := &ScryfallCard{
			Name:       "Test Mass Evasion",
			Layout:     "normal",
			ManaCost:   "{2}{U}",
			TypeLine:   "Enchantment",
			OracleText: oracle,
			Colors:     []string{"U"},
		}
		source, _, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatalf("GenerateExecutableCardSource(%q) err = %v", oracle, err)
		}
		if strings.Contains(source, "AffectedController: game.ControllerYou") &&
			strings.Contains(source, "game.RuleEffectCantBeBlocked") {
			t.Errorf("GenerateExecutableCardSource(%q) lowered a controller-scoped can't-be-blocked rule effect, want fail closed:\n%s", oracle, source)
		}
	}
}
