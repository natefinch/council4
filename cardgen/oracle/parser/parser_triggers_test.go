package parser

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestParsePhaseStepTriggerClauses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		event      string
		quantifier PhaseStepQuantifierKind
		player     TriggerPlayerSelectorKind
		phaseStep  PhaseStepNameKind
		attached   TriggerSelection
		next       bool
	}{
		{"standalone end of combat", "end of combat", PhaseStepQuantifierNone, TriggerPlayerSelectorAny, PhaseStepNameEndOfCombat, TriggerSelection{}, false},
		{"source controller upkeep", "the beginning of its controller's upkeep", PhaseStepQuantifierSingle, TriggerPlayerSelectorSourceController, PhaseStepNameUpkeep, TriggerSelection{}, false},
		{"your draw step", "the beginning of your draw step", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameDrawStep, TriggerSelection{}, false},
		{"your next upkeep", "the beginning of your next upkeep", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameUpkeep, TriggerSelection{}, true},
		{"each end step", "the beginning of each end step", PhaseStepQuantifierEach, TriggerPlayerSelectorAny, PhaseStepNameEndStep, TriggerSelection{}, false},
		{"each player upkeep", "the beginning of each player's upkeep", PhaseStepQuantifierEach, TriggerPlayerSelectorAny, PhaseStepNameUpkeep, TriggerSelection{}, false},
		{"each opponent draw step", "the beginning of each opponent's draw step", PhaseStepQuantifierEach, TriggerPlayerSelectorOpponent, PhaseStepNameDrawStep, TriggerSelection{}, false},
		{"combat on your turn", "the beginning of combat on your turn", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameCombat, TriggerSelection{}, false},
		{"combat on each turn", "the beginning of combat on each turn", PhaseStepQuantifierEach, TriggerPlayerSelectorAny, PhaseStepNameCombat, TriggerSelection{}, false},
		{"end combat on your turn", "the beginning of the end of combat on your turn", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameEndOfCombat, TriggerSelection{}, false},
		{"each end combat step", "the beginning of each end of combat step", PhaseStepQuantifierEach, TriggerPlayerSelectorAny, PhaseStepNameEndOfCombatStep, TriggerSelection{}, false},
		{"each of your first main phases", "the beginning of each of your first main phases", PhaseStepQuantifierEachOf, TriggerPlayerSelectorYou, PhaseStepNameFirstMainPhase, TriggerSelection{}, false},
		{"your second main phase", "the beginning of your second main phase", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameSecondMainPhase, TriggerSelection{}, false},
		{"your combat step", "the beginning of your combat step", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameCombatStep, TriggerSelection{}, false},
		{"attached controller", "the beginning of the upkeep of enchanted legendary white artifact creature's controller", PhaseStepQuantifierSingle, TriggerPlayerSelectorAttachedController, PhaseStepNameUpkeep, TriggerSelection{
			RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature},
			Supertypes:    []TriggerSupertype{TriggerSupertypeLegendary},
			ColorsAny:     []TriggerColor{TriggerColorWhite},
		}, false},
		{"attached union controller", "the beginning of the upkeep of enchanted artifact and/or creature's controller", PhaseStepQuantifierSingle, TriggerPlayerSelectorAttachedController, PhaseStepNameUpkeep, TriggerSelection{
			RequiredTypesAny: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature},
		}, false},
		{"attached constrained controller", "the beginning of the upkeep of enchanted permanent you control's controller", PhaseStepQuantifierSingle, TriggerPlayerSelectorAttachedController, PhaseStepNameUpkeep, TriggerSelection{
			Controller: ControllerYou,
		}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "At " + test.event + ", draw a card."
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.PhaseStep == nil {
				t.Fatalf("trigger = %#v, want typed phase/step clause", trigger)
			}
			if trigger.Introduction.Kind != TriggerIntroductionAt ||
				trigger.PhaseStep.Quantifier.Kind != test.quantifier ||
				trigger.PhaseStep.Player.Kind != test.player ||
				trigger.PhaseStep.Name.Kind != test.phaseStep ||
				trigger.PhaseStep.Next != test.next ||
				!reflect.DeepEqual(trigger.PhaseStep.Player.AttachedSubject.Selection, test.attached) {
				t.Fatalf("trigger = %#v", trigger)
			}
			assertTextSpan(t, "trigger clause", source, trigger.Span, trigger.Text)
			assertTextSpan(t, "trigger event", source, trigger.EventSpan, trigger.Event)
			assertSpanContains(t, "phase/step clause", trigger.EventSpan, trigger.PhaseStep.Span)
			assertSpanContains(t, "phase/step name", trigger.PhaseStep.Span, trigger.PhaseStep.Name.Span)
		})
	}
}

