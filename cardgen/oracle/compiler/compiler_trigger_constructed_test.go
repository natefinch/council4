package compiler

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileConstructedPhaseStepTriggerClauses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		quantifier parser.PhaseStepQuantifierKind
		player     parser.TriggerPlayerSelector
		phaseStep  parser.PhaseStepNameKind
		want       TriggerPattern
	}{
		{
			name:       "opponent postcombat main phase",
			quantifier: parser.PhaseStepQuantifierEach,
			player:     parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorOpponent},
			phaseStep:  parser.PhaseStepNamePostcombatMainPhase,
			want: TriggerPattern{
				Kind:       TriggerAt,
				Event:      TriggerEventBeginningOfStep,
				Controller: ControllerOpponent,
				Step:       TriggerStepPostcombatMain,
			},
		},
		{
			name:       "irregular first main phase",
			quantifier: parser.PhaseStepQuantifierEachOf,
			player:     parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorYou},
			phaseStep:  parser.PhaseStepNameFirstMainPhase,
			want: TriggerPattern{
				Kind:       TriggerAt,
				Event:      TriggerEventBeginningOfStep,
				Controller: ControllerYou,
				Step:       TriggerStepPrecombatMain,
			},
		},
		{
			name:       "attached selected permanent controller upkeep",
			quantifier: parser.PhaseStepQuantifierSingle,
			player: parser.TriggerPlayerSelector{
				Kind: parser.TriggerPlayerSelectorAttachedController,
				AttachedSubject: parser.TriggerAttachedSubject{
					Selection: parser.TriggerSelection{
						RequiredTypes: []parser.TriggerCardType{parser.TriggerCardTypeCreature},
						Supertypes:    []parser.TriggerSupertype{parser.TriggerSupertypeLegendary},
						ColorsAny:     []parser.TriggerColor{parser.TriggerColorWhite},
					},
				},
			},
			phaseStep: parser.PhaseStepNameUpkeep,
			want: TriggerPattern{
				Kind:  TriggerAt,
				Event: TriggerEventBeginningOfStep,
				Step:  TriggerStepUpkeep,
				StepPlayerSourceAttachedSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
					Supertypes:    []types.Super{types.Legendary},
					ColorsAny:     []color.Color{color.White},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			trigger := compileTrigger(&parser.Ability{
				Kind: parser.AbilityTriggered,
				Trigger: &parser.TriggerClause{
					Introduction: parser.TriggerIntroduction{Kind: parser.TriggerIntroductionAt},
					PhaseStep: &parser.PhaseStepTriggerClause{
						Quantifier: parser.PhaseStepQuantifier{Kind: test.quantifier},
						Player:     test.player,
						Name:       parser.PhaseStepName{Kind: test.phaseStep},
					},
				},
			}, Context{})
			if trigger.Event != "" {
				t.Fatalf("raw event = %q, want no Oracle wording", trigger.Event)
			}
			if !reflect.DeepEqual(trigger.Pattern, test.want) {
				t.Fatalf("pattern = %#v, want %#v", trigger.Pattern, test.want)
			}
		})
	}
}

func TestCompileComposedPhaseStepTriggerClauses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		controller ControllerKind
		step       TriggerStep
	}{
		{"At the beginning of combat on each player's turn, draw a card.", ControllerAny, TriggerStepBeginningOfCombat},
		{"At the beginning of each precombat main phase, draw a card.", ControllerAny, TriggerStepPrecombatMain},
		{"At the beginning of your end of combat step, draw a card.", ControllerYou, TriggerStepEndOfCombat},
		{"At the beginning of each of your upkeeps, draw a card.", ControllerYou, TriggerStepUpkeep},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			pattern := compilation.Abilities[0].Trigger.Pattern
			if pattern.Event != TriggerEventBeginningOfStep ||
				pattern.Controller != test.controller ||
				pattern.Step != test.step {
				t.Fatalf("pattern = %#v", pattern)
			}
		})
	}
}

