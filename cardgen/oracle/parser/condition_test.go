package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// parseSingleConditionClause parses an intervening-if condition and returns its
// sole typed clause. It fails the test when the wording produced anything other
// than exactly one clause, so meaning tests assert on fully typed syntax rather
// than source text.
func parseSingleConditionClause(t *testing.T, condition string) ConditionClause {
	t.Helper()
	document, diagnostics := Parse(
		"When this creature enters, if "+condition+", draw a card.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	clauses := document.Abilities[0].ConditionClauses
	if len(clauses) != 1 {
		t.Fatalf("condition %q clauses = %#v, want exactly one", condition, clauses)
	}
	return clauses[0]
}

func TestParseConditionPredicateMeaning(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		predicate ConditionPredicateKind
		threshold int
	}{
		{"controller life", "you have 7 or more life", ConditionPredicateControllerLifeAtLeast, 7},
		{"controller hand size", "you have one or more cards in hand", ConditionPredicateControllerHandSizeAtLeast, 1},
		{"controller hand empty", "you have no cards in hand", ConditionPredicateControllerHandEmpty, 0},
		{"any player life at most", "a player has 5 or less life", ConditionPredicateAnyPlayerLifeAtMost, 5},
		{"opponent count", "you have two or more opponents", ConditionPredicateOpponentCountAtLeast, 2},
		{"graveyard cards", "there are six or more cards in your graveyard", ConditionPredicateGraveyardCardCountAtLeast, 6},
		{"graveyard card types", "there are three or more card types among cards in your graveyard", ConditionPredicateGraveyardCardTypeCountAtLeast, 3},
		{"creature power diversity", "you control three or more creatures with different powers", ConditionPredicateCreaturePowerDiversityAtLeast, 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != test.predicate || clause.Threshold != test.threshold {
				t.Fatalf("clause = %#v, want predicate %s threshold %d", clause, test.predicate, test.threshold)
			}
		})
	}
}

func TestParseConditionControlsComposition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		condition     string
		scope         ConditionControlScope
		comparison    ConditionComparison
		compareValue  int
		requiredTypes []TriggerCardType
		supertypes    []ConditionSupertype
		subtypes      []types.Sub
		colors        []TriggerColor
		colorless     bool
		excludeSource bool
		tapped        ConditionTappedState
		power         int
	}{
		{
			name:          "singular creature",
			condition:     "you control a creature",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
		},
		{
			name:          "count artifacts",
			condition:     "you control two or more artifacts",
			comparison:    ConditionComparisonAtLeast,
			compareValue:  2,
			requiredTypes: []TriggerCardType{TriggerCardTypeArtifact},
		},
		{
			name:          "no creatures at most",
			condition:     "you control no creatures",
			comparison:    ConditionComparisonAtMost,
			compareValue:  0,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
		},
		{
			name:          "tapped creature",
			condition:     "you control a tapped creature",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			tapped:        ConditionTappedTrue,
		},
		{
			name:          "power filter",
			condition:     "you control a creature with power 5 or greater",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			power:         5,
		},
		{
			name:       "bare subtype implies no card type",
			condition:  "you control an Equipment",
			comparison: ConditionComparisonNone,
			subtypes:   []types.Sub{types.Equipment},
		},
		{
			name:         "land subtype implies no card type",
			condition:    "you control two or more Gates",
			comparison:   ConditionComparisonAtLeast,
			compareValue: 2,
			subtypes:     []types.Sub{types.Gate},
		},
		{
			name:          "basic land supertype",
			condition:     "you control a basic land",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeLand},
			supertypes:    []ConditionSupertype{ConditionSupertypeBasic},
		},
		{
			name:          "exclude source",
			condition:     "you control another creature",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			excludeSource: true,
		},
		{
			name:          "opponent scope",
			condition:     "an opponent controls a creature",
			scope:         ConditionControlScopeAnyOpponent,
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != ConditionPredicateControls {
				t.Fatalf("clause = %#v, want controls predicate", clause)
			}
			if clause.Scope != test.scope ||
				clause.Comparison != test.comparison ||
				clause.CompareValue != test.compareValue {
				t.Fatalf("clause = %#v, want scope %s comparison %s value %d", clause, test.scope, test.comparison, test.compareValue)
			}
			selection := clause.Selection
			if !slices.Equal(selection.RequiredTypes, test.requiredTypes) ||
				!slices.Equal(selection.Supertypes, test.supertypes) ||
				!slices.Equal(selection.SubtypesAny, test.subtypes) ||
				!slices.Equal(selection.ColorsAny, test.colors) ||
				selection.Colorless != test.colorless ||
				selection.ExcludeSource != test.excludeSource ||
				selection.Tapped != test.tapped ||
				selection.PowerAtLeast != test.power {
				t.Fatalf("selection = %#v", selection)
			}
			if test.power != 0 && !selection.MatchPowerAtLeast {
				t.Fatalf("selection = %#v, want power match", selection)
			}
		})
	}
}