func TestParsePhaseStepTriggerClausesComposePreviouslyUnsupportedSlots(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"At the beginning of combat on each player's turn, draw a card.",
		"At the beginning of each precombat main phase, draw a card.",
		"At the beginning of your end of combat step, draw a card.",
		"At the beginning of each of your upkeeps, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if trigger := document.Abilities[0].Trigger; trigger == nil || trigger.PhaseStep == nil {
				t.Fatalf("trigger = %#v, want composed phase/step grammar", trigger)
			}
		})
	}
}

func TestParseEveryPreviouslySupportedPhaseStepTriggerClause(t *testing.T) {
	t.Parallel()
	var events []string
	for _, relation := range []string{
		"your",
		"its controller's",
		"the",
		"each",
		"each player's",
		"each opponent's",
	} {
		for _, name := range []string{"upkeep", "draw step", "end step"} {
			events = append(events, "the beginning of "+relation+" "+name)
		}
	}
	events = append(events,
		"end of combat",
		"the end of combat",
		"the beginning of combat on your turn",
		"the beginning of combat on each turn",
		"the beginning of combat on each opponent's turn",
		"the beginning of each combat",
		"the beginning of the end of combat",
		"the beginning of the end of combat on your turn",
		"the beginning of each end of combat step",
		"the beginning of your precombat main phase",
		"the beginning of each player's precombat main phase",
		"the beginning of each opponent's precombat main phase",
		"the beginning of your postcombat main phase",
		"the beginning of each player's postcombat main phase",
		"the beginning of each opponent's postcombat main phase",
		"the beginning of your first main phase",
		"the beginning of each of your first main phases",
		"the beginning of each player's first main phase",
		"the beginning of each opponent's first main phase",
		"the beginning of your second main phase",
		"the beginning of each player's second main phase",
		"the beginning of each opponent's second main phase",
		"the beginning of each of your postcombat main phases",
		"the beginning of your combat step",
		"the beginning of the upkeep of enchanted creature's controller",
		"the beginning of the draw step of enchanted creature's controller",
		"the beginning of the end step of enchanted creature's controller",
		"the beginning of the upkeep of enchanted permanent's controller",
	)
	for _, event := range events {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse("At "+event+", draw a card.", Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if trigger := document.Abilities[0].Trigger; trigger == nil || trigger.PhaseStep == nil {
				t.Fatalf("trigger = %#v, want typed phase/step clause", trigger)
			}
		})
	}
}