func TestCompileConstructedPhaseStepTriggerClausesFailClosed(t *testing.T) {
	t.Parallel()
	for _, clause := range []parser.PhaseStepTriggerClause{
		{
			Quantifier: parser.PhaseStepQuantifier{Kind: parser.PhaseStepQuantifierUnknown},
			Player:     parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorYou},
			Name:       parser.PhaseStepName{Kind: parser.PhaseStepNameUpkeep},
		},
		{
			Quantifier: parser.PhaseStepQuantifier{Kind: parser.PhaseStepQuantifierSingle},
			Player:     parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorUnknown},
			Name:       parser.PhaseStepName{Kind: parser.PhaseStepNameUpkeep},
		},
		{
			Quantifier: parser.PhaseStepQuantifier{Kind: parser.PhaseStepQuantifierSingle},
			Player:     parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorYou},
			Name:       parser.PhaseStepName{Kind: parser.PhaseStepNameUnknown},
		},
	} {
		trigger := compileTrigger(&parser.Ability{
			Kind: parser.AbilityTriggered,
			Trigger: &parser.TriggerClause{
				Introduction: parser.TriggerIntroduction{Kind: parser.TriggerIntroductionAt},
				PhaseStep:    &clause,
			},
		}, Context{})
		if trigger.Pattern.Event != TriggerEventUnknown {
			t.Fatalf("pattern = %#v, want unknown event", trigger.Pattern)
		}
	}
}