func TestParseConditionEventSubjectAndSourceState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		predicate ConditionPredicateKind
		binding   ConditionObjectBinding
		subtypes  []types.Sub
	}{
		{"event was creature", "it was a creature", ConditionPredicateObjectMatches, ConditionObjectBindingEventPermanent, nil},
		{"event was human subtype", "it was a Human", ConditionPredicateObjectMatches, ConditionObjectBindingEventPermanent, []types.Sub{types.Human}},
		{"event was kicked", "it was kicked", ConditionPredicateEventSubjectWasKicked, ConditionObjectBindingNone, nil},
		{"event was cast", "it was cast", ConditionPredicateEventSubjectWasCast, ConditionObjectBindingNone, nil},
		{"event had counters", "it had counters on it", ConditionPredicateEventSubjectHadCounters, ConditionObjectBindingEventPermanent, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != test.predicate ||
				clause.ObjectBinding != test.binding ||
				!slices.Equal(clause.Selection.SubtypesAny, test.subtypes) {
				t.Fatalf("clause = %#v", clause)
			}
		})
	}
}

// TestParseConditionPriorInstruction covers the affirmative "you do" and
// negative "you don't" reflexive prior-instruction clauses used by optional
// resolving flow ("you may X. If you do/don't, Y").
func TestParseConditionPriorInstruction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		body      string
		predicate ConditionPredicateKind
	}{
		{"if you do", "You may discard a card. If you do, draw a card.", ConditionPredicatePriorInstructionAccepted},
		{"if you don't", "You may discard a card. If you don't, draw a card.", ConditionPredicatePriorInstructionNotAccepted},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.body, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			clauses := document.Abilities[0].ConditionClauses
			if len(clauses) != 1 || clauses[0].Predicate != test.predicate {
				t.Fatalf("clauses = %#v, want predicate %s", clauses, test.predicate)
			}
		})
	}
}

func TestParseConditionNearMissFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording is one normalization away from a supported clause, but uses
	// an unsupported selection filter, comparison, polarity, or noun. The parser
	// must emit no typed clause so the compiler fails the condition closed rather
	// than guessing a meaning.
	conditions := []string{
		"you control a creature with flying",
		"you control two or fewer creatures with the same power",
		"you have exactly three cards in hand",
		"there are six or more creature cards in your graveyard",
		"there are three or more card types among cards in an opponent's graveyard",
		"you control two or more artifacts with flying",
		"a player has 5 or more life",
		"you gain control of a creature",
		"you control a creature creature",
	}
	for _, condition := range conditions {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(
				"When this creature enters, if "+condition+", draw a card.",
				Context{},
			)
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			if clauses := document.Abilities[0].ConditionClauses; len(clauses) != 0 {
				t.Fatalf("condition %q clauses = %#v, want none", condition, clauses)
			}
		})
	}
}