func TestParseSimpleStaticRulesAsComposedSyntax(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source         string
		subject        StaticRuleSubjectKind
		subjectText    string
		constraint     StaticRuleConstraintKind
		constraintText string
		operation      StaticRuleOperationKind
		voice          StaticRuleVoice
		operationText  string
		qualifiers     []StaticRuleQualifierKind
	}{
		"cannot block contraction": {
			source:         "This creature can't block.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "can't",
			operation:      StaticRuleOperationBlock,
			voice:          StaticRuleVoiceActive,
			operationText:  "block",
		},
		"cannot block": {
			source:         "This creature cannot block.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "cannot",
			operation:      StaticRuleOperationBlock,
			voice:          StaticRuleVoiceActive,
			operationText:  "block",
		},
		"cannot be blocked contraction": {
			source:         "This creature can't be blocked.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "can't",
			operation:      StaticRuleOperationBlock,
			voice:          StaticRuleVoicePassive,
			operationText:  "be blocked",
		},
		"cannot be blocked": {
			source:         "This creature cannot be blocked.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "cannot",
			operation:      StaticRuleOperationBlock,
			voice:          StaticRuleVoicePassive,
			operationText:  "be blocked",
		},
		"implicit attack requirement": {
			source:         "This creature attacks each combat if able.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintRequirement,
			constraintText: "attacks each combat if able",
			operation:      StaticRuleOperationAttack,
			voice:          StaticRuleVoiceActive,
			operationText:  "attacks",
			qualifiers:     []StaticRuleQualifierKind{StaticRuleQualifierEachCombat, StaticRuleQualifierIfAble},
		},
		"explicit attack requirement": {
			source:         "This creature must attack each combat if able.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintRequirement,
			constraintText: "must",
			operation:      StaticRuleOperationAttack,
			voice:          StaticRuleVoiceActive,
			operationText:  "attack",
			qualifiers:     []StaticRuleQualifierKind{StaticRuleQualifierEachCombat, StaticRuleQualifierIfAble},
		},
		"cannot be countered contraction": {
			source:         "This spell can't be countered.",
			subject:        StaticRuleSubjectSourceSpell,
			subjectText:    "This spell",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "can't",
			operation:      StaticRuleOperationCounter,
			voice:          StaticRuleVoicePassive,
			operationText:  "be countered",
		},
		"cannot be countered": {
			source:         "This spell cannot be countered.",
			subject:        StaticRuleSubjectSourceSpell,
			subjectText:    "This spell",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "cannot",
			operation:      StaticRuleOperationCounter,
			voice:          StaticRuleVoicePassive,
			operationText:  "be countered",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := document.Abilities[0]
			if ability.Kind != AbilityStatic || len(ability.Sentences) != 1 || ability.Sentences[0].StaticRule == nil {
				t.Fatalf("ability = %#v, want one typed static rule", ability)
			}
			rule := ability.Sentences[0].StaticRule
			if rule.Subject.Kind != test.subject ||
				rule.Constraint.Kind != test.constraint ||
				rule.Operation.Kind != test.operation ||
				rule.Operation.Voice != test.voice {
				t.Fatalf("static rule = %#v", rule)
			}
			assertTextSpan(t, "rule", test.source, rule.Span, test.source)
			assertTextSpan(t, "subject", test.source, rule.Subject.Span, test.subjectText)
			assertTextSpan(t, "constraint", test.source, rule.Constraint.Span, test.constraintText)
			assertTextSpan(t, "operation", test.source, rule.Operation.Span, test.operationText)
			if len(rule.Qualifiers) != len(test.qualifiers) {
				t.Fatalf("qualifiers = %#v, want %v", rule.Qualifiers, test.qualifiers)
			}
			for i, qualifier := range rule.Qualifiers {
				if qualifier.Kind != test.qualifiers[i] {
					t.Fatalf("qualifier %d = %#v, want %v", i, qualifier, test.qualifiers[i])
				}
			}
			if len(rule.Qualifiers) == 2 {
				assertTextSpan(t, "each-combat qualifier", test.source, rule.Qualifiers[0].Span, "each combat")
				assertTextSpan(t, "if-able qualifier", test.source, rule.Qualifiers[1].Span, "if able")
			}
		})
	}
}

func TestParsePhaseStepTriggerClausesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"At the beginning of a player's upkeep, draw a card.",
		"At the beginning of your declare attackers step, draw a card.",
		"At the beginning of each of your upkeep, draw a card.",
		"At the beginning of each of your main phases, draw a card.",
		"At the beginning of combat on the turn, draw a card.",
		"At the beginning of end of combat on your turn, draw a card.",
		"Whenever the beginning of your upkeep, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.Event == "" || trigger.EventSpan == (shared.Span{}) {
				t.Fatalf("trigger = %#v, want source-spanned unrecognized clause", trigger)
			}
			if trigger.PhaseStep != nil {
				t.Fatalf("trigger = %#v, want unrecognized phase/step grammar", trigger)
			}
		})
	}
}

func TestParseSimpleStaticRulesFailClosedOnNearMisses(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"This creature can block.",
		"This creature can't attacks.",
		"This creature can't be countered.",
		"This creature attacks each combat.",
		"This creature must be blocked.",
		"This creature must be blocked each combat if able.",
		"This creature attack each combat if able.",
		"This creature attacks each turn if able.",
		"This creature must attack if able.",
		"This creature must attacks each combat if able.",
		"This spell can't block.",
		"This spell can't be blocked.",
		"This spell can't be countered by spells.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, sentence := range document.Abilities[0].Sentences {
				if sentence.StaticRule != nil {
					t.Fatalf("%q parsed as %#v", source, sentence.StaticRule)
				}
			}
		})
	}
}

