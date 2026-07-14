package parser

import (
	"reflect"
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

func TestParseDamageBySourceCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		condition         string
		colors            []TriggerColor
		requiredTypes     []TriggerCardType
		excludeSource     bool
		recipientOpponent bool
		noncombatOnly     bool
		anyController     bool
	}{
		{
			name:      "any source you control to any recipient",
			condition: "a source you control would deal damage to a permanent or player",
		},
		{
			name:          "another red source excludes self",
			condition:     "another red source you control would deal damage to a permanent or player",
			colors:        []TriggerColor{TriggerColorRed},
			excludeSource: true,
		},
		{
			name:              "red source to opponent recipient",
			condition:         "a red source you control would deal damage to an opponent or a permanent an opponent controls",
			colors:            []TriggerColor{TriggerColorRed},
			recipientOpponent: true,
		},
		{
			name:              "noncombat to opponent recipient",
			condition:         "a source you control would deal noncombat damage to an opponent or a permanent an opponent controls",
			recipientOpponent: true,
			noncombatOnly:     true,
		},
		{
			name:          "creature source",
			condition:     "a creature you control would deal damage to a permanent or player",
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
		},
		{
			name:          "any controller source",
			condition:     "a source would deal damage to a permanent or player",
			anyController: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != ConditionPredicateDamageByControlledSource {
				t.Fatalf("predicate = %v, want DamageByControlledSource", clause.Predicate)
			}
			selection := clause.Selection
			if !slices.Equal(selection.ColorsAny, test.colors) ||
				!slices.Equal(selection.RequiredTypes, test.requiredTypes) ||
				selection.ExcludeSource != test.excludeSource ||
				selection.DamageRecipientOpponent != test.recipientOpponent ||
				selection.DamageNoncombatOnly != test.noncombatOnly ||
				selection.DamageSourceAnyController != test.anyController {
				t.Fatalf("selection = %#v", selection)
			}
		})
	}
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
		{"controller life at least", "you have at least 7 life", ConditionPredicateControllerLifeAtLeast, 7},
		{"controller life at most", "you have 5 or less life", ConditionPredicateControllerLifeAtMost, 5},
		{"controller life at most zero", "you have 0 or less life", ConditionPredicateControllerLifeAtMost, 0},
		{"controller life above starting", "you have at least 10 life more than your starting life total", ConditionPredicateControllerLifeAtLeastAboveStarting, 10},
		{"controller hand size", "you have one or more cards in hand", ConditionPredicateControllerHandSizeAtLeast, 1},
		{"controller hand empty", "you have no cards in hand", ConditionPredicateControllerHandEmpty, 0},
		{"any player life at most", "a player has 5 or less life", ConditionPredicateAnyPlayerLifeAtMost, 5},
		{"opponent count", "you have two or more opponents", ConditionPredicateOpponentCountAtLeast, 2},
		{"graveyard cards", "there are six or more cards in your graveyard", ConditionPredicateGraveyardCardCountAtLeast, 6},
		{"graveyard card types", "there are three or more card types among cards in your graveyard", ConditionPredicateGraveyardCardTypeCountAtLeast, 3},
		{"graveyard creature cards", "twenty or more creature cards are in your graveyard", ConditionPredicateGraveyardCardOfTypeCountAtLeast, 20},
		{"graveyard land cards", "seven or more land cards are in your graveyard", ConditionPredicateGraveyardCardOfTypeCountAtLeast, 7},
		{"creature power diversity", "you control three or more creatures with different powers", ConditionPredicateCreaturePowerDiversityAtLeast, 3},
		{"opponent poison counters", "an opponent has three or more poison counters", ConditionPredicateAnyOpponentPoisonAtLeast, 3},
		{"controller hand size exactly", "you have exactly seven cards in hand", ConditionPredicateControllerHandSizeExactly, 7},
		{"controller hand size exactly your hand", "you have exactly thirteen cards in your hand", ConditionPredicateControllerHandSizeExactly, 13},
		{"controller library size", "you have 200 or more cards in your library", ConditionPredicateControllerLibrarySizeAtLeast, 200},
		{"controller life exactly", "you have exactly 1 life", ConditionPredicateControllerLifeExactly, 1},
		{"controls twenty creatures", "you control twenty or more creatures", ConditionPredicateControls, 0},
		{"created token this turn", "you created a token this turn", ConditionPredicateCreatedTokenThisTurn, 0},
		{"active token creation", "an effect would create one or more tokens under your control", ConditionPredicateTokenCreationUnderController, 0},
		{"passive token creation", "one or more tokens would be created under your control", ConditionPredicateTokenCreationUnderController, 0},
		{"any-player active token creation", "an effect would create one or more tokens", ConditionPredicateTokenCreationAnyController, 0},
		{"any-player passive token creation", "one or more tokens would be created", ConditionPredicateTokenCreationAnyController, 0},
		{"typed controller token creation", "you would create one or more Treasure tokens", ConditionPredicateTokenCreationUnderController, 0},
		{"cast during main phase", "you cast this spell during your main phase", ConditionPredicateCastDuringControllerMainPhase, 0},
		{"spell was kicked", "this spell was kicked", ConditionPredicateSpellWasKicked, 0},
		{"spell was bargained", "this spell was bargained", ConditionPredicateSpellWasBargained, 0},
		{"spell is bargained", "it's bargained", ConditionPredicateSpellWasBargained, 0},
		{"controls commander", "you control your commander", ConditionPredicateControllerControlsCommander, 0},
		{"spell X at least more", "X is 10 or more", ConditionPredicateSpellXAtLeast, 10},
		{"spell X at least greater", "X is 4 or greater", ConditionPredicateSpellXAtLeast, 4},
		{"controller is monarch", "you're the monarch", ConditionPredicateControllerIsMonarch, 0},
		{"controller is monarch long form", "you are the monarch", ConditionPredicateControllerIsMonarch, 0},
		{"an opponent is monarch", "an opponent is the monarch", ConditionPredicateAnOpponentIsMonarch, 0},
		{"controller was monarch at turn start", "you were the monarch as the turn began", ConditionPredicateControllerWasMonarchAtTurnStart, 0},
		{"defending player is monarch", "defending player is the monarch", ConditionPredicateDefendingPlayerIsMonarch, 0},
		{"that player is monarch", "that player is the monarch", ConditionPredicateThatPlayerIsMonarch, 0},
		{"controller has initiative", "you have the initiative", ConditionPredicateControllerHasInitiative, 0},
		{"controller has city's blessing", "you have the city's blessing", ConditionPredicateControllerHasCityBlessing, 0},
		{"event spell no mana spent", "no mana was spent to cast it", ConditionPredicateEventSpellNoManaSpentToCast, 0},
		{"event spell mana spent at least", "at least four mana was spent to cast it", ConditionPredicateEventSpellManaSpentToCastAtLeast, 4},
		{"event spell mana spent that spell", "at least eight mana was spent to cast that spell", ConditionPredicateEventSpellManaSpentToCastAtLeast, 8},
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