func TestCompileConstructedPlayerEventTriggerClauses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		kind         parser.TriggerIntroductionKind
		player       parser.TriggerPlayerSelectorKind
		action       parser.PlayerEventActionKind
		card         parser.PlayerEventCardKind
		cardRequired []parser.TriggerCardType
		cardExcluded []parser.TriggerCardType
		occurrence   parser.PlayerEventOccurrence
		want         TriggerPattern
	}{
		{
			name:       "opponent batches discards",
			kind:       parser.TriggerIntroductionWhenever,
			player:     parser.TriggerPlayerSelectorOpponent,
			action:     parser.PlayerEventActionDiscard,
			card:       parser.PlayerEventCardOneOrMore,
			occurrence: parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceAny},
			want: TriggerPattern{
				Kind:      TriggerWhenever,
				Event:     TriggerEventCardDiscarded,
				Player:    TriggerPlayerOpponent,
				OneOrMore: true,
			},
		},
		{
			name:         "discards a creature card",
			kind:         parser.TriggerIntroductionWhenever,
			player:       parser.TriggerPlayerSelectorYou,
			action:       parser.PlayerEventActionDiscard,
			card:         parser.PlayerEventCardSingle,
			cardRequired: []parser.TriggerCardType{parser.TriggerCardTypeCreature},
			occurrence:   parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceAny},
			want: TriggerPattern{
				Kind:          TriggerWhenever,
				Event:         TriggerEventCardDiscarded,
				Player:        TriggerPlayerYou,
				CardSelection: TriggerSelection{RequiredTypes: []types.Card{types.Creature}},
			},
		},
		{
			name:         "discards a noncreature, nonland card",
			kind:         parser.TriggerIntroductionWhenever,
			player:       parser.TriggerPlayerSelectorYou,
			action:       parser.PlayerEventActionDiscard,
			card:         parser.PlayerEventCardSingle,
			cardExcluded: []parser.TriggerCardType{parser.TriggerCardTypeCreature, parser.TriggerCardTypeLand},
			occurrence:   parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceAny},
			want: TriggerPattern{
				Kind:          TriggerWhenever,
				Event:         TriggerEventCardDiscarded,
				Player:        TriggerPlayerYou,
				CardSelection: TriggerSelection{ExcludedTypes: []types.Card{types.Creature, types.Land}},
			},
		},
		{
			name:       "any player discards another card",
			kind:       parser.TriggerIntroductionWhenever,
			player:     parser.TriggerPlayerSelectorAny,
			action:     parser.PlayerEventActionDiscard,
			card:       parser.PlayerEventCardAnother,
			occurrence: parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceAny},
			want: TriggerPattern{
				Kind:        TriggerWhenever,
				Event:       TriggerEventCardDiscarded,
				Player:      TriggerPlayerAny,
				ExcludeSelf: true,
			},
		},
		{
			name:       "cycle or discard maps to discard",
			kind:       parser.TriggerIntroductionWhenever,
			player:     parser.TriggerPlayerSelectorYou,
			action:     parser.PlayerEventActionCycleOrDiscard,
			card:       parser.PlayerEventCardSingle,
			occurrence: parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceAny},
			want: TriggerPattern{
				Kind:   TriggerWhenever,
				Event:  TriggerEventCardDiscarded,
				Player: TriggerPlayerYou,
			},
		},
		{
			name:       "any player gains life",
			kind:       parser.TriggerIntroductionWhenever,
			player:     parser.TriggerPlayerSelectorAny,
			action:     parser.PlayerEventActionGainLife,
			card:       parser.PlayerEventCardNone,
			occurrence: parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceAny},
			want: TriggerPattern{
				Kind:   TriggerWhenever,
				Event:  TriggerEventLifeGained,
				Player: TriggerPlayerAny,
			},
		},
		{
			name:   "ordinal draw",
			kind:   parser.TriggerIntroductionWhenever,
			player: parser.TriggerPlayerSelectorYou,
			action: parser.PlayerEventActionDraw,
			card:   parser.PlayerEventCardSingle,
			occurrence: parser.PlayerEventOccurrence{
				Kind:    parser.PlayerEventOccurrenceOrdinalEachTurn,
				Ordinal: 4,
			},
			want: TriggerPattern{
				Kind:                       TriggerWhenever,
				Event:                      TriggerEventCardDrawn,
				Player:                     TriggerPlayerYou,
				PlayerEventOrdinalThisTurn: 4,
			},
		},
		{
			name:   "first surveil with when",
			kind:   parser.TriggerIntroductionWhen,
			player: parser.TriggerPlayerSelectorOpponent,
			action: parser.PlayerEventActionSurveil,
			card:   parser.PlayerEventCardNone,
			occurrence: parser.PlayerEventOccurrence{
				Kind:    parser.PlayerEventOccurrenceFirstEachTurn,
				Ordinal: 1,
			},
			want: TriggerPattern{
				Kind:                       TriggerWhen,
				Event:                      TriggerEventSurveil,
				Player:                     TriggerPlayerOpponent,
				PlayerEventOrdinalThisTurn: 1,
			},
		},
		{
			name:   "except first draw in draw step",
			kind:   parser.TriggerIntroductionWhenever,
			player: parser.TriggerPlayerSelectorOpponent,
			action: parser.PlayerEventActionDraw,
			card:   parser.PlayerEventCardSingle,
			occurrence: parser.PlayerEventOccurrence{
				Kind: parser.PlayerEventOccurrenceExceptFirstInDrawStep,
			},
			want: TriggerPattern{
				Kind:                       TriggerWhenever,
				Event:                      TriggerEventCardDrawn,
				Player:                     TriggerPlayerOpponent,
				ExcludeFirstDrawInDrawStep: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			trigger := compileTrigger(&parser.Ability{
				Kind: parser.AbilityTriggered,
				Trigger: &parser.TriggerClause{
					Introduction: parser.TriggerIntroduction{Kind: test.kind},
					Event:        "irrelevant source wording",
					PlayerEvent: &parser.PlayerEventTriggerClause{
						Player: parser.TriggerPlayerSelector{Kind: test.player},
						Action: parser.PlayerEventAction{Kind: test.action},
						Card: parser.PlayerEventCard{
							Kind:          test.card,
							RequiredTypes: test.cardRequired,
							ExcludedTypes: test.cardExcluded,
						},
						Occurrence: test.occurrence,
					},
				},
			}, Context{})
			if !reflect.DeepEqual(trigger.Pattern, test.want) {
				t.Fatalf("pattern = %#v, want %#v", trigger.Pattern, test.want)
			}
		})
	}
}

