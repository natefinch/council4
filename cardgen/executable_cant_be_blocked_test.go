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
		"RequiredTypesAny: []types.Card{types.Creature}",
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
		"ExcludeSource: true",
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

// TestGenerateExecutableCardSourceCantBeBlockedThisTurnSourceBackReference covers
// the "It can't be blocked this turn." back-reference to the source permanent that
// follows a prior sentence in the same effect (Kappa Cannoneer, Sahagin,
// Razzle-Dazzler), lowering to an ApplyRule on the source permanent for the turn.
func TestGenerateExecutableCardSourceCantBeBlockedThisTurnSourceBackReference(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Razzle-Dazzler",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Creature — Human Wizard",
		Colors:     []string{"U"},
		Power:      new("1"),
		Toughness:  new("2"),
		OracleText: "Whenever you cast your second spell each turn, put a +1/+1 counter on this creature. It can't be blocked this turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"TriggeredAbilities:",
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

// TestGenerateExecutableCardSourceCantBeBlockedThisTurnTargetBackReference covers
// the "That creature can't be blocked this turn." back-reference to a creature
// targeted by a prior sentence (Stealth Mission, Assassin Den), lowering to an
// ApplyRule on that same chosen target for the turn.
func TestGenerateExecutableCardSourceCantBeBlockedThisTurnTargetBackReference(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Stealth Mission",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Sorcery",
		Colors:     []string{"U"},
		OracleText: "Put two +1/+1 counters on target creature you control. That creature can't be blocked this turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target creature you control"`,
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

// TestGenerateExecutableCardSourceCantBeBlockedThisCombatEventBackReference covers
// the "it can't be blocked this combat." back-reference to the event permanent of
// an "attacks alone" trigger (Ma Chao, Western Warrior), lowering to an ApplyRule
// on the attacking event permanent for the combat rather than the whole turn.
func TestGenerateExecutableCardSourceCantBeBlockedThisCombatEventBackReference(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Ma Chao, Western Warrior",
		Layout:     "normal",
		ManaCost:   "{3}{R}{R}",
		TypeLine:   "Legendary Creature — Human Soldier Warrior",
		Colors:     []string{"R"},
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: "Horsemanship (This creature can't be blocked except by creatures with horsemanship.)\nWhenever Ma Chao attacks alone, it can't be blocked this combat.",
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
		"Primitive: game.ApplyRule{",
		"Object: opt.Val(game.EventPermanentReference()),",
		"Kind: game.RuleEffectCantBeBlocked,",
		"Duration: game.DurationUntilEndOfCombat,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCantBeBlockedThisTurnBackReferenceFailsClosed
// ensures the back-reference recognizer does not over-match: each wording deviates
// from the exact "<back-reference> can't be blocked this turn/combat." restriction,
// so generation must fail closed with a diagnostic and never lower a can't-be-
// blocked rule effect.
func TestGenerateExecutableCardSourceCantBeBlockedThisTurnBackReferenceFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Put a +1/+1 counter on this creature. It can't be blocked.",
		"Put a +1/+1 counter on this creature. It can't be blocked until end of turn.",
		"Put a +1/+1 counter on this creature. It can't be blocked this turn except by Walls.",
		"Put a +1/+1 counter on target creature you control. That creature can't be blocked this turn unless its controller pays {2}.",
	}
	for _, oracle := range rejected {
		card := &ScryfallCard{
			Name:       "Test Back-Reference Fail Closed",
			Layout:     "normal",
			ManaCost:   "{U}",
			TypeLine:   "Sorcery",
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

func TestGenerateExecutableCardSourceCantBeBlockedThisTurnFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording deviates from the exact "Target creature can't be blocked this
	// turn." restriction, so generation must fail closed: it must emit at least
	// one diagnostic and never lower an ApplyRule / RuleEffectCantBeBlocked.
	rejected := []string{
		"Target creature can't be blocked.",
		"Target creature can't be blocked until end of turn.",
		"Target creature can't be blocked this turn except by Walls.",
		"Target creature can't be blocked this turn unless its controller pays {2}.",
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