func TestParseConditionGraveyardCardOfTypeCount(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t,
		"twenty or more creature cards are in your graveyard")
	if clause.Predicate != ConditionPredicateGraveyardCardOfTypeCountAtLeast {
		t.Fatalf("predicate = %s, want graveyard-card-of-type count", clause.Predicate)
	}
	if clause.Threshold != 20 {
		t.Fatalf("threshold = %d, want 20", clause.Threshold)
	}
	if clause.GraveyardCountCardType != TriggerCardTypeCreature {
		t.Fatalf("card type = %s, want creature", clause.GraveyardCountCardType)
	}
}

func TestParseCounterPlacementControlledTypeUnion(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t,
		"one or more +1/+1 counters would be put on an artifact or creature you control")
	if clause.Predicate != ConditionPredicateCounterPlacementOnControlledPermanent {
		t.Fatalf("predicate = %s, want controlled-permanent counter placement", clause.Predicate)
	}
	if clause.Counter != ConditionCounterPlusOnePlusOne {
		t.Fatalf("counter = %s, want +1/+1", clause.Counter)
	}
	want := []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature}
	if !slices.Equal(clause.CounterRecipientTypesAny, want) {
		t.Fatalf("recipient types = %v, want %v", clause.CounterRecipientTypesAny, want)
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
		{
			name:          "multi subtype disjunction",
			condition:     "you control another Wolf or Werewolf",
			comparison:    ConditionComparisonNone,
			subtypes:      []types.Sub{types.Wolf, types.Werewolf},
			excludeSource: true,
		},
		{
			name:          "defending player scope",
			condition:     "defending player controls a creature",
			scope:         ConditionControlScopeDefendingPlayer,
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

// TestParseEventSubjectNonTokenCondition covers the negated intervening-if "if
// it's not a token" (Life of the Party) and its equivalents. Each must produce
// the event-permanent object-matches predicate with a NonToken selection (the
// negation of TokenOnly), so the gate tests that the entering permanent is not a
// token.
func TestParseEventSubjectNonTokenCondition(t *testing.T) {
	t.Parallel()
	conditions := []string{
		"it's not a token",
		"it isn't a token",
		"it is not a token",
		"it wasn't a token",
		"it was not a token",
	}
	for _, condition := range conditions {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, condition)
			if clause.Predicate != ConditionPredicateObjectMatches ||
				clause.ObjectBinding != ConditionObjectBindingEventPermanent ||
				!clause.Selection.NonToken ||
				clause.Selection.TokenOnly {
				t.Fatalf("clause = %#v", clause)
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
		{"event was bargained", "it was bargained", ConditionPredicateEventSubjectWasBargained, ConditionObjectBindingNone, nil, ConditionCombatAny, 0},
		{"source named was bargained", "this creature was bargained", ConditionPredicateEventSubjectWasBargained, ConditionObjectBindingNone, nil, ConditionCombatAny, 0},
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

// TestParseEventSubjectHadCounterCondition covers the affirmative counter-presence
// intervening-if on the dying creature's last-known information, the mirror of the
// Undying/Persist negative "it had no <kind> counters" gate. Both the singular
// "it had a <kind> counter on it" and the equivalent "it had one or more <kind>
// counters on it" spellings test the presence of at least one counter of that
// kind, binding the event permanent.
func TestParseEventSubjectHadCounterCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		predicate ConditionPredicateKind
		counter   ConditionCounterKind
	}{
		{"had a plus counter", "it had a +1/+1 counter on it", ConditionPredicateEventSubjectHadCounter, ConditionCounterPlusOnePlusOne},
		{"had a minus counter", "it had a -1/-1 counter on it", ConditionPredicateEventSubjectHadCounter, ConditionCounterMinusOneMinusOne},
		{"had one or more plus counters", "it had one or more +1/+1 counters on it", ConditionPredicateEventSubjectHadCounter, ConditionCounterPlusOnePlusOne},
		{"had a plus counter no suffix", "it had a +1/+1 counter", ConditionPredicateEventSubjectHadCounter, ConditionCounterPlusOnePlusOne},
		{"had no plus counters", "it had no +1/+1 counters on it", ConditionPredicateEventSubjectHadNoCounter, ConditionCounterPlusOnePlusOne},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != test.predicate ||
				clause.ObjectBinding != ConditionObjectBindingNone ||
				clause.Counter != test.counter {
				t.Fatalf("clause = %#v", clause)
			}
		})
	}
}

