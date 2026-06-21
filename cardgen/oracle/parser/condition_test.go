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
		{"opponent poison counters", "an opponent has three or more poison counters", ConditionPredicateAnyOpponentPoisonAtLeast, 3},
		{"controller hand size exactly", "you have exactly seven cards in hand", ConditionPredicateControllerHandSizeExactly, 7},
		{"created token this turn", "you created a token this turn", ConditionPredicateCreatedTokenThisTurn, 0},
		{"active token creation", "an effect would create one or more tokens under your control", ConditionPredicateTokenCreationUnderController, 0},
		{"passive token creation", "one or more tokens would be created under your control", ConditionPredicateTokenCreationUnderController, 0},
		{"cast during main phase", "you cast this spell during your main phase", ConditionPredicateCastDuringControllerMainPhase, 0},
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
		multicolored  bool
		tokenOnly     bool
		excludeSource bool
		tapped        ConditionTappedState
		power         int
		keyword       KeywordKind
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
			name:          "keyword filter",
			condition:     "you control a creature with flying",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			keyword:       KeywordFlying,
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
			name:          "legendary creature supertype",
			condition:     "you control a legendary creature",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			supertypes:    []ConditionSupertype{ConditionSupertypeLegendary},
		},
		{
			name:          "legendary color-qualified creature supertype",
			condition:     "you control a legendary green creature",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			supertypes:    []ConditionSupertype{ConditionSupertypeLegendary},
			colors:        []TriggerColor{TriggerColorGreen},
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
		{
			name:         "count color permanents plural",
			condition:    "you control two or more red permanents",
			comparison:   ConditionComparisonAtLeast,
			compareValue: 2,
			colors:       []TriggerColor{TriggerColorRed},
		},
		{
			name:         "count snow permanents plural",
			condition:    "you control four or more snow permanents",
			comparison:   ConditionComparisonAtLeast,
			compareValue: 4,
			supertypes:   []ConditionSupertype{ConditionSupertypeSnow},
		},
		{
			name:          "color creatures plural",
			condition:     "you control no other colorless creatures",
			comparison:    ConditionComparisonAtMost,
			compareValue:  0,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			colorless:     true,
			excludeSource: true,
		},
		{
			name:       "token",
			condition:  "you control a token",
			comparison: ConditionComparisonNone,
			tokenOnly:  true,
		},
		{
			name:          "another multicolored permanent",
			condition:     "you control another multicolored permanent",
			comparison:    ConditionComparisonNone,
			multicolored:  true,
			excludeSource: true,
		},
		{
			name:          "multicolored creature",
			condition:     "you control a multicolored creature",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			multicolored:  true,
		},
		{
			name:          "typed subtype creature",
			condition:     "you control a Griffin creature",
			comparison:    ConditionComparisonNone,
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			subtypes:      []types.Sub{types.Griffin},
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
				selection.Multicolored != test.multicolored ||
				selection.TokenOnly != test.tokenOnly ||
				selection.ExcludeSource != test.excludeSource ||
				selection.Tapped != test.tapped ||
				selection.PowerAtLeast != test.power ||
				selection.Keyword != test.keyword {
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
		combat    ConditionCombatState
		power     int
	}{
		{"event was creature", "it was a creature", ConditionPredicateObjectMatches, ConditionObjectBindingEventPermanent, nil, ConditionCombatAny, 0},
		{"event was human subtype", "it was a Human", ConditionPredicateObjectMatches, ConditionObjectBindingEventPermanent, []types.Sub{types.Human}, ConditionCombatAny, 0},
		{"event was kicked", "it was kicked", ConditionPredicateEventSubjectWasKicked, ConditionObjectBindingNone, nil, ConditionCombatAny, 0},
		{"event was cast", "it was cast", ConditionPredicateEventSubjectWasCast, ConditionObjectBindingNone, nil, ConditionCombatAny, 0},
		{"event had counters", "it had counters on it", ConditionPredicateEventSubjectHadCounters, ConditionObjectBindingEventPermanent, nil, ConditionCombatAny, 0},
		{"event name unique", "it doesn't have the same name as another creature you control or a creature card in your graveyard", ConditionPredicateEventSubjectNameUnique, ConditionObjectBindingEventPermanent, nil, ConditionCombatAny, 0},
		{"source attacking", "this creature is attacking", ConditionPredicateObjectMatches, ConditionObjectBindingSource, nil, ConditionCombatAttacking, 0},
		{"source blocking", "this creature is blocking", ConditionPredicateObjectMatches, ConditionObjectBindingSource, nil, ConditionCombatBlocking, 0},
		{"source attacking or blocking", "this creature is attacking or blocking", ConditionPredicateObjectMatches, ConditionObjectBindingSource, nil, ConditionCombatAttackingOrBlocking, 0},
		{"source power", "this creature's power is 4 or greater", ConditionPredicateObjectMatches, ConditionObjectBindingSource, nil, ConditionCombatAny, 4},
		{"event power", "its power is 3 or greater", ConditionPredicateObjectMatches, ConditionObjectBindingEventPermanent, nil, ConditionCombatAny, 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != test.predicate ||
				clause.ObjectBinding != test.binding ||
				!slices.Equal(clause.Selection.SubtypesAny, test.subtypes) ||
				clause.Selection.CombatState != test.combat ||
				clause.Selection.PowerAtLeast != test.power {
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
		{"when you do", "You may discard a card. When you do, draw a card.", ConditionPredicatePriorInstructionAccepted},
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

func TestParseConditionControlsTotalPower(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t, "creatures you control have total power 8 or greater")
	if clause.Predicate != ConditionPredicateControls {
		t.Fatalf("clause = %#v, want controls predicate", clause)
	}
	if clause.Scope != ConditionControlScopeController ||
		clause.Comparison != ConditionComparisonNone ||
		clause.CompareValue != 0 {
		t.Fatalf("clause = %#v, want controller scope no comparison", clause)
	}
	selection := clause.Selection
	if !slices.Equal(selection.RequiredTypes, []TriggerCardType{TriggerCardTypeCreature}) {
		t.Fatalf("selection = %#v, want creature type", selection)
	}
	if !selection.MatchTotalPowerAtLeast || selection.TotalPowerAtLeast != 8 {
		t.Fatalf("selection = %#v, want total power 8", selection)
	}
	if selection.MatchPowerAtLeast || selection.PowerAtLeast != 0 {
		t.Fatalf("selection = %#v, total-power qualifier must not set per-permanent power", selection)
	}
}

// TestParseConditionControlComparison covers cross-player control-count
// comparison conditions ("an opponent controls more lands than you" and its
// variants). The parser must record which player scope is on each side of the
// comparison and the direction, and fail closed when neither or both sides are
// the controller.
func TestParseConditionControlComparison(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		left      ConditionControlScope
		right     ConditionControlScope
		greater   bool
		cardType  TriggerCardType
	}{
		{
			name:      "opponent controls more lands than you",
			condition: "an opponent controls more lands than you",
			left:      ConditionControlScopeAnyOpponent,
			right:     ConditionControlScopeController,
			greater:   true,
			cardType:  TriggerCardTypeLand,
		},
		{
			name:      "you control fewer lands than an opponent",
			condition: "you control fewer lands than an opponent",
			left:      ConditionControlScopeController,
			right:     ConditionControlScopeAnyOpponent,
			greater:   false,
			cardType:  TriggerCardTypeLand,
		},
		{
			name:      "opponent controls more creatures than you",
			condition: "an opponent controls more creatures than you",
			left:      ConditionControlScopeAnyOpponent,
			right:     ConditionControlScopeController,
			greater:   true,
			cardType:  TriggerCardTypeCreature,
		},
		{
			name:      "you control more lands than each opponent",
			condition: "you control more lands than each opponent",
			left:      ConditionControlScopeController,
			right:     ConditionControlScopeEachOpponent,
			greater:   true,
			cardType:  TriggerCardTypeLand,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != ConditionPredicateControlComparison {
				t.Fatalf("clause = %#v, want control comparison predicate", clause)
			}
			comparison := clause.ControlComparison
			if comparison.LeftScope != test.left ||
				comparison.RightScope != test.right ||
				comparison.Greater != test.greater {
				t.Fatalf("comparison = %#v, want left %s right %s greater %t",
					comparison, test.left, test.right, test.greater)
			}
			if !slices.Equal(clause.Selection.RequiredTypes, []TriggerCardType{test.cardType}) {
				t.Fatalf("selection = %#v, want card type %s", clause.Selection, test.cardType)
			}
		})
	}
}

// TestParseConditionControlComparisonNearMissFailsClosed rejects comparisons
// whose two sides do not contrast the controller against an opponent scope, so
// the comparison would have no well-defined direction.
func TestParseConditionControlComparisonNearMissFailsClosed(t *testing.T) {
	t.Parallel()
	conditions := []string{
		"an opponent controls more lands than each opponent",
		"you control more lands than you",
		"an opponent controls the same number of lands as you",
		"an opponent controls more lands than a player",
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
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == ConditionPredicateControlComparison {
					t.Fatalf("condition %q produced comparison clause %#v, want none", condition, clause)
				}
			}
		})
	}
}

func TestParseConditionTotalPowerNearMissFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording resembles the total-power qualifier but uses an unrecognized
	// scope, comparison polarity, or noun, so the parser must emit no clause.
	conditions := []string{
		"creatures your opponents control have total power 8 or greater",
		"creatures you control have total power 8 or less",
		"creatures you control have total toughness 8 or greater",
		"creatures you control have power 8 or greater",
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

func TestParseConditionGraveyardControls(t *testing.T) {
	t.Parallel()
	// The Incarnation-cycle condition "this card is in your graveyard and you
	// control a <land>" marks graveyard function on the clause while delegating
	// the trailing requirement to the controls recognizer.
	tests := []struct {
		name      string
		condition string
		subtype   types.Sub
	}{
		{"anger", "this card is in your graveyard and you control a Mountain", types.Sub("Mountain")},
		{"wonder", "this card is in your graveyard and you control an Island", types.Sub("Island")},
		{"this creature", "this creature is in your graveyard and you control a Forest", types.Sub("Forest")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != ConditionPredicateControls {
				t.Fatalf("clause = %#v, want controls predicate", clause)
			}
			if !clause.SourceInGraveyard {
				t.Fatalf("clause = %#v, want SourceInGraveyard", clause)
			}
			if !slices.Equal(clause.Selection.SubtypesAny, []types.Sub{test.subtype}) {
				t.Fatalf("selection = %#v, want subtype %s", clause.Selection, test.subtype)
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
		"you control a creature with deathtouch and flying",
		"you control two or fewer creatures with the same power",
		"you have exactly seven cards in your graveyard",
		"there are six or more creature cards in your graveyard",
		"there are three or more card types among cards in an opponent's graveyard",
		"you control a creature with banding",
		"a player has 5 or more life",
		"you gain control of a creature",
		"you control a creature creature",
		"you control a world enchantment",
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
