package cardgen

import (
	"strings"
	"testing"
)

// brothersWarCard is The Brothers' War, whose chapter II directed forced-attack
// and chapter III dynamic two-target damage exercise the two new compiler
// surfaces this package adds.
func brothersWarCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "The Brothers' War",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		ManaCost: "{2}{R}",
		Colors:   []string{"R"},
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Create two tapped Powerstone tokens.\n" +
			"II — Choose two target players. Until your next turn, each creature they control attacks the other chosen player each combat if able.\n" +
			"III — This Saga deals X damage to any target and X damage to any other target, where X is the number of artifacts you control.",
	}
}

// TestGenerateExecutableCardSourceBrothersWar asserts the whole Saga generates
// without diagnostics and that chapters II and III lower to their expected
// structures: chapter II to a reciprocal pair of directed RuleEffectMustAttack
// rule effects over two distinct player targets, and chapter III to two Damage
// instructions sharing one artifact-count dynamic amount across two targets.
func TestGenerateExecutableCardSourceBrothersWar(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(brothersWarCard(), "brotherswar")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for wanted, count := range map[string]int{
		"AffectedPlayerRef:":               2,
		"RequiredAttackTargetRef:":         2,
		"game.TargetPlayerReference(0)":    2,
		"game.TargetPlayerReference(1)":    2,
		"game.RuleEffectMustAttack":        2,
		"game.AnyTargetDamageRecipient(0)": 1,
		"game.AnyTargetDamageRecipient(1)": 1,
		"game.DynamicAmountCountSelector":  2,
	} {
		if got := strings.Count(source, wanted); got != count {
			t.Fatalf("source has %d occurrences of %q, want %d:\n%s", got, wanted, count, source)
		}
	}
	// Both the chapter II second player and the chapter III "any other target"
	// slot are distinct from their prior target.
	if got := strings.Count(source, "DistinctFromPriorTargets: true"); got != 2 {
		t.Fatalf("expected 2 distinct target slots, got %d:\n%s", got, source)
	}
}

// TestGenerateExecutableCardSourceDirectedMustAttackFailsClosed asserts the
// directed forced-attack lowering binds only the exact two-target-player chapter
// II wording. A near-miss that names a single player group ("each creature you
// control") routes to the ordinary group forced-attack path and must not emit the
// directed reciprocal structure.
func TestGenerateExecutableCardSourceDirectedMustAttackFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Goad Test",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Sorcery",
		Colors:     []string{"R"},
		OracleText: "Until your next turn, creatures you control attack each combat if able.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "goadtest")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Contains(source, "AffectedPlayerRef") || strings.Contains(source, "RequiredAttackTargetRef") {
		t.Fatalf("group forced-attack must not emit directed fields:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceDynamicTwoTargetDamage covers the chapter III
// "deals X damage to any target and X damage to any other target, where X is ..."
// shape in isolation, confirming both Damage instructions share the dynamic
// amount and the second slot is distinct.
func TestGenerateExecutableCardSourceDynamicTwoTargetDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Forge Bolt",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Sorcery",
		Colors:     []string{"R"},
		OracleText: "Forge Bolt deals X damage to any target and X damage to any other target, where X is the number of artifacts you control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "forgebolt")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if got := strings.Count(source, "Primitive: game.Damage"); got != 2 {
		t.Fatalf("expected 2 damage instructions, got %d:\n%s", got, source)
	}
	if got := strings.Count(source, "game.DynamicAmountCountSelector"); got != 2 {
		t.Fatalf("expected both damage amounts to be the artifact-count dynamic, got %d:\n%s", got, source)
	}
	if got := strings.Count(source, "DistinctFromPriorTargets: true"); got != 1 {
		t.Fatalf("expected exactly one distinct target slot, got %d:\n%s", got, source)
	}
}