// TestParseEventSubjectPastTensePowerCondition covers the past-tense power
// threshold "its power was N or greater" used by dies triggers (Deathknell
// Berserker). It must produce the same event-permanent power-at-least predicate
// as the present-tense "its power is N or greater" enter form, because the dying
// creature's power is read from its last-known information.
func TestParseEventSubjectPastTensePowerCondition(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"its power is 3 or greater",
		"its power was 3 or greater",
	} {
		clause := parseSingleConditionClause(t, condition)
		if clause.Predicate != ConditionPredicateObjectMatches ||
			clause.ObjectBinding != ConditionObjectBindingEventPermanent ||
			!clause.Selection.MatchPowerAtLeast ||
			clause.Selection.PowerAtLeast != 3 {
			t.Fatalf("condition %q clause = %#v", condition, clause)
		}
	}
}

// TestParseSelfNamePossessivePowerCondition covers the source-power activation
// condition written with a possessive self-name subject ("Kitsa's power is 3 or
// greater") rather than "this creature's power ...". The possessive self-name
// binds the source permanent, so it must produce the same source power-threshold
// predicate as the "this creature" spelling.
func TestParseSelfNamePossessivePowerCondition(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"{2}, {T}: Draw a card. Activate only if Kitsa's power is 3 or greater.",
		Context{CardName: "Kitsa, Otterball Elite"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	clauses := document.Abilities[0].ConditionClauses
	if len(clauses) != 1 {
		t.Fatalf("clauses = %#v, want exactly one", clauses)
	}
	clause := clauses[0]
	if clause.Predicate != ConditionPredicateObjectMatches ||
		clause.ObjectBinding != ConditionObjectBindingSource ||
		!clause.Selection.MatchPowerAtLeast ||
		clause.Selection.PowerAtLeast != 3 {
		t.Fatalf("clause = %#v", clause)
	}
}

// TestParseAttachedCreatureStateCondition covers the conditional-grant gate
// "equipped/enchanted creature is <state>" used by Equipment and Auras
// ("As long as equipped creature is legendary, it has hexproof."). The subject
// binds the attached object and a bare supertype state sets the supertype
// filter.
func TestParseAttachedCreatureStateCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		selection ConditionSelection
	}{
		{"equipped legendary", "equipped creature is legendary", ConditionSelection{Supertypes: []ConditionSupertype{ConditionSupertypeLegendary}}},
		{"enchanted legendary", "enchanted creature is legendary", ConditionSelection{Supertypes: []ConditionSupertype{ConditionSupertypeLegendary}}},
		{"enchanted permanent is creature", "enchanted permanent is a creature", ConditionSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}}},
		{"enchanted land is creature", "enchanted land is a creature", ConditionSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}}},
		{"equipped subtype", "equipped creature is a Human", ConditionSelection{SubtypesAny: []types.Sub{types.Sub("Human")}}},
		{"equipped subtype disjunction", "equipped creature is a Human or an Angel", ConditionSelection{SubtypesAny: []types.Sub{types.Sub("Human"), types.Sub("Angel")}}},
		{"enchanted color red", "enchanted creature is red", ConditionSelection{ColorsAny: []TriggerColor{TriggerColorRed}}},
		{"enchanted color blue", "enchanted creature is blue", ConditionSelection{ColorsAny: []TriggerColor{TriggerColorBlue}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != ConditionPredicateObjectMatches ||
				clause.ObjectBinding != ConditionObjectBindingSourceAttached ||
				!reflect.DeepEqual(clause.Selection, test.selection) {
				t.Fatalf("clause = %#v", clause)
			}
		})
	}
}

