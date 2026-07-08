package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
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

// TestLowerTriggerSubjectKindCounter proves a kind-specific "with a +1/+1
// counter on it" subject qualifier lowers onto Selection.MatchCounter with the
// named RequiredCounter kind. The subject is a battlefield permanent for an
// enters event, so the runtime can read its counters to honor the filter.
func TestLowerTriggerSubjectKindCounter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Counter Warden",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "Whenever another creature you control with a +1/+1 counter on it enters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	selection := face.TriggeredAbilities[0].Trigger.Pattern.SubjectSelection
	if !selection.MatchCounter || selection.MatchAnyCounter || selection.RequiredCounter != counter.PlusOnePlusOne {
		t.Fatalf("subject selection = %#v, want MatchCounter with +1/+1 RequiredCounter", selection)
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

// TestLowerTriggerSubjectTypeSubtypeUnion proves a triggered ability whose
// subject unions a card type with a subtype ("creature or Vehicle") lowers onto
// Selection.AnyOf, the sole disjunction the runtime honors, with one alternative
// per member. A conjunctive RequiredTypesAny cannot express a type-or-subtype
// mix, so the union must fan out into AnyOf alternatives.
func TestLowerTriggerSubjectTypeSubtypeUnion(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Union Warden",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "Whenever another creature or Vehicle you control enters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	selection := face.TriggeredAbilities[0].Trigger.Pattern.SubjectSelection
	if len(selection.AnyOf) != 2 {
		t.Fatalf("subject AnyOf = %d alternatives, want 2 (%#v)", len(selection.AnyOf), selection)
	}
	if len(selection.AnyOf[0].RequiredTypesAny) != 1 || selection.AnyOf[0].RequiredTypesAny[0] != types.Creature {
		t.Fatalf("AnyOf[0] = %#v, want RequiredTypesAny=[Creature]", selection.AnyOf[0])
	}
	if len(selection.AnyOf[1].SubtypesAny) != 1 || selection.AnyOf[1].SubtypesAny[0] != types.Vehicle {
		t.Fatalf("AnyOf[1] = %#v, want SubtypesAny=[Vehicle]", selection.AnyOf[1])
	}
	if got := face.TriggeredAbilities[0].Trigger.Pattern.Controller; got != game.TriggerControllerYou {
		t.Fatalf("trigger controller = %v, want TriggerControllerYou", got)
	}
}