func TestCompileConstructedPlayerEventTriggerClausesFailClosed(t *testing.T) {
	t.Parallel()
	valid := parser.PlayerEventTriggerClause{
		Player:     parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorYou},
		Action:     parser.PlayerEventAction{Kind: parser.PlayerEventActionDraw},
		Card:       parser.PlayerEventCard{Kind: parser.PlayerEventCardSingle},
		Occurrence: parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceAny},
	}
	tests := []struct {
		name   string
		kind   parser.TriggerIntroductionKind
		clause parser.PlayerEventTriggerClause
	}{
		{name: "wrong introduction", kind: parser.TriggerIntroductionAt, clause: valid},
		{name: "simple when", kind: parser.TriggerIntroductionWhen, clause: valid},
		{name: "unknown player", kind: parser.TriggerIntroductionWhenever, clause: func() parser.PlayerEventTriggerClause {
			clause := valid
			clause.Player.Kind = parser.TriggerPlayerSelectorUnknown
			return clause
		}()},
		{name: "attached player", kind: parser.TriggerIntroductionWhenever, clause: func() parser.PlayerEventTriggerClause {
			clause := valid
			clause.Player.Kind = parser.TriggerPlayerSelectorAttachedController
			return clause
		}()},
		{name: "unknown action", kind: parser.TriggerIntroductionWhenever, clause: func() parser.PlayerEventTriggerClause {
			clause := valid
			clause.Action.Kind = parser.PlayerEventActionUnknown
			return clause
		}()},
		{name: "draw without card", kind: parser.TriggerIntroductionWhenever, clause: func() parser.PlayerEventTriggerClause {
			clause := valid
			clause.Card.Kind = parser.PlayerEventCardNone
			return clause
		}()},
		{name: "scry with card", kind: parser.TriggerIntroductionWhenever, clause: func() parser.PlayerEventTriggerClause {
			clause := valid
			clause.Action.Kind = parser.PlayerEventActionScry
			return clause
		}()},
		{name: "draw one or more", kind: parser.TriggerIntroductionWhenever, clause: func() parser.PlayerEventTriggerClause {
			clause := valid
			clause.Card.Kind = parser.PlayerEventCardOneOrMore
			return clause
		}()},
		{name: "discard first time", kind: parser.TriggerIntroductionWhenever, clause: func() parser.PlayerEventTriggerClause {
			clause := valid
			clause.Action.Kind = parser.PlayerEventActionDiscard
			clause.Occurrence = parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceFirstEachTurn, Ordinal: 1}
			return clause
		}()},
		{name: "any player life first time", kind: parser.TriggerIntroductionWhenever, clause: func() parser.PlayerEventTriggerClause {
			clause := valid
			clause.Player.Kind = parser.TriggerPlayerSelectorAny
			clause.Action.Kind = parser.PlayerEventActionGainLife
			clause.Card.Kind = parser.PlayerEventCardNone
			clause.Occurrence = parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceFirstEachTurn, Ordinal: 1}
			return clause
		}()},
		{name: "sixth draw", kind: parser.TriggerIntroductionWhenever, clause: func() parser.PlayerEventTriggerClause {
			clause := valid
			clause.Occurrence = parser.PlayerEventOccurrence{Kind: parser.PlayerEventOccurrenceOrdinalEachTurn, Ordinal: 6}
			return clause
		}()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			trigger := compileTrigger(&parser.Ability{
				Kind: parser.AbilityTriggered,
				Trigger: &parser.TriggerClause{
					Introduction: parser.TriggerIntroduction{Kind: test.kind},
					Event:        "text must not rescue invalid typed syntax",
					PlayerEvent:  &test.clause,
				},
			}, Context{})
			if trigger.Pattern.Event != TriggerEventUnknown {
				t.Fatalf("pattern = %#v, want unknown event", trigger.Pattern)
			}
		})
	}
}