// TestParseEnteredOrCastFromGraveyardCondition covers the enters-the-battlefield
// intervening condition that gates on the entering object(s) having come from a
// graveyard, in both the singular self form and the plural group form, and
// confirms unrelated zone wording fails closed.
func TestParseEnteredOrCastFromGraveyardCondition(t *testing.T) {
	t.Parallel()
	recognized := []struct {
		name      string
		condition string
		predicate ConditionPredicateKind
	}{
		{"controller full", "it entered from your graveyard or you cast it from your graveyard", ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard},
		{"controller plural", "they entered from your graveyard or you cast them from your graveyard", ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard},
		{"any singular", "it entered or was cast from a graveyard", ConditionPredicateEventSubjectEnteredOrCastFromGraveyard},
		{"any plural", "they entered or were cast from a graveyard", ConditionPredicateEventSubjectEnteredOrCastFromGraveyard},
	}
	for _, test := range recognized {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != test.predicate {
				t.Fatalf("clause = %#v, want %s", clause, test.predicate)
			}
		})
	}
	rejected := []struct {
		name      string
		condition string
	}{
		{"from exile", "it entered from exile"},
		{"from hand", "you cast it from your hand"},
		{"entered tapped", "it entered tapped"},
	}
	for _, test := range rejected {
		t.Run("reject_"+test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(
				"When this creature enters, if "+test.condition+", draw a card.",
				Context{},
			)
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == ConditionPredicateEventSubjectEnteredOrCastFromGraveyard ||
					clause.Predicate == ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard {
					t.Fatalf("condition %q unexpectedly matched a graveyard zone-change predicate", test.condition)
				}
			}
		})
	}
}

