package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileOrdinalPlayerEventDuringThatPlayersTurn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		action     parser.PlayerEventActionKind
		wantEvent  TriggerEvent
		wantPlayer TriggerPlayerRelation
	}{
		{
			name:       "draw",
			action:     parser.PlayerEventActionDraw,
			wantEvent:  TriggerEventCardDrawn,
			wantPlayer: TriggerPlayerAny,
		},
		{
			name:      "cast",
			action:    parser.PlayerEventActionCast,
			wantEvent: TriggerEventSpellCast,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			trigger := compileTrigger(&parser.Ability{
				Kind: parser.AbilityTriggered,
				Trigger: &parser.TriggerClause{
					Introduction: parser.TriggerIntroduction{Kind: parser.TriggerIntroductionWhenever},
					PlayerEvent: &parser.PlayerEventTriggerClause{
						Player: parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorAny},
						Action: parser.PlayerEventAction{Kind: test.action},
						Card:   parser.PlayerEventCard{Kind: parser.PlayerEventCardSingle},
						Occurrence: parser.PlayerEventOccurrence{
							Kind:    parser.PlayerEventOccurrenceOrdinalEachTurn,
							Ordinal: 2,
						},
						TurnRelation: parser.TriggerCastTurnRelationEventPlayerTurn,
					},
				},
			}, Context{})
			pattern := trigger.Pattern
			if pattern.Event != test.wantEvent ||
				pattern.Player != test.wantPlayer ||
				pattern.PlayerEventOrdinalThisTurn != 2 ||
				pattern.CastDuringTurn != TriggerCastTurnEventPlayer {
				t.Fatalf("pattern = %#v", pattern)
			}
		})
	}
}
