package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceCantAttackThisTurn(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Solo Attack",
		Layout:     "normal",
		ManaCost:   "{W}",
		TypeLine:   "Sorcery",
		OracleText: "Target creature can't attack this turn.",
		Colors:     []string{"W"},
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
		"RequiredTypesAny: []types.Card{types.Creature}",
		"Primitive: game.ApplyRule{",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Kind: game.RuleEffectCantAttack,",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "game.RuleEffectCantBlock") {
		t.Fatalf("can't-attack lowered an unexpected can't-block rule effect:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceCantAttackOrBlockThisTurn(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Off Balance",
		Layout:     "normal",
		ManaCost:   "{W}",
		TypeLine:   "Instant",
		OracleText: "Target creature can't attack or block this turn.",
		Colors:     []string{"W"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	// The combined restriction applies both a can't-attack and a can't-block rule
	// effect inside one ApplyRule on the single targeted creature.
	for _, wanted := range []string{
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Kind: game.RuleEffectCantAttack,",
		"Kind: game.RuleEffectCantBlock,",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if got := strings.Count(source, "game.ApplyRule{"); got != 1 {
		t.Fatalf("expected one ApplyRule, got %d:\n%s", got, source)
	}
}

func TestGenerateExecutableCardSourceTargetMustAttackThisTurn(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Forced Attack",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Creature",
		OracleText: "{2}{R}: Target creature attacks this turn if able.",
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
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Kind: game.RuleEffectMustAttack,",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCantAttackMultiTarget(t *testing.T) {
	t.Parallel()
	// "Up to two target creatures can't attack this turn." lowers to a single
	// MinTargets 0 / MaxTargets 2 spec whose sequence applies one can't-attack
	// restriction per target slot.
	card := &ScryfallCard{
		Name:       "Test Multi Attack",
		Layout:     "normal",
		ManaCost:   "{1}{W}",
		TypeLine:   "Sorcery",
		OracleText: "Up to two target creatures can't attack this turn.",
		Colors:     []string{"W"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"MaxTargets: 2,",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Object: opt.Val(game.TargetPermanentReference(1)),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Count(source, "game.RuleEffectCantAttack") != 2 {
		t.Fatalf("expected two can't-attack applications:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceCantAttackFamilyFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording deviates from an exact temporary single-target combat
	// requirement or restriction, so generation must fail closed: it must emit at
	// least one diagnostic and never lower a combat rule effect for it.
	rejected := []string{
		"Creatures can't attack.",
		"Target creature can't attack.",
		"Target creature attacks target opponent this turn if able.",
		"Target creature can't attack you this turn.",
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
		for _, leaked := range []string{
			"game.RuleEffectCantAttack",
			"game.RuleEffectMustAttack",
		} {
			if strings.Contains(source, leaked) {
				t.Errorf("GenerateExecutableCardSource(%q) lowered %s, want fail closed:\n%s", oracle, leaked, source)
			}
		}
	}
}