func TestParsePlayerEventTriggerClauses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		event      string
		player     TriggerPlayerSelectorKind
		action     PlayerEventActionKind
		card       PlayerEventCardKind
		occurrence PlayerEventOccurrenceKind
		ordinal    int
	}{
		{"Whenever you draw a card", TriggerPlayerSelectorYou, PlayerEventActionDraw, PlayerEventCardSingle, PlayerEventOccurrenceAny, 0},
		{"Whenever an opponent discards a card", TriggerPlayerSelectorOpponent, PlayerEventActionDiscard, PlayerEventCardSingle, PlayerEventOccurrenceAny, 0},
		{"Whenever a player cycles a card", TriggerPlayerSelectorAny, PlayerEventActionCycle, PlayerEventCardSingle, PlayerEventOccurrenceAny, 0},
		{"Whenever you discard one or more cards", TriggerPlayerSelectorYou, PlayerEventActionDiscard, PlayerEventCardOneOrMore, PlayerEventOccurrenceAny, 0},
		{"Whenever you cycle another card", TriggerPlayerSelectorYou, PlayerEventActionCycle, PlayerEventCardAnother, PlayerEventOccurrenceAny, 0},
		{"Whenever you cycle or discard another card", TriggerPlayerSelectorYou, PlayerEventActionCycleOrDiscard, PlayerEventCardAnother, PlayerEventOccurrenceAny, 0},
		{"Whenever you scry", TriggerPlayerSelectorYou, PlayerEventActionScry, PlayerEventCardNone, PlayerEventOccurrenceAny, 0},
		{"Whenever a player surveils", TriggerPlayerSelectorAny, PlayerEventActionSurveil, PlayerEventCardNone, PlayerEventOccurrenceAny, 0},
		{"Whenever an opponent gains life", TriggerPlayerSelectorOpponent, PlayerEventActionGainLife, PlayerEventCardNone, PlayerEventOccurrenceAny, 0},
		{"Whenever you lose life", TriggerPlayerSelectorYou, PlayerEventActionLoseLife, PlayerEventCardNone, PlayerEventOccurrenceAny, 0},
		{"Whenever a player draws their fourth card each turn", TriggerPlayerSelectorAny, PlayerEventActionDraw, PlayerEventCardSingle, PlayerEventOccurrenceOrdinalEachTurn, 4},
		{"When you surveil for the first time each turn", TriggerPlayerSelectorYou, PlayerEventActionSurveil, PlayerEventCardNone, PlayerEventOccurrenceFirstEachTurn, 1},
	}
	for _, test := range tests {
		t.Run(test.event, func(t *testing.T) {
			t.Parallel()
			source := test.event + ", draw a card."
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.PlayerEvent == nil {
				t.Fatalf("trigger = %#v, want typed player-event clause", trigger)
			}
			if trigger.PlayerEvent.Player.Kind != test.player ||
				trigger.PlayerEvent.Action.Kind != test.action ||
				trigger.PlayerEvent.Card.Kind != test.card ||
				trigger.PlayerEvent.Occurrence.Kind != test.occurrence ||
				trigger.PlayerEvent.Occurrence.Ordinal != test.ordinal {
				t.Fatalf("trigger = %#v", trigger)
			}
			assertTextSpan(t, "trigger clause", source, trigger.Span, trigger.Text)
			assertTextSpan(t, "trigger event", source, trigger.EventSpan, trigger.Event)
			assertSpanContains(t, "player-event clause", trigger.EventSpan, trigger.PlayerEvent.Span)
			assertSpanContains(t, "player selector", trigger.PlayerEvent.Span, trigger.PlayerEvent.Player.Span)
			assertSpanContains(t, "player action", trigger.PlayerEvent.Span, trigger.PlayerEvent.Action.Span)
		})
	}
}

func TestParsePlayerEventDiscardCardTypeFilters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		event    string
		card     PlayerEventCardKind
		required []TriggerCardType
		excluded []TriggerCardType
	}{
		{"you discard a creature card", PlayerEventCardSingle, []TriggerCardType{TriggerCardTypeCreature}, nil},
		{"you discard a land card", PlayerEventCardSingle, []TriggerCardType{TriggerCardTypeLand}, nil},
		{"you discard a nonland card", PlayerEventCardSingle, nil, []TriggerCardType{TriggerCardTypeLand}},
		{"you discard a noncreature, nonland card", PlayerEventCardSingle, nil, []TriggerCardType{TriggerCardTypeCreature, TriggerCardTypeLand}},
		{"you discard one or more artifact cards", PlayerEventCardOneOrMore, []TriggerCardType{TriggerCardTypeArtifact}, nil},
		{"an opponent discards a creature card", PlayerEventCardSingle, []TriggerCardType{TriggerCardTypeCreature}, nil},
	}
	for _, test := range tests {
		t.Run(test.event, func(t *testing.T) {
			t.Parallel()
			source := "Whenever " + test.event + ", draw a card."
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.PlayerEvent == nil {
				t.Fatalf("trigger = %#v, want typed player-event clause", trigger)
			}
			card := trigger.PlayerEvent.Card
			if card.Kind != test.card {
				t.Fatalf("card kind = %q, want %q", card.Kind, test.card)
			}
			if !slices.Equal(card.RequiredTypes, test.required) {
				t.Fatalf("required types = %#v, want %#v", card.RequiredTypes, test.required)
			}
			if !slices.Equal(card.ExcludedTypes, test.excluded) {
				t.Fatalf("excluded types = %#v, want %#v", card.ExcludedTypes, test.excluded)
			}
		})
	}
}

