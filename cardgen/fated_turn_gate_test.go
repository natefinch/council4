package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceFatedIntervention covers the "If it's your
// turn," per-effect sequence gate: the token creation always resolves, and the
// trailing scry is gated on the controller being the active player
// (SourceControllerTurn). This is the Fated cycle's shared bonus rider.
func TestGenerateExecutableCardSourceFatedIntervention(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Fated Intervention",
		Layout:     "normal",
		ManaCost:   "{3}{G}{G}",
		TypeLine:   "Instant",
		OracleText: "Create two 3/3 green Centaur enchantment creature tokens. If it's your turn, scry 2.",
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.CreateToken{",
		"Primitive: game.Scry{",
		"Condition: opt.Val(game.EffectCondition{",
		"SourceControllerTurn: true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	// The scry gate must not be negated; only the always-on token creation is
	// ungated.
	if strings.Contains(source, "Negate:") {
		t.Fatalf("unexpected negated gate in single-branch Fated card:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceStolenVitality covers the "If it's your turn,
// ... Otherwise, ..." two-branch form: the trample branch is gated on
// SourceControllerTurn and the first-strike "Otherwise" branch is gated on the
// negation, so exactly one of the two keyword grants resolves. The first-strike
// branch must bind the target permanent directly rather than chaining a linked
// reference off the mutually-exclusive trample branch (which is skipped whenever
// the first-strike branch fires), so it applies to the correct creature.
func TestGenerateExecutableCardSourceStolenVitality(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Stolen Vitality",
		Layout:   "normal",
		ManaCost: "{1}{R}",
		TypeLine: "Instant",
		OracleText: "Target creature gets +3/+1 until end of turn. If it's your turn, that creature gains trample until end of turn. " +
			"Otherwise, it gains first strike until end of turn.",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Trample,",
		"game.FirstStrike,",
		"SourceControllerTurn: true,",
		"Negate:               true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	// The negated "Otherwise," first-strike branch must not chain off the
	// mutually-exclusive "your turn" trample branch's linked key: that publisher
	// is skipped exactly when this branch fires, so the reference would resolve
	// to nothing. Only the ungated leading ModifyPT publishes a linked key.
	if strings.Count(source, "PublishLinked:") != 1 {
		t.Fatalf("expected exactly one linked publisher (the ungated ModifyPT):\n%s", source)
	}
	if strings.Contains(source, `LinkedObjectReference("gain-keyword-2")`) {
		t.Fatalf("first-strike branch wrongly chains off the mutually-exclusive trample branch:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceFaithfulPikemaster covers the static "As long
// as it's your turn," gate, which shares the controller-turn wording with the
// per-effect Fated gate: the granted first strike applies only while the
// controller is the active player.
func TestGenerateExecutableCardSourceFaithfulPikemaster(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Faithful Pikemaster",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "When this creature enters, scry 2.\nAs long as it's your turn, this creature has first strike.",
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.FirstStrike",
		"SourceControllerTurn: true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
