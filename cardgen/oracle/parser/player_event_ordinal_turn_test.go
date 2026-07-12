package parser

import "testing"

func TestParseOrdinalPlayerEventDuringThatPlayersTurn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		event  string
		action PlayerEventActionKind
	}{
		{
			name:   "draw",
			event:  "Whenever a player draws their second card during their turn, you draw a card.",
			action: PlayerEventActionDraw,
		},
		{
			name:   "cast",
			event:  "Whenever a player casts their second spell during their turn, you create a 2/2 white Knight creature token.",
			action: PlayerEventActionCast,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.event, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.PlayerEvent == nil {
				t.Fatalf("trigger = %#v, want player event", trigger)
			}
			event := trigger.PlayerEvent
			if event.Player.Kind != TriggerPlayerSelectorAny ||
				event.Action.Kind != test.action ||
				event.Card.Kind != PlayerEventCardSingle ||
				event.Occurrence.Kind != PlayerEventOccurrenceOrdinalEachTurn ||
				event.Occurrence.Ordinal != 2 ||
				event.TurnRelation != TriggerCastTurnRelationEventPlayerTurn {
				t.Fatalf("player event = %#v", event)
			}
		})
	}
}
