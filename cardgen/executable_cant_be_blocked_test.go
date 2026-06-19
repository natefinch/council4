package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceCantBeBlockedThisTurnInstant(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Slip Through",
		Layout:     "normal",
		ManaCost:   "{U}",
		TypeLine:   "Instant",
		OracleText: "Target creature can't be blocked this turn.",
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
		`Constraint: "target creature"`,
		"PermanentTypes: []types.Card{types.Creature}",
		"Primitive: game.ApplyRule{",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Kind: game.RuleEffectCantBeBlocked,",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCantBeBlockedThisTurnActivatedAbility(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Rogue's Passage",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{4}, {T}: Target creature can't be blocked this turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"ActivatedAbilities:",
		"AdditionalCosts: cost.Tap,",
		`Constraint: "target creature"`,
		"Primitive: game.ApplyRule{",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Kind: game.RuleEffectCantBeBlocked,",
		"Duration: game.DurationThisTurn,",
		"game.TapManaAbility(mana.C)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCantBeBlockedThisTurnFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording deviates from the exact "Target creature can't be blocked this
	// turn." restriction, so generation must fail closed: it must emit at least
	// one diagnostic and never lower an ApplyRule / RuleEffectCantBeBlocked.
	rejected := []string{
		"Target creature can't be blocked.",
		"Target creature can't be blocked until end of turn.",
		"Target creature can't be blocked this turn except by Walls.",
		"Up to two target creatures can't be blocked this turn.",
		"Target creature can't block this turn.",
		"Target creature can't attack this turn.",
		"Target creature can't be blocked this turn if it's tapped.",
	}
	for _, oracle := range rejected {
		card := &ScryfallCard{
			Name:       "Test Fail Closed",
			Layout:     "normal",
			ManaCost:   "{U}",
			TypeLine:   "Instant",
			OracleText: oracle,
			Colors:     []string{"U"},
		}
		source, diagnostics, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatalf("GenerateExecutableCardSource(%q) err = %v", oracle, err)
		}
		if len(diagnostics) == 0 {
			t.Errorf("GenerateExecutableCardSource(%q) produced no diagnostics, want fail closed", oracle)
		}
		if strings.Contains(source, "game.RuleEffectCantBeBlocked") {
			t.Errorf("GenerateExecutableCardSource(%q) lowered a can't-be-blocked rule effect, want fail closed:\n%s", oracle, source)
		}
	}
}