func TestParseEveryPreviouslySupportedSimplePlayerEventTriggerClause(t *testing.T) {
	t.Parallel()
	for _, event := range []string{
		"you draw a card",
		"an opponent draws a card",
		"a player draws a card",
		"you discard a card",
		"an opponent discards a card",
		"a player discards a card",
		"you discard one or more cards",
		"you cycle a card",
		"an opponent cycles a card",
		"a player cycles a card",
		"you cycle another card",
		"you scry",
		"an opponent scries",
		"a player scries",
		"you surveil",
		"an opponent surveils",
		"a player surveils",
		"you cycle or discard a card",
		"you cycle or discard another card",
		"you gain life",
		"an opponent gains life",
		"you lose life",
		"an opponent loses life",
	} {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse("Whenever "+event+", draw a card.", Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if trigger := document.Abilities[0].Trigger; trigger == nil || trigger.PlayerEvent == nil {
				t.Fatalf("trigger = %#v, want typed player-event clause", trigger)
			}
		})
	}
}

func TestParseEveryPreviouslySupportedPlayerEventOccurrenceTriggerClause(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever you draw your second card each turn, draw a card.",
		"Whenever an opponent draws their first card each turn, draw a card.",
		"Whenever a player draws their fifth card each turn, draw a card.",
		"Whenever you draw a card for the first time each turn, draw a card.",
		"Whenever an opponent draws a card for the first time each turn, draw a card.",
		"Whenever a player draws a card for the first time each turn, draw a card.",
		"Whenever you scry for the first time each turn, draw a card.",
		"Whenever an opponent scries for the first time each turn, draw a card.",
		"Whenever a player scries for the first time each turn, draw a card.",
		"Whenever you surveil for the first time each turn, draw a card.",
		"Whenever an opponent surveils for the first time each turn, draw a card.",
		"Whenever a player surveils for the first time each turn, draw a card.",
		"Whenever you gain life for the first time each turn, draw a card.",
		"Whenever an opponent gains life for the first time each turn, draw a card.",
		"When you lose life for the first time each turn, draw a card.",
		"When an opponent loses life for the first time each turn, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil ||
				trigger.PlayerEvent == nil ||
				trigger.PlayerEvent.Occurrence.Kind == PlayerEventOccurrenceAny {
				t.Fatalf("trigger = %#v, want typed player-event occurrence", trigger)
			}
		})
	}
}

func TestParsePlayerEventTriggerClausesComposePreviouslyUnsupportedSlots(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever an opponent discards one or more cards, draw a card.",
		"Whenever a player discards another card, draw a card.",
		"Whenever an opponent cycles or discards another card, draw a card.",
		"Whenever a player gains life, draw a card.",
		"Whenever a player loses life, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if trigger := document.Abilities[0].Trigger; trigger == nil || trigger.PlayerEvent == nil {
				t.Fatalf("trigger = %#v, want composed player-event grammar", trigger)
			}
		})
	}
}

func TestParsePlayerEventTriggerClausesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever your draw a card, draw a card.",
		"Whenever you draws a card, draw a card.",
		"Whenever a player draw a card, draw a card.",
		"When you draw a card, draw a card.",
		"Whenever you scry a card, draw a card.",
		"Whenever you draw one or more cards, draw a card.",
		"Whenever you cycle one or more cards, draw a card.",
		"Whenever you cycle or discard one or more cards, draw a card.",
		"Whenever you gain a life, draw a card.",
		"Whenever you discard another cards, draw a card.",
		"Whenever you discard a card for the second time each turn, draw a card.",
		"Whenever a player draws your second card each turn, draw a card.",
		"Whenever a player gains life for the first time each turn, draw a card.",
		"When a player loses life for the first time each turn, draw a card.",
		"At you draw a card, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.Event == "" || trigger.EventSpan == (shared.Span{}) {
				t.Fatalf("trigger = %#v, want source-spanned unrecognized clause", trigger)
			}
			if trigger.PlayerEvent != nil {
				t.Fatalf("trigger = %#v, want unrecognized player-event grammar", trigger)
			}
		})
	}
}