// TestParseCastFromControllerHandCondition covers the enters-the-battlefield
// intervening condition that gates on the entering object having been cast by
// the controller from their hand ("if you cast it from your hand"), and
// confirms it stays distinct from the bare "you cast it" controller form.
func TestParseCastFromControllerHandCondition(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t, "you cast it from your hand")
	if clause.Predicate != ConditionPredicateEventSubjectWasCastFromControllerHand {
		t.Fatalf("clause = %#v, want %s", clause, ConditionPredicateEventSubjectWasCastFromControllerHand)
	}
	bare := parseSingleConditionClause(t, "you cast it")
	if bare.Predicate != ConditionPredicateEventSubjectWasCastByController {
		t.Fatalf("bare clause = %#v, want %s", bare, ConditionPredicateEventSubjectWasCastByController)
	}
}

// TestParseConditionPriorInstruction covers the affirmative "you do" and
// negative "you don't" reflexive prior-instruction clauses used by optional
// resolving flow ("you may X. If you do/don't, Y"), plus the affirmative
// non-controller player-subject wordings "if they do", "if the player does",
// and "if that player does" that name the same resolving-success gate for an
// action taken by a non-controller player.
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
		{"if they do", "Target opponent discards a card. If they do, they draw a card.", ConditionPredicatePriorInstructionAccepted},
		{"if the player does", "Target opponent discards a card. If the player does, they draw a card.", ConditionPredicatePriorInstructionAccepted},
		{"if that player does", "Target opponent discards a card. If that player does, they draw a card.", ConditionPredicatePriorInstructionAccepted},
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

// TestParseConditionDestroyedThisWay covers the outcome-worded resolving success
// gate "If a <permanent> is destroyed this way, ..." (Noxious Gearhulk), which
// follows a preceding optional destroy and maps to its own destroyed-this-way
// predicate, distinct from the literal "if you do" gate.
func TestParseConditionDestroyedThisWay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		body      string
		predicate ConditionPredicateKind
	}{
		{
			name:      "a creature is destroyed this way",
			body:      "You may destroy target creature. If a creature is destroyed this way, you gain 2 life.",
			predicate: ConditionPredicateDestroyedThisWay,
		},
		{
			name:      "a permanent is destroyed this way",
			body:      "You may destroy target permanent. If a permanent is destroyed this way, you gain 2 life.",
			predicate: ConditionPredicateDestroyedThisWay,
		},
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

// TestParseConditionAttackersAttackingController covers the Mangara combat
// intervening-if "if two or more of those creatures are attacking you and/or
// planeswalkers you control", which maps to its own attacker-count-by-defender
// predicate carrying the threshold, alongside the typed opponent-attack trigger.
func TestParseConditionAttackersAttackingController(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever an opponent attacks with creatures, if two or more of those creatures are attacking you and/or planeswalkers you control, draw a card.",
		Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.TriggerEvent == nil {
		t.Fatalf("trigger event not typed: %#v", ability.Trigger)
	}
	if ability.Trigger.TriggerEvent.Kind != TriggerEventKindAttack ||
		ability.Trigger.TriggerEvent.Actor.Kind != TriggerEventActorOpponent {
		t.Fatalf("trigger event = %#v, want opponent attack", ability.Trigger.TriggerEvent)
	}
	clauses := ability.ConditionClauses
	if len(clauses) != 1 ||
		clauses[0].Predicate != ConditionPredicateAttackersAttackingControllerAtLeast ||
		clauses[0].Threshold != 2 {
		t.Fatalf("clauses = %#v, want attackers-attacking-controller threshold 2", clauses)
	}
}

// TestParseConditionGainedLifeThisTurn covers the intervening-if condition
// "if you gained N or more life this turn" (Angelic Accord, Griffin Aerie),
// which gates an end-step trigger on the controller's accumulated life gain.
func TestParseConditionGainedLifeThisTurn(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"At the beginning of each end step, if you gained 4 or more life this turn, draw a card.",
		Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	clauses := document.Abilities[0].ConditionClauses
	if len(clauses) != 1 ||
		clauses[0].Predicate != ConditionPredicateControllerGainedLifeThisTurnAtLeast ||
		clauses[0].Threshold != 4 {
		t.Fatalf("clauses = %#v, want gained-life-this-turn threshold 4", clauses)
	}
}

// TestParseConditionDestroyedThisWayRejectsOtherWording confirms the recognizer
// fails closed on wording it does not model, leaving an unsupported condition
// rather than a silently-wrong success gate.
func TestParseConditionDestroyedThisWayRejectsOtherWording(t *testing.T) {
	t.Parallel()
	bodies := []string{
		"You may destroy target creature. If a creature is exiled this way, you gain 2 life.",
		"You may destroy target creature. If a spell is destroyed this way, you gain 2 life.",
	}
	for _, body := range bodies {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(body, Context{InstantOrSorcery: true})
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == ConditionPredicatePriorInstructionAccepted ||
					clause.Predicate == ConditionPredicateDestroyedThisWay {
					t.Fatalf("clause unexpectedly recognized as prior-instruction success: %#v", clause)
				}
			}
		})
	}
}

