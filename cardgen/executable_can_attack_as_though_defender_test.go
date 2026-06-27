package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceCanAttackAsThoughDefenderActivated covers a
// self grant "This creature can attack this turn as though it didn't have
// defender." on an activated ability (Glade Watcher's Formidable ability),
// lowering to an ApplyRule on the source permanent with the
// RuleEffectCanAttackAsThoughDefender rule effect.
func TestGenerateExecutableCardSourceCanAttackAsThoughDefenderActivated(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Glade Watcher",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Creature — Elemental",
		Colors:     []string{"G"},
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: "Defender\n{G}: This creature can attack this turn as though it didn't have defender.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"ActivatedAbilities:",
		"Primitive: game.ApplyRule{",
		"Object: opt.Val(game.SourcePermanentReference()),",
		"Kind: game.RuleEffectCanAttackAsThoughDefender,",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCanAttackAsThoughDefenderFailsClosed ensures
// wordings that deviate from the exact "<source> can attack this turn as though
// it didn't have defender." permission fail closed: generation must emit at
// least one diagnostic and never lower the rule effect.
func TestGenerateExecutableCardSourceCanAttackAsThoughDefenderFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"{G}: This creature can attack as though it didn't have defender.",
		"{G}: This creature can't attack this turn.",
		"{G}: This creature can attack this turn as though it weren't tapped.",
	}
	for _, oracle := range rejected {
		card := &ScryfallCard{
			Name:       "Test Fail Closed",
			Layout:     "normal",
			ManaCost:   "{1}{G}",
			TypeLine:   "Creature — Elemental",
			Colors:     []string{"G"},
			Power:      new("3"),
			Toughness:  new("3"),
			OracleText: oracle,
		}
		source, diagnostics, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatalf("GenerateExecutableCardSource(%q) err = %v", oracle, err)
		}
		if len(diagnostics) == 0 {
			t.Errorf("GenerateExecutableCardSource(%q) produced no diagnostics, want fail closed", oracle)
		}
		if strings.Contains(source, "game.RuleEffectCanAttackAsThoughDefender") {
			t.Errorf("GenerateExecutableCardSource(%q) lowered a can-attack-as-though-defender rule effect, want fail closed:\n%s", oracle, source)
		}
	}
}
