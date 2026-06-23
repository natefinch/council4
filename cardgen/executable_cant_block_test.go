package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceCantBlockThisTurnSorcery(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Solo Block",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Sorcery",
		OracleText: "Target creature can't block this turn.",
		Colors:     []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target creature"`,
		"PermanentTypes: []types.Card{types.Creature}",
		"Primitive: game.ApplyRule{",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Kind: game.RuleEffectCantBlock,",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCantBlockThisTurnMultiTarget(t *testing.T) {
	t.Parallel()
	// "Up to three target creatures can't block this turn." (Unearthly Blizzard)
	// lowers to a single MinTargets 0 / MaxTargets 3 spec whose sequence applies
	// one can't-block restriction per target slot.
	card := &ScryfallCard{
		Name:       "Test Blizzard",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Up to three target creatures can't block this turn.",
		Colors:     []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"MaxTargets: 3,",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Object: opt.Val(game.TargetPermanentReference(1)),",
		"Object: opt.Val(game.TargetPermanentReference(2)),",
		"Kind: game.RuleEffectCantBlock,",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Count(source, "game.RuleEffectCantBlock") != 3 {
		t.Fatalf("expected three can't-block applications:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceCantBlockThisTurnFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording deviates from the exact "<targets> can't block this turn."
	// restriction, so generation must fail closed: it must emit at least one
	// diagnostic and never lower an ApplyRule / RuleEffectCantBlock for it.
	rejected := []string{
		"Creatures can't block.",
		"Target creature can't block.",
		"Target creature can't block this turn unless you pay {1}.",
		"Any number of target creatures can't block this turn.",
		"Target creature can't attack this turn.",
	}
	for _, oracle := range rejected {
		card := &ScryfallCard{
			Name:       "Test Fail Closed",
			Layout:     "normal",
			ManaCost:   "{R}",
			TypeLine:   "Sorcery",
			OracleText: oracle,
			Colors:     []string{"R"},
		}
		source, diagnostics, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatalf("GenerateExecutableCardSource(%q) err = %v", oracle, err)
		}
		if len(diagnostics) == 0 {
			t.Errorf("GenerateExecutableCardSource(%q) produced no diagnostics, want fail closed", oracle)
		}
		if strings.Contains(source, "game.RuleEffectCantBlock") {
			t.Errorf("GenerateExecutableCardSource(%q) lowered a can't-block rule effect, want fail closed:\n%s", oracle, source)
		}
	}
}