// TestParseConditionControlsDistinctNames covers the "you control N or more
// <selection> with different names" qualifier (Field of the Dead). The parser
// records the distinct-name threshold on the selection while still emitting the
// controls predicate.
func TestParseConditionControlsDistinctNames(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t, "you control seven or more lands with different names")
	if clause.Predicate != ConditionPredicateControls {
		t.Fatalf("clause = %#v, want controls predicate", clause)
	}
	if clause.Scope != ConditionControlScopeController ||
		clause.Comparison != ConditionComparisonAtLeast ||
		clause.CompareValue != 7 {
		t.Fatalf("clause = %#v, want controller scope at-least 7", clause)
	}
	selection := clause.Selection
	if !slices.Equal(selection.RequiredTypes, []TriggerCardType{TriggerCardTypeLand}) {
		t.Fatalf("selection = %#v, want land type", selection)
	}
	if !selection.MatchDistinctNamesAtLeast || selection.DistinctNamesAtLeast != 7 {
		t.Fatalf("selection = %#v, want distinct names 7", selection)
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

// TestParseConditionControlsTotalPowerAtMost covers the upper-bound total-power
// qualifier ("N or less"), the counterpart the total-power condition recognizer
// gained alongside the "N or greater" lower bound. No currently supported card
// uses it, but the recognizer is general so a future card ("If those creatures
// have total power N or less, ...") lowers without a bespoke path.
func TestParseConditionControlsTotalPowerAtMost(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t, "creatures you control have total power 8 or less")
	if clause.Predicate != ConditionPredicateControls {
		t.Fatalf("clause = %#v, want controls predicate", clause)
	}
	if clause.Scope != ConditionControlScopeController {
		t.Fatalf("clause = %#v, want controller scope", clause)
	}
	selection := clause.Selection
	if !slices.Equal(selection.RequiredTypes, []TriggerCardType{TriggerCardTypeCreature}) {
		t.Fatalf("selection = %#v, want creature type", selection)
	}
	if !selection.MatchTotalPowerAtMost || selection.TotalPowerAtMost != 8 {
		t.Fatalf("selection = %#v, want total power at most 8", selection)
	}
	if selection.MatchTotalPowerAtLeast {
		t.Fatalf("selection = %#v, upper bound must not set the lower-bound qualifier", selection)
	}
}

// TestParseConditionThoseCreaturesTotalPower covers the anaphoric subject the
// total-power recognizer resolves to the attacking creatures bound by the
// trigger ("those creatures"), the form Ultra Magnus, Armored Carrier uses to
// gate its convert on the just-granted attackers' combined power.
func TestParseConditionThoseCreaturesTotalPower(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t, "those creatures have total power 8 or greater")
	if clause.Predicate != ConditionPredicateControls {
		t.Fatalf("clause = %#v, want controls predicate", clause)
	}
	selection := clause.Selection
	if selection.CombatState != ConditionCombatAttacking {
		t.Fatalf("selection = %#v, want attacking combat state", selection)
	}
	if !selection.MatchTotalPowerAtLeast || selection.TotalPowerAtLeast != 8 {
		t.Fatalf("selection = %#v, want total power 8", selection)
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
		{
			name:      "that player controls more lands than you",
			condition: "that player controls more lands than you",
			left:      ConditionControlScopeTriggeringPlayer,
			right:     ConditionControlScopeController,
			greater:   true,
			cardType:  TriggerCardTypeLand,
		},
		{
			name:      "you control more lands than that player",
			condition: "you control more lands than that player",
			left:      ConditionControlScopeController,
			right:     ConditionControlScopeTriggeringPlayer,
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
		"that player controls more lands than an opponent",
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
		"there are three or more card types among cards in an opponent's graveyard",
		"you control a creature with phasing",
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
