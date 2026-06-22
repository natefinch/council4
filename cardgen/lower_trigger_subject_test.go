package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/opt"
)

// TestLowerTriggerSubjectBasePower proves a triggered ability whose subject
// carries a "with base power N" qualifier lowers onto the canonical
// Selection.Power filter, the dimension the unified projector inherits from
// CompiledSelector.
func TestLowerTriggerSubjectBasePower(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Augmenter",
		Layout:     "normal",
		TypeLine:   "Creature — Human Artificer",
		OracleText: "Whenever another creature you control with base power 1 enters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	power := face.TriggeredAbilities[0].Trigger.Pattern.SubjectSelection.Power
	if power != opt.Val(compare.Int{Op: compare.Equal, Value: 1}) {
		t.Fatalf("subject power = %#v, want {Equal, 1}", power)
	}
}

// TestLowerTriggerSubjectAnyCounter proves a triggered ability whose subject
// carries a "with a counter on it" qualifier lowers onto Selection.MatchAnyCounter.
func TestLowerTriggerSubjectAnyCounter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Warden",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "Whenever a creature you control with a counter on it deals combat damage to a player, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	selection := face.TriggeredAbilities[0].Trigger.Pattern.DamageSourceSelection
	if !selection.MatchAnyCounter || selection.MatchCounter {
		t.Fatalf("subject selection = %#v, want MatchAnyCounter", selection)
	}
}

// TestLowerTriggerSubjectKindCounterFailsClosed proves a kind-specific
// "with a +1/+1 counter on it" subject qualifier stays unsupported (fails
// closed): the runtime trigger event data exposes no per-kind counter
// information for a subject, so the wording must not silently lower to a
// kind-agnostic counter match.
func TestLowerTriggerSubjectKindCounterFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Counter Warden",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "Whenever another creature you control with a +1/+1 counter on it enters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "testCounterWarden")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource error = %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("diagnostics = 0, want the kind-specific counter qualifier to fail closed")
	}
}

// TestLowerTriggerSubjectUnknownCounterFailsClosed proves the counter subject
// qualifier fails closed on a counter kind the parser does not recognize, so an
// unrepresentable wording stays unsupported rather than silently dropping the
// filter.
func TestLowerTriggerSubjectUnknownCounterFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Bogus Warden",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "Whenever a creature you control with a bogus counter on it deals combat damage to a player, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "testBogusWarden")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource error = %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("diagnostics = 0, want the unrecognized counter kind to fail closed")
	}
}
