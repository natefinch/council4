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

// TestGenerateExecutableCardSourceCantBeBlockedThisTurnSelf covers a self grant
// "<source> can't be blocked this turn." on an activated ability with a discard
// cost (Ghostly Pilferer's evasion ability), lowering to an ApplyRule on the
// source permanent.
func TestGenerateExecutableCardSourceCantBeBlockedThisTurnSelf(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Pilferer",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Creature — Spirit Rogue",
		Colors:     []string{"U"},
		OracleText: "Discard a card: This creature can't be blocked this turn.",
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
		"Kind: game.RuleEffectCantBeBlocked,",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCantBeBlockedThisTurnUpToOne covers the
// up-to-one target form "Up to one target creature can't be blocked this turn."
// (Key to the City), lowering to a min-zero, max-one permanent target spec.
func TestGenerateExecutableCardSourceCantBeBlockedThisTurnUpToOne(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Key",
		Layout:     "normal",
		ManaCost:   "{2}",
		TypeLine:   "Artifact",
		OracleText: "{T}, Discard a card: Up to one target creature can't be blocked this turn.",
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
		"Kind: cost.AdditionalTap,",
		"Kind:   cost.AdditionalDiscard,",
		"MinTargets: 0,",
		"MaxTargets: 1,",
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

// TestGenerateExecutableCardSourceCantBeBlockedThisTurnSourceAndTarget covers the
// compound "<source> and up to one other target creature can't be blocked this
// turn." subject (Martha Jones), lowering to two ApplyRule instructions: the
// source permanent and the up-to-one chosen target, with the "other" qualifier
// excluding the source from being chosen twice.
func TestGenerateExecutableCardSourceCantBeBlockedThisTurnSourceAndTarget(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Martha Jones",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Legendary Creature — Human Cleric",
		Colors:     []string{"U"},
		Power:      new("2"),
		Toughness:  new("3"),
		OracleText: "Whenever you sacrifice a Clue, Martha Jones and up to one other target creature can't be blocked this turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"TriggeredAbilities:",
		"MinTargets: 0,",
		"MaxTargets: 1,",
		"Another:        true,",
		"Object: opt.Val(game.SourcePermanentReference()),",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Kind: game.RuleEffectCantBeBlocked,",
		"Duration: game.DurationThisTurn,",
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
